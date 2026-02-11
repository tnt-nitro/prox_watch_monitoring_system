package watcher

import (
	"context"
	"os"
	"path/filepath"
	"prox-watch/internal/rules"
	"testing"
	"time"
)

func TestDefaultPowerCycleConfig(t *testing.T) {
	cfg := DefaultPowerCycleConfig()

	if cfg.Enabled != false {
		t.Errorf("Expected Enabled=false, got %v", cfg.Enabled)
	}
	if cfg.GPIOPin != 24 {
		t.Errorf("Expected GPIOPin=24, got %d", cfg.GPIOPin)
	}
	if cfg.RelayActiveHigh != false {
		t.Errorf("Expected RelayActiveHigh=false, got %v", cfg.RelayActiveHigh)
	}
	if cfg.MaxAttempts != 1 {
		t.Errorf("Expected MaxAttempts=1, got %d", cfg.MaxAttempts)
	}
	if cfg.MinDowntimeSeconds != 15 {
		t.Errorf("Expected MinDowntimeSeconds=15, got %d", cfg.MinDowntimeSeconds)
	}
	if cfg.RetryAfterSeconds != 900 {
		t.Errorf("Expected RetryAfterSeconds=900, got %d", cfg.RetryAfterSeconds)
	}
	if cfg.RequireManualArm != true {
		t.Errorf("Expected RequireManualArm=true, got %v", cfg.RequireManualArm)
	}
}

func TestPowerCycleConfig_Validate(t *testing.T) {
	// Test: Deaktiviert → immer gültig
	cfg := DefaultPowerCycleConfig()
	cfg.Enabled = false
	if err := cfg.Validate(); err != nil {
		t.Errorf("Expected no error for disabled config, got %v", err)
	}

	// Test: Aktiviert, gültig
	cfg.Enabled = true
	if err := cfg.Validate(); err != nil {
		t.Errorf("Expected no error for valid config, got %v", err)
	}

	// Test: GPIOPin <= 0
	cfg.GPIOPin = 0
	if err := cfg.Validate(); err == nil {
		t.Error("Expected error for GPIOPin <= 0, got nil")
	}
	cfg.GPIOPin = 24 // Reset

	// Test: MaxAttempts < 1
	cfg.MaxAttempts = 0
	if err := cfg.Validate(); err == nil {
		t.Error("Expected error for MaxAttempts < 1, got nil")
	}
	cfg.MaxAttempts = 1 // Reset

	// Test: MinDowntimeSeconds < 1
	cfg.MinDowntimeSeconds = 0
	if err := cfg.Validate(); err == nil {
		t.Error("Expected error for MinDowntimeSeconds < 1, got nil")
	}
	cfg.MinDowntimeSeconds = 15 // Reset

	// Test: RetryAfterSeconds < 1
	cfg.RetryAfterSeconds = 0
	if err := cfg.Validate(); err == nil {
		t.Error("Expected error for RetryAfterSeconds < 1, got nil")
	}
	cfg.RetryAfterSeconds = 900 // Reset

	// Test: RequireManualArm ohne ArmFilePath
	cfg.RequireManualArm = true
	cfg.ArmFilePath = ""
	if err := cfg.Validate(); err == nil {
		t.Error("Expected error for RequireManualArm without ArmFilePath, got nil")
	}
}

func TestPowerCycler_Attempt_NotEnabled(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "watcher_state.db")

	store, err := NewSQLiteStateStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	cfg := DefaultPowerCycleConfig()
	cfg.Enabled = false

	cycler, err := NewPowerCycler(cfg, store)
	if err != nil {
		t.Fatalf("Failed to create powercycler: %v", err)
	}

	err = cycler.Attempt(context.Background())
	if err == nil {
		t.Error("Expected error when not enabled, got nil")
	}
}

func TestPowerCycler_Attempt_NotArmed(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "watcher_state.db")
	armPath := filepath.Join(tmpDir, "arm_powercycle")

	store, err := NewSQLiteStateStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	cfg := DefaultPowerCycleConfig()
	cfg.Enabled = true
	cfg.ArmFilePath = armPath
	cfg.RequireManualArm = true

	cycler, err := NewPowerCycler(cfg, store)
	if err != nil {
		t.Fatalf("Failed to create powercycler: %v", err)
	}

	// ARM-Datei existiert nicht
	err = cycler.Attempt(context.Background())
	if err == nil {
		t.Error("Expected error when not armed, got nil")
	}
}

