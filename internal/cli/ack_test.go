package cli

import (
	"testing"
	"time"

	"prox-watch/internal/config"
	"prox-watch/internal/rules"
	"prox-watch/internal/state"
)

func TestRunAck(t *testing.T) {
	// Create temporary state DB
	tmpDir := t.TempDir()
	statePath := tmpDir + "/state.db"

	cfg := config.DefaultConfig()
	cfg.Paths.StateDB = statePath

	// Create store and add test event
	store, err := state.NewSQLiteStore(statePath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Add test event
	now := time.Now()
	store.Increment("test.event.1", now)
	store.SetSeverity("test.event.1", int(rules.SeverityCrit))

	// Create config file
	configPath := tmpDir + "/config.yaml"
	if err := cfg.Save(configPath); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Acknowledge event
	duration := 1 * time.Hour
	if err := RunAck(configPath, "test.event.1", duration); err != nil {
		t.Errorf("RunAck() error = %v", err)
	}

	// Verify acknowledge
	if !store.IsAcked("test.event.1", time.Now()) {
		t.Error("Event should be acknowledged")
	}
}

func TestRunAck_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := tmpDir + "/config.yaml"

	cfg := config.DefaultConfig()
	cfg.Paths.StateDB = tmpDir + "/state.db"
	if err := cfg.Save(configPath); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Try to acknowledge non-existent event
	err := RunAck(configPath, "nonexistent.event", 1*time.Hour)
	if err == nil {
		t.Error("RunAck() should return error for non-existent event")
	}
}

func TestRunAckDefault(t *testing.T) {
	// Create temporary state DB
	tmpDir := t.TempDir()
	statePath := tmpDir + "/state.db"

	cfg := config.DefaultConfig()
	cfg.Paths.StateDB = statePath

	// Create store and add test event
	store, err := state.NewSQLiteStore(statePath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Add test event
	now := time.Now()
	store.Increment("test.event.1", now)

	// Create config file
	configPath := tmpDir + "/config.yaml"
	if err := cfg.Save(configPath); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Acknowledge with default duration (24h)
	if err := RunAckDefault(configPath, "test.event.1"); err != nil {
		t.Errorf("RunAckDefault() error = %v", err)
	}

	// Verify acknowledge (should be valid for 24h)
	if !store.IsAcked("test.event.1", time.Now()) {
		t.Error("Event should be acknowledged")
	}
}
