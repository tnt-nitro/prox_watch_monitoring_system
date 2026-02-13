package cli

import (
	"os"
	"testing"
	"time"

	"prox-watch/internal/config"
	"prox-watch/internal/rules"
	"prox-watch/internal/state"
)

func TestRunStatus(t *testing.T) {
	// Create temporary state DB
	tmpDir := t.TempDir()
	statePath := tmpDir + "/state.db"

	cfg := config.DefaultConfig()
	cfg.Paths.StateDB = statePath

	// Create store and add test events
	store, err := state.NewSQLiteStore(statePath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Add test events
	now := time.Now()
	store.Increment("test.event.1", now)
	store.Increment("test.event.2", now.Add(1*time.Minute))

	// Create config file
	configPath := tmpDir + "/config.yaml"
	if err := cfg.Save(configPath); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Run status
	if err := RunStatus(configPath); err != nil {
		t.Errorf("RunStatus() error = %v", err)
	}
}

func TestRunStatusEvent(t *testing.T) {
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

	// Run status for specific event
	if err := RunStatusEvent(configPath, "test.event.1"); err != nil {
		t.Errorf("RunStatusEvent() error = %v", err)
	}
}

func TestRunStatusEvent_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := tmpDir + "/config.yaml"

	cfg := config.DefaultConfig()
	cfg.Paths.StateDB = tmpDir + "/state.db"
	if err := cfg.Save(configPath); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Run status for non-existent event
	err := RunStatusEvent(configPath, "nonexistent.event")
	if err == nil {
		t.Error("RunStatusEvent() should return error for non-existent event")
	}
}
