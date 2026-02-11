package core

import (
	"context"
	"os"
	"testing"
	"time"

	"prox-watch/internal/config"
	"prox-watch/internal/journal"
	"prox-watch/internal/pattern"
	"prox-watch/internal/rules"
	"prox-watch/internal/state"
)

func TestRunner_ProcessEntry(t *testing.T) {
	// Create temporary state DB
	tmpDir := t.TempDir()
	statePath := tmpDir + "/state.db"

	cfg := config.DefaultConfig()
	cfg.Paths.StateDB = statePath

	// Create runner
	runner, err := NewRunner(cfg)
	if err != nil {
		t.Fatalf("NewRunner() error = %v", err)
	}
	defer runner.Stop()

	// Create fake journal entry
	now := time.Now()
	entry := journal.Entry{
		Priority:  6, // ERROR
		Source:    "kernel",
		Timestamp: now,
	}

	// Manually register a pattern in matcher
	// For testing, we create a pattern and load it
	testPattern := pattern.Pattern{
		PatternID: "host.kernel.error",
		Source:    "kernel",
		MatchType: pattern.MatchTypeEvent,
		Severity:  rules.SeverityCrit,
	}
	// Use LoadPatterns with a test file or directly access registry
	// For MVP, we'll skip pattern loading in this test
	_ = testPattern

	// Process entry
	ctx := context.Background()
	if err := runner.ProcessEntry(ctx, entry); err != nil {
		t.Fatalf("ProcessEntry() error = %v", err)
	}

	// Verify state was updated
	countState, err := runner.store.Get("host.kernel.error")
	if err != nil {
		t.Fatalf("store.Get() error = %v", err)
	}

	if countState.Count != 1 {
		t.Errorf("ProcessEntry() count = %d, want 1", countState.Count)
	}
}

func TestRunner_Stop(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := tmpDir + "/state.db"

	cfg := config.DefaultConfig()
	cfg.Paths.StateDB = statePath

	runner, err := NewRunner(cfg)
	if err != nil {
		t.Fatalf("NewRunner() error = %v", err)
	}

	if err := runner.Stop(); err != nil {
		t.Errorf("Stop() error = %v", err)
	}

	// Verify store is closed
	// This would fail if store is already closed
	_, err = runner.store.Get("test")
	if err == nil {
		t.Error("Stop() should close store")
	}
}

func TestLifecycle_StartStop(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := tmpDir + "/state.db"

	cfg := config.DefaultConfig()
	cfg.Paths.StateDB = statePath

	lifecycle, err := NewLifecycle(cfg)
	if err != nil {
		t.Fatalf("NewLifecycle() error = %v", err)
	}

	// Start lifecycle in goroutine
	errChan := make(chan error, 1)
	go func() {
		// This will block until signal
		errChan <- lifecycle.Start()
	}()

	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)

	// Stop lifecycle
	if err := lifecycle.Stop(); err != nil {
		t.Errorf("Stop() error = %v", err)
	}

	// Wait for start to finish
	select {
	case err := <-errChan:
		if err != nil {
			t.Errorf("Start() error = %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Error("Start() did not return")
	}
}

