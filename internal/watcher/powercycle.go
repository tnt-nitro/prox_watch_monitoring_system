package watcher

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// PowerCycler ist das Interface für Power-Cycle-Operationen.
// Phase 3: Safety Core - Logik ohne GPIO-Schaltung.
type PowerCycler interface {
	// Attempt führt einen Power-Cycle-Versuch durch.
	// Phase 3: Prüft alle Sicherheitsbedingungen vor dem Versuch.
	Attempt(ctx context.Context) error
}

// PowerCycleConfig enthält die Konfiguration für Power-Cycle.
// Phase 3: Safety Core + GPIO-Schaltung.
type PowerCycleConfig struct {
	Enabled            bool
	GPIOPin            int
	RelayActiveHigh    bool
	RelayMode          RelayMode // Phase 3: NO/NC Klarstellung ("cut_power_on_active" oder "cut_power_on_inactive")
	MaxAttempts        int
	MinDowntimeSeconds int
	RetryAfterSeconds  int
	RequireManualArm   bool
	ArmFilePath        string // Pfad zur ARM-Datei (z.B. /var/lib/prox-watch-watcher/arm_powercycle)
}

// DefaultPowerCycleConfig gibt eine Standard-Power-Cycle-Konfiguration zurück.
func DefaultPowerCycleConfig() PowerCycleConfig {
	return PowerCycleConfig{
		Enabled:            false,                    // Default: deaktiviert
		GPIOPin:            24,                      // BCM Pin 24
		RelayActiveHigh:    false,                   // LOW-aktiv (typisch für Optokoppler)
		RelayMode:          "",                      // Phase 3: Muss gesetzt werden (kein Default)
		MaxAttempts:        1,                       // Max. 1 Versuch
		MinDowntimeSeconds: 15,                      // Mindestens 15 Sekunden Downtime
		RetryAfterSeconds:  900,                     // 15 Minuten Retry-Cooldown
		RequireManualArm:   true,                     // ARM-Datei erforderlich
		ArmFilePath:        "/var/lib/prox-watch-watcher/arm_powercycle", // Muss innerhalb ReadWritePaths liegen
	}
}

// Validate prüft die Power-Cycle-Konfiguration auf Konsistenz.
func (c PowerCycleConfig) Validate() error {
	if !c.Enabled {
		return nil // Deaktiviert → immer gültig
	}

	if c.GPIOPin <= 0 {
		return errors.New("powercycle: gpio_pin must be > 0")
	}

	if c.MaxAttempts < 1 {
		return errors.New("powercycle: max_attempts must be >= 1")
	}

	if c.MinDowntimeSeconds < 1 {
		return errors.New("powercycle: min_downtime_seconds must be >= 1")
	}

	if c.RetryAfterSeconds < 1 {
		return errors.New("powercycle: retry_after_seconds must be >= 1")
	}

	if c.RequireManualArm && c.ArmFilePath == "" {
		return errors.New("powercycle: arm_file_path must be set if require_manual_arm is true")
	}

	// Phase 3: RelayMode muss gesetzt sein (wenn Enabled)
	if c.Enabled && c.RelayMode == "" {
		return errors.New("powercycle: relay_mode must be set when enabled")
	}

	// Phase 3: RelayMode muss gültig sein
	if c.RelayMode != "" && c.RelayMode != RelayModeCutPowerOnActive && c.RelayMode != RelayModeCutPowerOnInactive {
		return fmt.Errorf("powercycle: invalid relay_mode %q, must be %q or %q",
			c.RelayMode, RelayModeCutPowerOnActive, RelayModeCutPowerOnInactive)
	}

	return nil
}

// powerCycler ist die Implementierung des PowerCycler-Interfaces.
// Phase 3: Safety Core - Logik ohne GPIO-Schaltung.
type powerCycler struct {
	config PowerCycleConfig
	store  StateStore
}

// NewPowerCycler erstellt einen neuen PowerCycler.
// Phase 3: Safety Core - noch ohne GPIO-Schaltung.
func NewPowerCycler(config PowerCycleConfig, store StateStore) (PowerCycler, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid powercycle config: %w", err)
	}

	return &powerCycler{
		config: config,
		store:  store,
	}, nil
}

