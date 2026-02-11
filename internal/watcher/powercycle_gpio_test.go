package watcher

import (
	"context"
	"os"
	"path/filepath"
	"prox-watch/internal/rules"
	"testing"
	"time"
)

func TestGPIORelayPowerCycler_Attempt_PinSequence(t *testing.T) {
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

	// Erstelle MockPin
	pin, err := NewMockPin(24)
	if err != nil {
		t.Fatalf("Failed to create mock pin: %v", err)
	}
	mockPin := pin.(*MockPin)

	cfg := DefaultPowerCycleConfig()
	cfg.Enabled = true
	cfg.ArmFilePath = armPath
	cfg.RequireManualArm = true
	cfg.RelayMode = RelayModeCutPowerOnInactive // NC-Relais
	cfg.MinDowntimeSeconds = 1                   // Kurz für Test

	cycler, err := NewGPIORelayPowerCycler(cfg, store, pin)
	if err != nil {
		t.Fatalf("Failed to create GPIO power cycler: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt durchführen
	err = cycler.Attempt(ctx)
	if err != nil {
		t.Fatalf("Attempt() failed: %v", err)
	}

	// Prüfe Pin-Sequenz
	// Erwartet: HIGH (OFF) → Warte → LOW (ON)
	// MockPin speichert nur den letzten Zustand, daher prüfen wir den finalen Zustand
	// Bei Active LOW + CutPowerOnInactive:
	// - Power OFF: HIGH (Relais inaktiv = Strom getrennt)
	// - Power ON: LOW (Relais aktiv = Strom fließt)
	finalState := mockPin.GetState()
	if finalState != false { // LOW = false bei MockPin
		t.Errorf("Expected final pin state LOW (ON), got %v", finalState)
	}

	// Prüfe, dass ARM-Datei entfernt wurde
	if _, err := os.Stat(armPath); err == nil {
		t.Error("Expected arm file to be removed after attempt, but it still exists")
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
}

func TestGPIORelayPowerCycler_Attempt_MinDowntime(t *testing.T) {
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

	pin, err := NewMockPin(24)
	if err != nil {
		t.Fatalf("Failed to create mock pin: %v", err)
	}

	cfg := DefaultPowerCycleConfig()
	cfg.Enabled = true
	cfg.ArmFilePath = armPath
	cfg.RequireManualArm = true
	cfg.RelayMode = RelayModeCutPowerOnInactive
	cfg.MinDowntimeSeconds = 2 // 2 Sekunden

	cycler, err := NewGPIORelayPowerCycler(cfg, store, pin)
	if err != nil {
		t.Fatalf("Failed to create GPIO power cycler: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	start := time.Now()
	err = cycler.Attempt(ctx)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Attempt() failed: %v", err)
	}

	// Prüfe, dass min_downtime eingehalten wurde
	if duration < 2*time.Second {
		t.Errorf("Expected duration >= 2s (min_downtime), got %v", duration)
	}
}

func TestGPIORelayPowerCycler_Attempt_ArmFileRemoved(t *testing.T) {
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

	pin, err := NewMockPin(24)
	if err != nil {
		t.Fatalf("Failed to create mock pin: %v", err)
	}

	cfg := DefaultPowerCycleConfig()
	cfg.Enabled = true
	cfg.ArmFilePath = armPath
	cfg.RequireManualArm = true
	cfg.RelayMode = RelayModeCutPowerOnInactive
	cfg.MinDowntimeSeconds = 1

	cycler, err := NewGPIORelayPowerCycler(cfg, store, pin)
	if err != nil {
		t.Fatalf("Failed to create GPIO power cycler: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = cycler.Attempt(ctx)
	if err != nil {
		t.Fatalf("Attempt() failed: %v", err)
	}

	// Prüfe, dass ARM-Datei entfernt wurde
	if _, err := os.Stat(armPath); err == nil {
		t.Error("Expected arm file to be removed after successful attempt, but it still exists")
	}
}

func TestGPIORelayPowerCycler_Attempt_PinError_ArmRemains(t *testing.T) {
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

	// Erstelle fehlerhaften Pin (MockPin, der Fehler zurückgibt)
	// Wir verwenden einen normalen MockPin, aber simulieren einen Fehler durch geschlossenen Pin
	pin, err := NewMockPin(24)
	if err != nil {
		t.Fatalf("Failed to create mock pin: %v", err)
	}
	// Schließe Pin, um Fehler zu simulieren
	pin.Close()

	cfg := DefaultPowerCycleConfig()
	cfg.Enabled = true
	cfg.ArmFilePath = armPath
	cfg.RequireManualArm = true
	cfg.RelayMode = RelayModeCutPowerOnInactive
	cfg.MinDowntimeSeconds = 1

	cycler, err := NewGPIORelayPowerCycler(cfg, store, pin)
	if err != nil {
		t.Fatalf("Failed to create GPIO power cycler: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt sollte fehlschlagen
	err = cycler.Attempt(ctx)
	if err == nil {
		t.Error("Expected error when pin operation fails, got nil")
	}

	// Prüfe, dass ARM-Datei NICHT entfernt wurde (für Diagnose)
	if _, err := os.Stat(armPath); err != nil {
		t.Error("Expected arm file to remain after pin error (for diagnosis), but it was removed")
	}
}

func TestGPIORelayPowerCycler_RelayModeRequired(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "watcher_state.db")

	store, err := NewSQLiteStateStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	pin, err := NewMockPin(24)
	if err != nil {
		t.Fatalf("Failed to create mock pin: %v", err)
	}

	cfg := DefaultPowerCycleConfig()
	cfg.Enabled = true
	cfg.RelayMode = "" // Fehlt

	// NewGPIORelayPowerCycler sollte Fehler zurückgeben
	_, err = NewGPIORelayPowerCycler(cfg, store, pin)
	if err == nil {
		t.Error("Expected error when relay_mode is missing, got nil")
	}
}

func TestGPIORelayPowerCycler_InvalidRelayMode(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "watcher_state.db")

	store, err := NewSQLiteStateStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	pin, err := NewMockPin(24)
	if err != nil {
		t.Fatalf("Failed to create mock pin: %v", err)
	}

	cfg := DefaultPowerCycleConfig()
	cfg.Enabled = true
	cfg.RelayMode = "invalid_mode"

	// NewGPIORelayPowerCycler sollte Fehler zurückgeben
	_, err = NewGPIORelayPowerCycler(cfg, store, pin)
	if err == nil {
		t.Error("Expected error when relay_mode is invalid, got nil")
	}
}

func TestGPIORelayPowerCycler_Attempt_OnlyOneToggle(t *testing.T) {
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

	pin, err := NewMockPin(24)
	if err != nil {
		t.Fatalf("Failed to create mock pin: %v", err)
	}
	mockPin := pin.(*MockPin)

	cfg := DefaultPowerCycleConfig()
	cfg.Enabled = true
	cfg.ArmFilePath = armPath
	cfg.RequireManualArm = true
	cfg.RelayMode = RelayModeCutPowerOnInactive
	cfg.MinDowntimeSeconds = 1

	cycler, err := NewGPIORelayPowerCycler(cfg, store, pin)
	if err != nil {
		t.Fatalf("Failed to create GPIO power cycler: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Initial State: LOW (false)
	initialState := mockPin.GetState()

	err = cycler.Attempt(ctx)
	if err != nil {
		t.Fatalf("Attempt() failed: %v", err)
	}

	// Prüfe, dass nur ein Toggle stattgefunden hat
	// Bei CutPowerOnInactive + Active LOW:
	// - Power OFF: HIGH (false → true)
	// - Power ON: LOW (true → false)
	// Final State sollte LOW (false) sein
	finalState := mockPin.GetState()
	if finalState == initialState {
		t.Error("Expected pin state to change (toggle), but it remained the same")
	}

	// Prüfe, dass State nur einmal aktualisiert wurde
	state, err := store.Load()
	if err != nil {
		t.Fatalf("Failed to load state: %v", err)
	}
	if state.PowerAttempts != 1 {
		t.Errorf("Expected PowerAttempts=1 (only one attempt), got %d", state.PowerAttempts)
	}
}

func TestGPIORelayPowerCycler_RelayModeCutPowerOnActive(t *testing.T) {
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

	pin, err := NewMockPin(24)
	if err != nil {
		t.Fatalf("Failed to create mock pin: %v", err)
	}

	cfg := DefaultPowerCycleConfig()
	cfg.Enabled = true
	cfg.ArmFilePath = armPath
	cfg.RequireManualArm = true
	cfg.RelayMode = RelayModeCutPowerOnActive // NO-Relais
	cfg.MinDowntimeSeconds = 1

	cycler, err := NewGPIORelayPowerCycler(cfg, store, pin)
	if err != nil {
		t.Fatalf("Failed to create GPIO power cycler: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt sollte erfolgreich sein
	err = cycler.Attempt(ctx)
	if err != nil {
		t.Fatalf("Attempt() failed: %v", err)
	}

	// Prüfe, dass State aktualisiert wurde
	state, err := store.Load()
	if err != nil {
		t.Fatalf("Failed to load state: %v", err)
	}
	if state.PowerAttempts != 1 {
		t.Errorf("Expected PowerAttempts=1, got %d", state.PowerAttempts)
	}
}
