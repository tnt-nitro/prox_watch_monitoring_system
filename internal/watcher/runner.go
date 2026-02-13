package watcher

import (
	"context"
	"fmt"
	"prox-watch/internal/rules"
	"time"
)

// Runner orchestriert den Watcher-Prozess.
// Phase 2: Mit Persistenz und Cooldown.
// Single-Thread, 30s Intervall, Health-Check → Counter → Severity → Push → GPIO
// Siehe docs/19_watcher_runner.md für vollständige Spezifikation.
type Runner interface {
	Run(ctx context.Context) error
	Stop() error
	Wait() error
}

// runner ist die Implementierung des Runner-Interfaces.
// Phase 3: Erweitert um PowerCycler.
type runner struct {
	health        HealthChecker
	counter       Counter
	push          *PushService
	gpio          GPIO
	powerCycler   PowerCycler // Phase 3: Power-Cycle-Interface
	state         *State
	store         StateStore
	interval      time.Duration
	warnThreshold int
	critThreshold int
	cooldownSecs  int
	powerCycleCfg PowerCycleConfig // Phase 3: Power-Cycle-Konfiguration

	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{}
}

// RunnerConfig enthält die Konfiguration für den Runner.
// Phase 3: Erweitert um PowerCycler und PowerCycleConfig.
type RunnerConfig struct {
	Health        HealthChecker
	Counter       Counter
	Push          *PushService
	GPIO          GPIO
	PowerCycler   PowerCycler      // Phase 3: Power-Cycle-Interface (optional)
	Store         StateStore
	Interval      time.Duration
	WarnThreshold int
	CritThreshold int
	CooldownSecs  int              // Cooldown in Sekunden (Default: 600 = 10 Minuten)
	PowerCycleCfg PowerCycleConfig // Phase 3: Power-Cycle-Konfiguration
}

// NewRunner erstellt einen neuen Runner.
// Phase 2: Lädt persistenten State beim Start.
func NewRunner(config RunnerConfig) (Runner, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// State mit Thresholds erstellen
	state := NewStateWithThresholds(config.WarnThreshold, config.CritThreshold)

	// Cooldown-Default: 600 Sekunden (10 Minuten)
	cooldownSecs := config.CooldownSecs
	if cooldownSecs <= 0 {
		cooldownSecs = 600
	}

	runner := &runner{
		health:        config.Health,
		counter:       config.Counter,
		push:          config.Push,
		gpio:          config.GPIO,
		powerCycler:   config.PowerCycler, // Phase 3: Power-Cycle-Interface
		state:         state,
		store:         config.Store,
		interval:      config.Interval,
		warnThreshold: config.WarnThreshold,
		critThreshold: config.CritThreshold,
		cooldownSecs:  cooldownSecs,
		powerCycleCfg: config.PowerCycleCfg, // Phase 3: Power-Cycle-Konfiguration
		ctx:           ctx,
		cancel:        cancel,
		done:          make(chan struct{}),
	}

	// Phase 2: Lade persistenten State beim Start
	if config.Store != nil {
		persisted, err := config.Store.Load()
		if err != nil {
			// Fehler beim Laden: Weiter mit Default-State (kein Panic)
			// Logging optional, aber nicht blockierend
		} else {
			// State aus Persistenz wiederherstellen
			state.FailCount = persisted.FailCount
			state.CurrentSeverity = persisted.CurrentSeverity
			// LastEscalation wird in state.LastEscalation gespeichert (später)
		}
	}

	return runner, nil
}

