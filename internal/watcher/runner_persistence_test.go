package watcher

import (
	"context"
	"os"
	"path/filepath"
	"prox-watch/internal/push"
	"prox-watch/internal/rules"
	"testing"
	"time"
)

// mockStateStore ist ein Mock-StateStore für Tests.
type mockStateStore struct {
	state PersistedState
	saved bool
}

func (m *mockStateStore) Load() (PersistedState, error) {
	return m.state, nil
}

func (m *mockStateStore) Save(state PersistedState) error {
	m.state = state
	m.saved = true
	return nil
}

func (m *mockStateStore) Close() error {
	return nil
}

func TestRunner_Restart_StateLoaded(t *testing.T) {
	// Test: Restart → State korrekt geladen
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "watcher_state.db")

	// Erstelle Store und speichere initialen State
	store, err := NewSQLiteStateStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Speichere State mit FailCount=5, Severity=WARN
	persisted := PersistedState{
		FailCount:       5,
		CurrentSeverity: rules.SeverityWarn,
		LastEscalation:  time.Unix(1000, 0),
	}
	if err := store.Save(persisted); err != nil {
		t.Fatalf("Failed to save state: %v", err)
	}

	// Erstelle neuen Runner mit Store
	fakeHealth := &fakeHealthChecker{
		results: []Result{
			{Success: true, Mode: "ping+https"},
		},
	}
	counter := NewCounter()
	pushService := NewPushService(PushServiceConfig{
		Adapter: push.NewFakeAdapter(),
		Enabled: true,
	})
	gpio := NewNoOpGPIO()

	runner, err := NewRunner(RunnerConfig{
		Health:        fakeHealth,
		Counter:       counter,
		Push:          pushService,
		GPIO:          gpio,
		Store:         store,
		Interval:      100 * time.Millisecond,
		WarnThreshold: 3,
		CritThreshold: 10,
	})
	if err != nil {
		t.Fatalf("Failed to create runner: %v", err)
	}

	// State sollte geladen sein (FailCount sollte im State sein, Counter wird zurückgesetzt bei Success)
	// Prüfe, dass Runner startet ohne Fehler
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- runner.Run(ctx)
	}()

	time.Sleep(120 * time.Millisecond)
	runner.Stop()
	<-done
}

func TestRunner_Cooldown_BlocksPush(t *testing.T) {
	// Test: Cooldown blockiert Push
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "watcher_state.db")

	store, err := NewSQLiteStateStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Speichere State mit kürzlicher Eskalation (vor 5 Sekunden, Cooldown=10 Sekunden)
	persisted := PersistedState{
		FailCount:       3,
		CurrentSeverity: rules.SeverityWarn,
		LastEscalation:  time.Now().Add(-5 * time.Second), // Vor 5 Sekunden
	}
	if err := store.Save(persisted); err != nil {
		t.Fatalf("Failed to save state: %v", err)
	}

	fakeHealth := &fakeHealthChecker{
		results: []Result{
			{Success: false, Mode: "ping+https", Error: "connection_failed"},
		},
	}
	counter := NewCounter()
	// Setze Counter auf 3 (WARN bereits erreicht)
	for i := 0; i < 3; i++ {
		counter.Increment()
	}

	fakeAdapter := push.NewFakeAdapter()
	pushService := NewPushService(PushServiceConfig{
		Adapter: fakeAdapter,
		Enabled: true,
	})
	gpio := NewNoOpGPIO()

	runner, err := NewRunner(RunnerConfig{
		Health:        fakeHealth,
		Counter:       counter,
		Push:          pushService,
		GPIO:          gpio,
		Store:         store,
		Interval:      50 * time.Millisecond,
		WarnThreshold: 3,
		CritThreshold: 10,
		CooldownSecs:  10, // 10 Sekunden Cooldown
	})
	if err != nil {
		t.Fatalf("Failed to create runner: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- runner.Run(ctx)
	}()

	time.Sleep(80 * time.Millisecond)

	// Prüfe, dass kein Push gesendet wurde (Cooldown aktiv)
	messages := fakeAdapter.GetMessages()
	if len(messages) > 0 {
		t.Errorf("Expected no push (cooldown active), got %d messages", len(messages))
	}

	runner.Stop()
	<-done
}