func TestPowerCycler_Attempt_Armed(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "watcher_state.db")
	armPath := filepath.Join(tmpDir, "arm_powercycle")

	store, err := NewSQLiteStateStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Erstelle ARM-Datei
	if err := os.WriteFile(armPath, []byte(""), 0600); err != nil {
		t.Fatalf("Failed to create arm file: %v", err)
	}

	cfg := DefaultPowerCycleConfig()
	cfg.Enabled = true
	cfg.ArmFilePath = armPath
	cfg.RequireManualArm = true

	cycler, err := NewPowerCycler(cfg, store)
	if err != nil {
		t.Fatalf("Failed to create powercycler: %v", err)
	}

	// Attempt sollte erfolgreich sein
	err = cycler.Attempt(context.Background())
	if err != nil {
		t.Errorf("Expected no error when armed, got %v", err)
	}

	// Prüfe, dass State aktualisiert wurde
	state, err := store.Load()
	if err != nil {
		t.Fatalf("Failed to load state: %v", err)
	}
	if state.PowerAttempts != 1 {
		t.Errorf("Expected PowerAttempts=1, got %d", state.PowerAttempts)
	}
	if state.LastPowerAttempt.IsZero() {
		t.Error("Expected LastPowerAttempt to be set, got zero")
	}

	// Prüfe, dass ARM-Datei entfernt wurde
	if _, err := os.Stat(armPath); err == nil {
		t.Error("Expected arm file to be removed after attempt, but it still exists")
	}
}

func TestPowerCycler_Attempt_MaxAttemptsReached(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "watcher_state.db")
	armPath := filepath.Join(tmpDir, "arm_powercycle")

	store, err := NewSQLiteStateStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Setze State auf Max Attempts
	state := PersistedState{
		FailCount:        10,
		CurrentSeverity:  rules.SeverityCrit,
		LastEscalation:   time.Now(),
		PowerAttempts:    1, // Max erreicht
		LastPowerAttempt: time.Now(),
	}
	if err := store.Save(state); err != nil {
		t.Fatalf("Failed to save state: %v", err)
	}

	// Erstelle ARM-Datei
	if err := os.WriteFile(armPath, []byte(""), 0600); err != nil {
		t.Fatalf("Failed to create arm file: %v", err)
	}

	cfg := DefaultPowerCycleConfig()
	cfg.Enabled = true
	cfg.ArmFilePath = armPath
	cfg.RequireManualArm = true
	cfg.MaxAttempts = 1

	cycler, err := NewPowerCycler(cfg, store)
	if err != nil {
		t.Fatalf("Failed to create powercycler: %v", err)
	}

	// Attempt sollte blockiert sein
	err = cycler.Attempt(context.Background())
	if err == nil {
		t.Error("Expected error when max attempts reached, got nil")
	}
}

func TestPowerCycler_Attempt_RetryCooldownActive(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "watcher_state.db")
	armPath := filepath.Join(tmpDir, "arm_powercycle")

	store, err := NewSQLiteStateStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Setze State mit kürzlichem Power Attempt (vor 5 Sekunden, Cooldown=900 Sekunden)
	state := PersistedState{
		FailCount:        10,
		CurrentSeverity:  rules.SeverityCrit,
		LastEscalation:   time.Now(),
		PowerAttempts:    0,
		LastPowerAttempt: time.Now().Add(-5 * time.Second), // Vor 5 Sekunden
	}
	if err := store.Save(state); err != nil {
		t.Fatalf("Failed to save state: %v", err)
	}

	// Erstelle ARM-Datei
	if err := os.WriteFile(armPath, []byte(""), 0600); err != nil {
		t.Fatalf("Failed to create arm file: %v", err)
	}

	cfg := DefaultPowerCycleConfig()
	cfg.Enabled = true
	cfg.ArmFilePath = armPath
	cfg.RequireManualArm = true
	cfg.RetryAfterSeconds = 900 // 15 Minuten

	cycler, err := NewPowerCycler(cfg, store)
	if err != nil {
		t.Fatalf("Failed to create powercycler: %v", err)
	}

	// Attempt sollte blockiert sein (Retry-Cooldown aktiv)
	err = cycler.Attempt(context.Background())
	if err == nil {
		t.Error("Expected error when retry cooldown active, got nil")
	}
}

