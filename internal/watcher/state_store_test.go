package watcher

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"prox-watch/internal/rules"
)

// helper to create temp DB path
func newTempDBPath(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	return filepath.Join(dir, "watcher_state.db")
}

func TestStateStore_InitialLoadReturnsDefault(t *testing.T) {
	path := newTempDBPath(t)

	storeIface, err := NewSQLiteStateStore(path)
	if err != nil {
		t.Fatalf("NewSQLiteStateStore() error: %v", err)
	}
	defer storeIface.Close()

	state, err := storeIface.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if state.FailCount != 0 {
		t.Errorf("expected FailCount=0, got %d", state.FailCount)
	}
	if state.CurrentSeverity != rules.SeverityInfo {
		t.Errorf("expected CurrentSeverity=INFO, got %v", state.CurrentSeverity)
	}
	if !state.LastEscalation.Equal(time.Unix(0, 0)) {
		t.Errorf("expected LastEscalation=Unix(0), got %v", state.LastEscalation)
	}
}

func TestStateStore_SaveAndLoadRoundTrip(t *testing.T) {
	path := newTempDBPath(t)

	storeIface, err := NewSQLiteStateStore(path)
	if err != nil {
		t.Fatalf("NewSQLiteStateStore() error: %v", err)
	}
	defer storeIface.Close()

	now := time.Now().Truncate(time.Second)

	orig := PersistedState{
		FailCount:       5,
		CurrentSeverity: rules.SeverityWarn,
		LastEscalation:  now,
	}

	if err := storeIface.Save(orig); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	loaded, err := storeIface.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if loaded.FailCount != orig.FailCount {
		t.Errorf("expected FailCount=%d, got %d", orig.FailCount, loaded.FailCount)
	}
	if loaded.CurrentSeverity != orig.CurrentSeverity {
		t.Errorf("expected CurrentSeverity=%v, got %v", orig.CurrentSeverity, loaded.CurrentSeverity)
	}
	if !loaded.LastEscalation.Equal(orig.LastEscalation) {
		t.Errorf("expected LastEscalation=%v, got %v", orig.LastEscalation, loaded.LastEscalation)
	}
}

func TestStateStore_SaveMultipleOverwrites(t *testing.T) {
	path := newTempDBPath(t)

	storeIface, err := NewSQLiteStateStore(path)
	if err != nil {
		t.Fatalf("NewSQLiteStateStore() error: %v", err)
	}
	defer storeIface.Close()

	first := PersistedState{
		FailCount:       1,
		CurrentSeverity: rules.SeverityWarn,
		LastEscalation:  time.Unix(1000, 0),
	}
	second := PersistedState{
		FailCount:       10,
		CurrentSeverity: rules.SeverityCrit,
		LastEscalation:  time.Unix(2000, 0),
	}

	if err := storeIface.Save(first); err != nil {
		t.Fatalf("Save(first) error: %v", err)
	}
	if err := storeIface.Save(second); err != nil {
		t.Fatalf("Save(second) error: %v", err)
	}

	loaded, err := storeIface.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if loaded.FailCount != second.FailCount {
		t.Errorf("expected FailCount=%d, got %d", second.FailCount, loaded.FailCount)
	}
	if loaded.CurrentSeverity != second.CurrentSeverity {
		t.Errorf("expected CurrentSeverity=%v, got %v", second.CurrentSeverity, loaded.CurrentSeverity)
	}
	if !loaded.LastEscalation.Equal(second.LastEscalation) {
		t.Errorf("expected LastEscalation=%v, got %v", second.LastEscalation, loaded.LastEscalation)
	}
}

func TestStateStore_CreatesDBFile(t *testing.T) {
	path := newTempDBPath(t)

	_, err := NewSQLiteStateStore(path)
	if err != nil {
		t.Fatalf("NewSQLiteStateStore() error: %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected DB file to exist at %s, stat error: %v", path, err)
	}
}

