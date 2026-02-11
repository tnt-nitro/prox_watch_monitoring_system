package watcher

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"
)

// RelayMode beschreibt das Verhalten des Relais.
// Phase 3: Klarstellung NO/NC Verdrahtung.
type RelayMode string

const (
	// RelayModeCutPowerOnActive: Relais trennt Strom, wenn aktiv (LOW bei Active LOW)
	// Typisch für NO (Normally Open) Relais: LOW = Relais schließt = Strom fließt
	// Für "Strom trennen" muss Relais inaktiv sein (HIGH bei Active LOW)
	RelayModeCutPowerOnActive RelayMode = "cut_power_on_active"

	// RelayModeCutPowerOnInactive: Relais trennt Strom, wenn inaktiv (HIGH bei Active LOW)
	// Typisch für NC (Normally Closed) Relais: HIGH = Relais öffnet = Strom getrennt
	RelayModeCutPowerOnInactive RelayMode = "cut_power_on_inactive"
)

// GPIORelayPowerCycler ist die GPIO-Implementierung des PowerCycler-Interfaces.
// Phase 3: Steuert ein Relais über GPIO für Power-Cycle.
type GPIORelayPowerCycler struct {
	powerCycler *powerCycler // Basis-Logik (ARM, Retry, State)
	pin         Pin           // GPIO-Pin für Relais
	activeLow   bool          // true = Active LOW, false = Active HIGH
	relayMode   RelayMode     // NO/NC Klarstellung
	minDowntime time.Duration // Mindest-Downtime
}

// NewGPIORelayPowerCycler erstellt einen neuen GPIO-Relais-PowerCycler.
// Phase 3: Kombiniert Safety Core (powerCycler) mit GPIO-Schaltung.
func NewGPIORelayPowerCycler(config PowerCycleConfig, store StateStore, pin Pin) (PowerCycler, error) {
	// Validierung: RelayMode muss gesetzt sein
	if config.RelayMode == "" {
		return nil, errors.New("powercycle: relay_mode must be set for GPIO power cycle")
	}

	// Validierung: RelayMode muss gültig sein
	if config.RelayMode != RelayModeCutPowerOnActive && config.RelayMode != RelayModeCutPowerOnInactive {
		return nil, fmt.Errorf("powercycle: invalid relay_mode %q, must be %q or %q",
			config.RelayMode, RelayModeCutPowerOnActive, RelayModeCutPowerOnInactive)
	}

	// Basis-PowerCycler erstellen (Safety Core)
	baseCycler, err := NewPowerCycler(config, store)
	if err != nil {
		return nil, fmt.Errorf("failed to create base powercycler: %w", err)
	}

	// Cast zu *powerCycler (internes Interface)
	baseImpl, ok := baseCycler.(*powerCycler)
	if !ok {
		return nil, errors.New("powercycle: internal error: base cycler is not *powerCycler")
	}

	minDowntime := time.Duration(config.MinDowntimeSeconds) * time.Second

	return &GPIORelayPowerCycler{
		powerCycler: baseImpl,
		pin:         pin,
		activeLow:   !config.RelayActiveHigh, // Config: RelayActiveHigh, intern: activeLow
		relayMode:   config.RelayMode,
		minDowntime: minDowntime,
	}, nil
}

// Attempt führt einen Power-Cycle-Versuch durch.
// Phase 3: GPIO-Schaltung mit Safety Core Prüfungen.
func (g *GPIORelayPowerCycler) Attempt(ctx context.Context) error {
	// 1. Safety Core Prüfungen (ARM, Retry, Max Attempts)
	// Diese werden in powerCycler.Attempt() durchgeführt, aber wir müssen sie hier auch prüfen,
	// da wir die GPIO-Schaltung durchführen wollen.
	// Wir rufen die Basis-Logik auf, die alle Prüfungen macht und State speichert.
	// Aber wir müssen die GPIO-Schaltung selbst durchführen.

	// Prüfe ARM-Datei (wenn erforderlich)
	if g.powerCycler.config.RequireManualArm {
		if !g.isArmed() {
			return errors.New("powercycle: not armed (arm file missing)")
		}
	}

	// Lade State für Prüfungen
	if g.powerCycler.store == nil {
		return errors.New("powercycle: state store not available")
	}

	state, err := g.powerCycler.store.Load()
	if err != nil {
		return fmt.Errorf("powercycle: failed to load state: %w", err)
	}

	// Prüfe Max Attempts
	if state.PowerAttempts >= g.powerCycler.config.MaxAttempts {
		return fmt.Errorf("powercycle: max attempts (%d) reached", g.powerCycler.config.MaxAttempts)
	}

	// Prüfe Retry-Cooldown
	now := time.Now()
	if !state.LastPowerAttempt.IsZero() {
		retryDuration := time.Duration(g.powerCycler.config.RetryAfterSeconds) * time.Second
		if now.Sub(state.LastPowerAttempt) < retryDuration {
			return fmt.Errorf("powercycle: retry cooldown active (wait %d seconds)", g.powerCycler.config.RetryAfterSeconds)
		}
	}

	// 2. GPIO-Schaltung durchführen
	if err := g.performPowerCycle(ctx); err != nil {
		// Fehler bei GPIO-Schaltung → ARM-Datei bleibt (für Diagnose)
		return fmt.Errorf("powercycle: GPIO operation failed: %w", err)
	}

	// 3. State aktualisieren (nach erfolgreicher GPIO-Schaltung)
	state.PowerAttempts++
	state.LastPowerAttempt = now

	if err := g.powerCycler.store.Save(state); err != nil {
		return fmt.Errorf("powercycle: failed to save state: %w", err)
	}

	// 4. ARM-Datei entfernen (nach erfolgreichem Versuch)
	if g.powerCycler.config.RequireManualArm {
		if err := g.removeArmFile(); err != nil {
			// Fehler beim Entfernen der ARM-Datei ist nicht kritisch
			// (State wurde bereits gespeichert)
		}
	}

	return nil
}