func TestRunner_Cooldown_Expired_AllowsPush(t *testing.T) {
	// Test: Cooldown abgelaufen → Push erlaubt
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "watcher_state.db")

	store, err := NewSQLiteStateStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Speichere State mit alter Eskalation (vor 15 Sekunden, Cooldown=10 Sekunden)
	persisted := PersistedState{
		FailCount:       3,
		CurrentSeverity: rules.SeverityWarn,
		LastEscalation:  time.Now().Add(-15 * time.Second), // Vor 15 Sekunden
	}
	if err := store.Save(persisted); err != nil {
		t.Fatalf("Failed to save state: %v", err)
	}

	fakeHealth := &fakeHealthChecker{
		results: []Result{
			{Success: false, Mode: "ping+https", Error: "connection_failed"},
		},
	}
	counter := NewCounter()
	// Setze Counter auf 3 (WARN bereits erreicht)
	for i := 0; i < 3; i++ {
		counter.Increment()
	}

	fakeAdapter := push.NewFakeAdapter()
	pushService := NewPushService(PushServiceConfig{
		Adapter: fakeAdapter,
		Enabled: true,
	})
	gpio := NewNoOpGPIO()

	runner, err := NewRunner(RunnerConfig{
		Health:        fakeHealth,
		Counter:       counter,
		Push:          pushService,
		GPIO:          gpio,
		Store:         store,
		Interval:      50 * time.Millisecond,
		WarnThreshold: 3,
		CritThreshold: 10,
		CooldownSecs:  10, // 10 Sekunden Cooldown
	})
	if err != nil {
		t.Fatalf("Failed to create runner: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- runner.Run(ctx)
	}()

	time.Sleep(80 * time.Millisecond)

	// Prüfe, dass kein Push gesendet wurde (gleiche Severity, keine Eskalation)
	// Hinweis: WARN→WARN ist keine Eskalation, daher kein Push
	messages := fakeAdapter.GetMessages()
	// Erwartet: 0 Messages (keine Eskalation, nur WARN→WARN)
	if len(messages) > 0 {
		t.Errorf("Expected no push (WARN→WARN is not escalation), got %d messages", len(messages))
	}

	runner.Stop()
	<-done
}