// Run startet den Event-Loop.
// Phase 2: Mit Persistenz und Cooldown.
func (r *runner) Run(ctx context.Context) error {
	defer close(r.done)

	// Erstelle Ticker für Intervall
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// Context-Cancellation (extern)
			return ctx.Err()

		case <-r.ctx.Done():
			// Interne Cancellation (Stop())
			return nil

		case <-ticker.C:
			// Health-Check ausführen
			result, err := r.health.Check(ctx)
			if err != nil {
				// Health-Check-Fehler: Loggen und weiter mit nächstem Intervall
				// Kein Stopp, kein Retry
				// Verwende fmt.Printf für direkte Ausgabe (wird von systemd journal erfasst)
				fmt.Printf("ERROR: Health check failed: %v\n", err)
				continue
			}

			// Counter & Severity
			var newSeverity rules.Severity
			var failCountChanged, severityChanged bool

			if result.Success {
				// Erfolg: Counter zurücksetzen
				oldFailCount := r.state.FailCount
				oldSeverity := r.state.CurrentSeverity
				r.counter.Reset()
				r.state.FailCount = 0
				newSeverity = rules.SeverityInfo
				failCountChanged = (oldFailCount != 0)
				
				// Logge Erfolg nur bei Statuswechsel (von Fehler zu Erfolg)
				if oldSeverity != rules.SeverityInfo {
					fmt.Printf("✓ Health check OK - Status recovered to INFO (failures: 0)\n")
				}
			} else {
				// Fehler: Counter erhöhen
				r.counter.Increment()
				failCount := r.counter.GetCount()
				r.state.FailCount = failCount
				newSeverity = EvaluateSeverity(failCount, r.warnThreshold, r.critThreshold)
				failCountChanged = true
				
				// Logge Fehler immer
				fmt.Printf("ERROR: Health check failed - Failures: %d, Severity: %s\n", failCount, newSeverity.String())
			}

			severityChanged = (newSeverity != r.state.CurrentSeverity)

			// Push nur bei Eskalation (newSeverity > currentSeverity) UND Cooldown abgelaufen
			shouldPush := ShouldPush(r.state.CurrentSeverity, newSeverity)
			now := time.Now()

			// Cooldown-Prüfung: Lade LastEscalation aus Store (wenn vorhanden)
			var lastEscalation time.Time
			if r.store != nil {
				persisted, err := r.store.Load()
				if err == nil {
					lastEscalation = persisted.LastEscalation
				}
			}

			cooldownExpired := true
			if shouldPush && !lastEscalation.IsZero() {
				cooldownDuration := time.Duration(r.cooldownSecs) * time.Second
				cooldownExpired = now.Sub(lastEscalation) >= cooldownDuration
			}

			if shouldPush && cooldownExpired {
				// Push senden (nicht blockierend, Fehler werden ignoriert)
				_ = r.push.SendIfEscalation(ctx, r.state.CurrentSeverity, newSeverity)

				// LastEscalation aktualisieren
				lastEscalation = now
			}

			// GPIO Update (nur bei Severity-Änderung)
			if severityChanged {
				// LED setzen
				_ = r.gpio.SetLED(newSeverity)

				// Beep nur bei Eskalation zu CRIT
				if newSeverity == rules.SeverityCrit && shouldPush && cooldownExpired {
					_ = r.gpio.Beep()
				}
			}

			// State aktualisieren
			previousSeverity := r.state.CurrentSeverity
			r.state.CurrentSeverity = newSeverity

			// Phase 3: Power-Cycle prüfen (nur bei CRIT-Edge-Trigger)
			// Edge-Trigger: Nur wenn Severity von < CRIT → CRIT wechselt
			if newSeverity == rules.SeverityCrit &&
				previousSeverity < rules.SeverityCrit &&
				r.powerCycler != nil &&
				r.store != nil {
				// Lade State für Power-Cycle-Prüfung
				persisted, err := r.store.Load()
				if err == nil {
					// Prüfe, ob Power-Cycle erlaubt ist (ARM, Max Attempts, Retry-Cooldown)
					if AllowedPowerCycle(newSeverity, r.powerCycleCfg, persisted) {
						// Power-Cycle-Versuch (Fehler werden ignoriert, kein Panic)
						_ = r.powerCycler.Attempt(ctx)
						// State wird in Attempt() gespeichert
					}
				}
			}

			// Phase 2: Save nur bei Änderungen (FailCount, Severity, LastEscalation)
			// Phase 3: PowerAttempts werden in Attempt() gespeichert, nicht hier
			if r.store != nil && (failCountChanged || severityChanged || shouldPush) {
				// Lade aktuellen State (inkl. PowerAttempts)
				persisted, err := r.store.Load()
				if err == nil {
					persisted.FailCount = r.state.FailCount
					persisted.CurrentSeverity = r.state.CurrentSeverity
					persisted.LastEscalation = lastEscalation
					// PowerAttempts und LastPowerAttempt bleiben unverändert (werden nur in Attempt() gesetzt)
					// Save-Fehler werden ignoriert (kein Panic, kein Blocking)
					_ = r.store.Save(persisted)
				}
			}
		}
	}
}

// Stop stoppt den Runner (Graceful Shutdown).
func (r *runner) Stop() error {
	// Context-Cancellation
	r.cancel()
	return nil
}

// Wait wartet auf das Ende des Runners.
func (r *runner) Wait() error {
	<-r.done
	return nil
}
