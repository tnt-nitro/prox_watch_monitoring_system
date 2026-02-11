package state

import (
	"testing"
	"time"

	"prox-watch/internal/rules"
)

func TestSQLiteStore_Increment(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewSQLiteStore(tmpDir + "/test.db")
	if err != nil {
		t.Fatalf("NewSQLiteStore() error = %v", err)
	}
	defer store.Close()

	now := time.Now()
	countState, err := store.Increment("test.event.1", now)
	if err != nil {
		t.Fatalf("Increment() error = %v", err)
	}

	if countState.Count != 1 {
		t.Errorf("Increment() count = %d, want 1", countState.Count)
	}

	// Increment again
	countState, err = store.Increment("test.event.1", now.Add(1*time.Minute))
	if err != nil {
		t.Fatalf("Increment() error = %v", err)
	}

	if countState.Count != 2 {
		t.Errorf("Increment() count = %d, want 2", countState.Count)
	}
}

func TestSQLiteStore_Get(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewSQLiteStore(tmpDir + "/test.db")
	if err != nil {
		t.Fatalf("NewSQLiteStore() error = %v", err)
	}
	defer store.Close()

	now := time.Now()
	store.Increment("test.event.1", now)

	countState, err := store.Get("test.event.1")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if countState.Count != 1 {
		t.Errorf("Get() count = %d, want 1", countState.Count)
	}

	if countState.EventID != "test.event.1" {
		t.Errorf("Get() EventID = %s, want test.event.1", countState.EventID)
	}
}

func TestSQLiteStore_Get_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewSQLiteStore(tmpDir + "/test.db")
	if err != nil {
		t.Fatalf("NewSQLiteStore() error = %v", err)
	}
	defer store.Close()

	_, err = store.Get("nonexistent.event")
	if err == nil {
		t.Error("Get() should return error for non-existent event")
	}
}

func TestSQLiteStore_SetCooldown(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewSQLiteStore(tmpDir + "/test.db")
	if err != nil {
		t.Fatalf("NewSQLiteStore() error = %v", err)
	}
	defer store.Close()

	now := time.Now()
	until := now.Add(30 * time.Minute)

	// First increment to create event
	store.Increment("test.event.1", now)

	if err := store.SetCooldown("test.event.1", until); err != nil {
		t.Fatalf("SetCooldown() error = %v", err)
	}

	if !store.IsCooldown("test.event.1", now) {
		t.Error("IsCooldown() should return true during cooldown")
	}

	// After cooldown expires
	future := now.Add(1 * time.Hour)
	if store.IsCooldown("test.event.1", future) {
		t.Error("IsCooldown() should return false after cooldown expires")
	}
}

func TestSQLiteStore_Ack(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewSQLiteStore(tmpDir + "/test.db")
	if err != nil {
		t.Fatalf("NewSQLiteStore() error = %v", err)
	}
	defer store.Close()

	now := time.Now()
	until := now.Add(24 * time.Hour)

	// First increment to create event
	store.Increment("test.event.1", now)

	if err := store.Ack("test.event.1", until); err != nil {
		t.Fatalf("Ack() error = %v", err)
	}

	if !store.IsAcked("test.event.1", now) {
		t.Error("IsAcked() should return true when acknowledged")
	}

	// After ack expires
	future := now.Add(25 * time.Hour)
	if store.IsAcked("test.event.1", future) {
		t.Error("IsAcked() should return false after ack expires")
	}
}

func TestSQLiteStore_SetSeverity(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewSQLiteStore(tmpDir + "/test.db")
	if err != nil {
		t.Fatalf("NewSQLiteStore() error = %v", err)
	}
	defer store.Close()

	now := time.Now()
	store.Increment("test.event.1", now)

	if err := store.SetSeverity("test.event.1", rules.SeverityCrit); err != nil {
		t.Fatalf("SetSeverity() error = %v", err)
	}

	event, err := store.GetEvent("test.event.1")
	if err != nil {
		t.Fatalf("GetEvent() error = %v", err)
	}

	if event.Severity != rules.SeverityCrit {
		t.Errorf("SetSeverity() severity = %v, want %v", event.Severity, rules.SeverityCrit)
	}
}

func TestSQLiteStore_GetAllEvents(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewSQLiteStore(tmpDir + "/test.db")
	if err != nil {
		t.Fatalf("NewSQLiteStore() error = %v", err)
	}
	defer store.Close()

	now := time.Now()
	store.Increment("test.event.1", now)
	store.Increment("test.event.2", now.Add(1*time.Minute))
	store.Increment("test.event.3", now.Add(2*time.Minute))

	events, err := store.GetAllEvents()
	if err != nil {
		t.Fatalf("GetAllEvents() error = %v", err)
	}

	if len(events) != 3 {
		t.Errorf("GetAllEvents() len = %d, want 3", len(events))
	}

	// Should be sorted by last_seen DESC
	if events[0].EventID != "test.event.3" {
		t.Errorf("GetAllEvents() first event = %s, want test.event.3", events[0].EventID)
	}
}

func TestSQLiteStore_Persistence(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test.db"

	// Create store and add event
	store1, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteStore() error = %v", err)
	}

	now := time.Now()
	store1.Increment("test.event.1", now)
	store1.Close()

	// Reopen store
	store2, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteStore() error = %v", err)
	}
	defer store2.Close()

	countState, err := store2.Get("test.event.1")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if countState.Count != 1 {
		t.Errorf("Persistence: count = %d, want 1", countState.Count)
	}
}