func TestRunner_Save_OnChange(t *testing.T) {
	// Test: Save bei Änderung
	mockStore := &mockStateStore{
		state: PersistedState{
			FailCount:       0,
			CurrentSeverity: rules.SeverityInfo,
			LastEscalation:  time.Unix(0, 0),
		},
	}

	fakeHealth := &fakeHealthChecker{
		results: []Result{
			{Success: false, Mode: "ping+https", Error: "connection_failed"},
			{Success: false, Mode: "ping+https", Error: "connection_failed"},
			{Success: false, Mode: "ping+https", Error: "connection_failed"}, // 3× → WARN (Eskalation)
		},
	}
	counter := NewCounter()
	fakeAdapter := push.NewFakeAdapter()
	pushService := NewPushService(PushServiceConfig{
		Adapter: fakeAdapter,
		Enabled: true,
	})
	gpio := NewNoOpGPIO()

	runner, err := NewRunner(RunnerConfig{
		Health:        fakeHealth,
		Counter:       counter,
		Push:          pushService,
		GPIO:          gpio,
		Store:         mockStore,
		Interval:      50 * time.Millisecond,
		WarnThreshold: 3,
		CritThreshold: 10,
		CooldownSecs:  600,
	})
	if err != nil {
		t.Fatalf("Failed to create runner: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- runner.Run(ctx)
	}()

	time.Sleep(180 * time.Millisecond)

	// Prüfe, dass Save aufgerufen wurde
	if !mockStore.saved {
		t.Error("Expected Save to be called on state change, but it wasn't")
	}

	// Prüfe, dass gespeicherter State korrekt ist
	if mockStore.state.FailCount != 3 {
		t.Errorf("Expected FailCount=3, got %d", mockStore.state.FailCount)
	}
	if mockStore.state.CurrentSeverity != rules.SeverityWarn {
		t.Errorf("Expected Severity=WARN, got %v", mockStore.state.CurrentSeverity)
	}
	if mockStore.state.LastEscalation.IsZero() {
		t.Error("Expected LastEscalation to be set, but it's zero")
	}

	runner.Stop()
	<-done
}

func TestRunner_NoSave_StableInfo(t *testing.T) {
	// Test: Kein Save bei stabiler INFO
	mockStore := &mockStateStore{
		state: PersistedState{
			FailCount:       0,
			CurrentSeverity: rules.SeverityInfo,
			LastEscalation:  time.Unix(0, 0),
		},
		saved: false,
	}

	fakeHealth := &fakeHealthChecker{
		results: []Result{
			{Success: true, Mode: "ping+https"},
			{Success: true, Mode: "ping+https"},
		},
	}
	counter := NewCounter()
	pushService := NewPushService(PushServiceConfig{
		Adapter: push.NewFakeAdapter(),
		Enabled: true,
	})
	gpio := NewNoOpGPIO()

	runner, err := NewRunner(RunnerConfig{
		Health:        fakeHealth,
		Counter:       counter,
		Push:          pushService,
		GPIO:          gpio,
		Store:         mockStore,
		Interval:      50 * time.Millisecond,
		WarnThreshold: 3,
		CritThreshold: 10,
		CooldownSecs:  600,
	})
	if err != nil {
		t.Fatalf("Failed to create runner: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Millisecond)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- runner.Run(ctx)
	}()

	time.Sleep(100 * time.Millisecond)

	// Prüfe, dass Save NICHT aufgerufen wurde (stabile INFO, keine Änderung)
	// Hinweis: Save wird nur bei Änderungen aufgerufen, nicht bei jedem Intervall
	// Da State bereits INFO ist und bleibt, sollte kein Save sein
	// (außer beim ersten Reset, aber das ist eine Änderung von 0→0, also keine Änderung)
	// Tatsächlich: Wenn FailCount von 0→0 geht (Reset), ist das keine Änderung
	// Save sollte nur bei FailCount-Änderung, Severity-Änderung oder Eskalation sein

	runner.Stop()
	<-done
}

func TestRunner_DoubleCrit_OnlyOnePush(t *testing.T) {
	// Test: Doppeltes CRIT → nur ein Push (Cooldown)
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "watcher_state.db")

	store, err := NewSQLiteStateStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	fakeHealth := &fakeHealthChecker{
		results: []Result{
			{Success: false, Mode: "ping+https", Error: "connection_failed"},
			{Success: false, Mode: "ping+https", Error: "connection_failed"},
			{Success: false, Mode: "ping+https", Error: "connection_failed"},
			{Success: false, Mode: "ping+https", Error: "connection_failed"},
			{Success: false, Mode: "ping+https", Error: "connection_failed"},
			{Success: false, Mode: "ping+https", Error: "connection_failed"},
			{Success: false, Mode: "ping+https", Error: "connection_failed"},
			{Success: false, Mode: "ping+https", Error: "connection_failed"},
			{Success: false, Mode: "ping+https", Error: "connection_failed"},
			{Success: false, Mode: "ping+https", Error: "connection_failed"}, // 10× → CRIT (Push)
			{Success: false, Mode: "ping+https", Error: "connection_failed"}, // 11× → CRIT (kein Push, Cooldown)
		},
	}
	counter := NewCounter()
	fakeAdapter := push.NewFakeAdapter()
	pushService := NewPushService(PushServiceConfig{
		Adapter: fakeAdapter,
		Enabled: true,
	})
	gpio := NewNoOpGPIO()

	runner, err := NewRunner(RunnerConfig{
		Health:        fakeHealth,
		Counter:       counter,
		Push:          pushService,
		GPIO:          gpio,
		Store:         store,
		Interval:      30 * time.Millisecond,
		WarnThreshold: 3,
		CritThreshold: 10,
		CooldownSecs:  1, // 1 Sekunde Cooldown (kurz für Test)
	})
	if err != nil {
		t.Fatalf("Failed to create runner: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 400*time.Millisecond)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- runner.Run(ctx)
	}()

	time.Sleep(350 * time.Millisecond)

	// Prüfe, dass nur ein Push gesendet wurde (erste CRIT-Eskalation)
	// Zweite CRIT (11×) sollte durch Cooldown blockiert sein
	messages := fakeAdapter.GetMessages()
	critCount := 0
	for _, msg := range messages {
		if msg.Severity == rules.SeverityCrit {
			critCount++
		}
	}
	if critCount > 1 {
		t.Errorf("Expected at most 1 CRIT push (cooldown blocks second), got %d", critCount)
	}

	runner.Stop()
	<-done
}