func TestPowerCycler_Attempt_RetryCooldownExpired(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "watcher_state.db")
	armPath := filepath.Join(tmpDir, "arm_powercycle")

	store, err := NewSQLiteStateStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Setze State mit alter Power Attempt (vor 1000 Sekunden, Cooldown=900 Sekunden)
	state := PersistedState{
		FailCount:        10,
		CurrentSeverity:  rules.SeverityCrit,
		LastEscalation:   time.Now(),
		PowerAttempts:    0,
		LastPowerAttempt: time.Now().Add(-1000 * time.Second), // Vor 1000 Sekunden
	}
	if err := store.Save(state); err != nil {
		t.Fatalf("Failed to save state: %v", err)
	}

	// Erstelle ARM-Datei
	if err := os.WriteFile(armPath, []byte(""), 0600); err != nil {
		t.Fatalf("Failed to create arm file: %v", err)
	}

	cfg := DefaultPowerCycleConfig()
	cfg.Enabled = true
	cfg.ArmFilePath = armPath
	cfg.RequireManualArm = true
	cfg.RetryAfterSeconds = 900 // 15 Minuten
	cfg.MaxAttempts = 2 // Erlaube 2 Versuche

	cycler, err := NewPowerCycler(cfg, store)
	if err != nil {
		t.Fatalf("Failed to create powercycler: %v", err)
	}

	// Attempt sollte erlaubt sein (Retry-Cooldown abgelaufen)
	err = cycler.Attempt(context.Background())
	if err != nil {
		t.Errorf("Expected no error when retry cooldown expired, got %v", err)
	}
}

func TestAllowedPowerCycle(t *testing.T) {
	tmpDir := t.TempDir()
	armPath := filepath.Join(tmpDir, "arm_powercycle")

	cfg := DefaultPowerCycleConfig()
	cfg.Enabled = true
	cfg.ArmFilePath = armPath
	cfg.RequireManualArm = true

	// Test: Nicht CRIT → nicht erlaubt
	state := PersistedState{
		CurrentSeverity: rules.SeverityWarn,
	}
	if AllowedPowerCycle(rules.SeverityWarn, cfg, state) {
		t.Error("Expected PowerCycle not allowed for WARN severity")
	}

	// Test: CRIT, aber nicht aktiviert → nicht erlaubt
	cfg.Enabled = false
	if AllowedPowerCycle(rules.SeverityCrit, cfg, state) {
		t.Error("Expected PowerCycle not allowed when disabled")
	}
	cfg.Enabled = true

	// Test: CRIT, aber ARM-Datei fehlt → nicht erlaubt
	if AllowedPowerCycle(rules.SeverityCrit, cfg, state) {
		t.Error("Expected PowerCycle not allowed when not armed")
	}

	// Test: CRIT, ARM vorhanden, aber Max Attempts erreicht → nicht erlaubt
	if err := os.WriteFile(armPath, []byte(""), 0600); err != nil {
		t.Fatalf("Failed to create arm file: %v", err)
	}
	state.PowerAttempts = 1
	cfg.MaxAttempts = 1
	if AllowedPowerCycle(rules.SeverityCrit, cfg, state) {
		t.Error("Expected PowerCycle not allowed when max attempts reached")
	}

	// Test: CRIT, ARM vorhanden, Max Attempts OK, aber Retry-Cooldown aktiv → nicht erlaubt
	state.PowerAttempts = 0
	state.LastPowerAttempt = time.Now().Add(-5 * time.Second) // Vor 5 Sekunden
	cfg.RetryAfterSeconds = 900
	if AllowedPowerCycle(rules.SeverityCrit, cfg, state) {
		t.Error("Expected PowerCycle not allowed when retry cooldown active")
	}

	// Test: Alle Bedingungen erfüllt → erlaubt
	state.LastPowerAttempt = time.Now().Add(-1000 * time.Second) // Vor 1000 Sekunden
	if !AllowedPowerCycle(rules.SeverityCrit, cfg, state) {
		t.Error("Expected PowerCycle allowed when all conditions met")
	}
}

func TestPowerCycler_Attempt_StoreError(t *testing.T) {
	// Test: Store-Fehler → kein Attempt
	cfg := DefaultPowerCycleConfig()
	cfg.Enabled = true
	cfg.RequireManualArm = false // Kein ARM erforderlich für diesen Test

	cycler, err := NewPowerCycler(cfg, nil) // nil Store
	if err != nil {
		t.Fatalf("Failed to create powercycler: %v", err)
	}

	err = cycler.Attempt(context.Background())
	if err == nil {
		t.Error("Expected error when store is nil, got nil")
	}
}