// performPowerCycle führt die physische GPIO-Schaltung durch.
// Phase 3: Power OFF → warten → Power ON.
//
// Logik:
// - Active LOW: LOW = Relais aktiv (zieht an), HIGH = Relais inaktiv (fällt ab)
// - RelayModeCutPowerOnActive: Relais trennt Strom, wenn aktiv (LOW)
//   → Power OFF: LOW setzen (Relais aktiv = Strom getrennt)
//   → Power ON: HIGH setzen (Relais inaktiv = Strom fließt)
// - RelayModeCutPowerOnInactive: Relais trennt Strom, wenn inaktiv (HIGH)
//   → Power OFF: HIGH setzen (Relais inaktiv = Strom getrennt)
//   → Power ON: LOW setzen (Relais aktiv = Strom fließt)
func (g *GPIORelayPowerCycler) performPowerCycle(ctx context.Context) error {
	// Power OFF: Relais trennt Strom
	var powerOffErr error
	if g.relayMode == RelayModeCutPowerOnActive {
		// Relais trennt Strom, wenn aktiv (LOW bei Active LOW)
		// Für "Strom trennen" → LOW setzen
		if g.activeLow {
			powerOffErr = g.pin.Low() // Active LOW: LOW = Relais aktiv = Strom getrennt
		} else {
			powerOffErr = g.pin.High() // Active HIGH: HIGH = Relais aktiv = Strom getrennt
		}
	} else {
		// RelayModeCutPowerOnInactive: Relais trennt Strom, wenn inaktiv (HIGH bei Active LOW)
		// Für "Strom trennen" → HIGH setzen
		if g.activeLow {
			powerOffErr = g.pin.High() // Active LOW: HIGH = Relais inaktiv = Strom getrennt
		} else {
			powerOffErr = g.pin.Low() // Active HIGH: LOW = Relais inaktiv = Strom getrennt
		}
	}
	if powerOffErr != nil {
		return fmt.Errorf("failed to set pin (power OFF): %w", powerOffErr)
	}

	// Warten min_downtime_seconds
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(g.minDowntime):
		// Downtime abgelaufen
	}

	// Power ON: Relais schließt wieder (Strom fließt)
	var powerOnErr error
	if g.relayMode == RelayModeCutPowerOnActive {
		// Relais trennt Strom, wenn aktiv (LOW bei Active LOW)
		// Für "Strom fließt" → HIGH setzen (Relais inaktiv)
		if g.activeLow {
			powerOnErr = g.pin.High() // Active LOW: HIGH = Relais inaktiv = Strom fließt
		} else {
			powerOnErr = g.pin.Low() // Active HIGH: LOW = Relais inaktiv = Strom fließt
		}
	} else {
		// RelayModeCutPowerOnInactive: Relais trennt Strom, wenn inaktiv (HIGH bei Active LOW)
		// Für "Strom fließt" → LOW setzen (Relais aktiv)
		if g.activeLow {
			powerOnErr = g.pin.Low() // Active LOW: LOW = Relais aktiv = Strom fließt
		} else {
			powerOnErr = g.pin.High() // Active HIGH: HIGH = Relais aktiv = Strom fließt
		}
	}
	if powerOnErr != nil {
		return fmt.Errorf("failed to set pin (power ON): %w", powerOnErr)
	}

	// WICHTIG: Pin bleibt im "Power ON" Zustand
	// Kein automatisches Zurücksetzen (sonst könnte Relais im falschen Zustand bleiben)

	return nil
}

// isArmed prüft, ob die ARM-Datei existiert.
func (g *GPIORelayPowerCycler) isArmed() bool {
	if g.powerCycler.config.ArmFilePath == "" {
		return false
	}
	_, err := os.Stat(g.powerCycler.config.ArmFilePath)
	return err == nil
}

// removeArmFile entfernt die ARM-Datei.
func (g *GPIORelayPowerCycler) removeArmFile() error {
	if g.powerCycler.config.ArmFilePath == "" {
		return nil
	}
	if err := os.Remove(g.powerCycler.config.ArmFilePath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to remove arm file: %w", err)
	}
	return nil
}