// Attempt führt einen Power-Cycle-Versuch durch.
// Phase 3: Prüft alle Sicherheitsbedingungen vor dem Versuch.
// NOCH KEINE GPIO-SCHALTUNG - nur Logik & Guards.
func (p *powerCycler) Attempt(ctx context.Context) error {
	// 1. Prüfe, ob Power-Cycle aktiviert ist
	if !p.config.Enabled {
		return errors.New("powercycle: not enabled")
	}

	// 2. Prüfe ARM-Datei (wenn erforderlich)
	if p.config.RequireManualArm {
		if !p.isArmed() {
			return errors.New("powercycle: not armed (arm file missing)")
		}
	}

	// 3. Lade State aus Persistenz
	if p.store == nil {
		return errors.New("powercycle: state store not available")
	}

	state, err := p.store.Load()
	if err != nil {
		// Persistenzfehler → kein Versuch (Sicherheit)
		return fmt.Errorf("powercycle: failed to load state: %w", err)
	}

	// 4. Prüfe Max Attempts
	if state.PowerAttempts >= p.config.MaxAttempts {
		return fmt.Errorf("powercycle: max attempts (%d) reached", p.config.MaxAttempts)
	}

	// 5. Prüfe Retry-Cooldown
	now := time.Now()
	if !state.LastPowerAttempt.IsZero() {
		retryDuration := time.Duration(p.config.RetryAfterSeconds) * time.Second
		if now.Sub(state.LastPowerAttempt) < retryDuration {
			return fmt.Errorf("powercycle: retry cooldown active (wait %d seconds)", p.config.RetryAfterSeconds)
		}
	}

	// 6. Alle Bedingungen erfüllt → Versuch durchführen
	// Phase 3 Schritt 1: NOCH KEINE GPIO-SCHALTUNG
	// Nur State aktualisieren und ARM-Datei entfernen

	// State aktualisieren
	state.PowerAttempts++
	state.LastPowerAttempt = now

	// State speichern
	if err := p.store.Save(state); err != nil {
		return fmt.Errorf("powercycle: failed to save state: %w", err)
	}

	// ARM-Datei entfernen (nach erfolgreichem Versuch)
	if p.config.RequireManualArm {
		if err := p.removeArmFile(); err != nil {
			// Fehler beim Entfernen der ARM-Datei ist nicht kritisch
			// (State wurde bereits gespeichert)
			// Aber wir loggen es nicht (kein Log mit sensiblen Daten)
		}
	}

	// Phase 3 Schritt 1: NOCH KEINE GPIO-SCHALTUNG
	// TODO: GPIO-Schaltung in Schritt 2

	return nil
}

// isArmed prüft, ob die ARM-Datei existiert.
func (p *powerCycler) isArmed() bool {
	if p.config.ArmFilePath == "" {
		return false
	}

	// Prüfe, ob Datei existiert
	_, err := os.Stat(p.config.ArmFilePath)
	return err == nil
}

// removeArmFile entfernt die ARM-Datei.
func (p *powerCycler) removeArmFile() error {
	if p.config.ArmFilePath == "" {
		return nil
	}

	// Entferne Datei
	if err := os.Remove(p.config.ArmFilePath); err != nil {
		if os.IsNotExist(err) {
			// Datei existiert nicht → OK (bereits entfernt)
			return nil
		}
		return fmt.Errorf("failed to remove arm file: %w", err)
	}

	return nil
}

// Allowed prüft, ob ein Power-Cycle-Versuch erlaubt ist.
// Phase 3: Helper-Funktion für Runner.
func AllowedPowerCycle(severity rules.Severity, config PowerCycleConfig, state PersistedState) bool {
	// 1. Nur bei CRIT
	if severity != rules.SeverityCrit {
		return false
	}

	// 2. Power-Cycle muss aktiviert sein
	if !config.Enabled {
		return false
	}

	// 3. ARM-Datei muss existieren (wenn erforderlich)
	if config.RequireManualArm {
		if _, err := os.Stat(config.ArmFilePath); err != nil {
			return false
		}
	}

	// 4. Max Attempts nicht überschritten
	if state.PowerAttempts >= config.MaxAttempts {
		return false
	}

	// 5. Retry-Cooldown abgelaufen
	now := time.Now()
	if !state.LastPowerAttempt.IsZero() {
		retryDuration := time.Duration(config.RetryAfterSeconds) * time.Second
		if now.Sub(state.LastPowerAttempt) < retryDuration {
			return false
		}
	}

	return true
}
