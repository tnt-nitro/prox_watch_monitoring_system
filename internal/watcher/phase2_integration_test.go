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

// TestPhase2_Integration_Restart_ActiveCooldown testet:
// CRIT wurde gespeichert → Neustart → Cooldown noch aktiv → Kein Push
func TestPhase2_Integration_Restart_ActiveCooldown(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "watcher_state.db")

	// Erster Lauf: CRIT erreichen und Push senden
	store1, err := NewSQLiteStateStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store1: %v", err)
	}
	defer store1.Close()

	fakeHealth1 := &fakeHealthChecker{
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
			{Success: false, Mode: "ping+https", Error: "connection_failed"}, // 10× → CRIT
		},
	}
	counter1 := NewCounter()
	fakeAdapter1 := push.NewFakeAdapter()
	pushService1 := NewPushService(PushServiceConfig{
		Adapter: fakeAdapter1,
		Enabled: true,
	})
	gpio1 := NewNoOpGPIO()

	runner1, err := NewRunner(RunnerConfig{
		Health:        fakeHealth1,
		Counter:       counter1,
		Push:          pushService1,
		GPIO:          gpio1,
		Store:         store1,
		Interval:      30 * time.Millisecond,
		WarnThreshold: 3,
		CritThreshold: 10,
		CooldownSecs:  5, // 5 Sekunden Cooldown (kurz für Test)
	})
	if err != nil {
		t.Fatalf("Failed to create runner1: %v", err)
	}

	ctx1, cancel1 := context.WithTimeout(context.Background(), 400*time.Millisecond)
	defer cancel1()

	done1 := make(chan error, 1)
	go func() {
		done1 <- runner1.Run(ctx1)
	}()

	time.Sleep(350 * time.Millisecond)
	runner1.Stop()
	<-done1
	store1.Close()

	// Prüfe, dass CRIT-Push gesendet wurde
	messages1 := fakeAdapter1.GetMessages()
	critFound := false
	for _, msg := range messages1 {
		if msg.Severity == rules.SeverityCrit {
			critFound = true
		}
	}
	if !critFound {
		t.Fatal("Expected CRIT push in first run, got none")
	}

	// Prüfe, dass State gespeichert wurde
	persisted, err := store1.Load()
	if err != nil {
		t.Fatalf("Failed to load state: %v", err)
	}
	if persisted.CurrentSeverity != rules.SeverityCrit {
		t.Fatalf("Expected Severity=CRIT, got %v", persisted.CurrentSeverity)
	}
	if persisted.LastEscalation.IsZero() {
		t.Fatal("Expected LastEscalation to be set, got zero")
	}

	// Zweiter Lauf: Neustart mit aktivem Cooldown
	store2, err := NewSQLiteStateStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store2: %v", err)
	}
	defer store2.Close()

	fakeHealth2 := &fakeHealthChecker{
		results: []Result{
			{Success: false, Mode: "ping+https", Error: "connection_failed"}, // Weiterhin CRIT
		},
	}
	counter2 := NewCounter()
	// Setze Counter auf 10 (CRIT bereits erreicht)
	for i := 0; i < 10; i++ {
		counter2.Increment()
	}

	fakeAdapter2 := push.NewFakeAdapter()
	pushService2 := NewPushService(PushServiceConfig{
		Adapter: fakeAdapter2,
		Enabled: true,
	})
	gpio2 := NewNoOpGPIO()

	runner2, err := NewRunner(RunnerConfig{
		Health:        fakeHealth2,
		Counter:       counter2,
		Push:          pushService2,
		GPIO:          gpio2,
		Store:         store2,
		Interval:      50 * time.Millisecond,
		WarnThreshold: 3,
		CritThreshold: 10,
		CooldownSecs:  5, // 5 Sekunden Cooldown
	})
	if err != nil {
		t.Fatalf("Failed to create runner2: %v", err)
	}

	ctx2, cancel2 := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel2()

	done2 := make(chan error, 1)
	go func() {
		done2 <- runner2.Run(ctx2)
	}()

	time.Sleep(80 * time.Millisecond)
	runner2.Stop()
	<-done2

	// Prüfe, dass KEIN Push gesendet wurde (Cooldown aktiv)
	messages2 := fakeAdapter2.GetMessages()
	if len(messages2) > 0 {
		t.Errorf("Expected no push (cooldown active), got %d messages", len(messages2))
	}
}

// TestPhase2_Integration_Restart_ExpiredCooldown testet:
// CRIT gespeichert → Neustart → Cooldown vorbei → Push erlaubt
func TestPhase2_Integration_Restart_ExpiredCooldown(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "watcher_state.db")

	// Erster Lauf: CRIT erreichen und Push senden
	store1, err := NewSQLiteStateStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store1: %v", err)
	}
	defer store1.Close()

	fakeHealth1 := &fakeHealthChecker{
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
			{Success: false, Mode: "ping+https", Error: "connection_failed"}, // 10× → CRIT
		},
	}
	counter1 := NewCounter()
	fakeAdapter1 := push.NewFakeAdapter()
	pushService1 := NewPushService(PushServiceConfig{
		Adapter: fakeAdapter1,
		Enabled: true,
	})
	gpio1 := NewNoOpGPIO()

	runner1, err := NewRunner(RunnerConfig{
		Health:        fakeHealth1,
		Counter:       counter1,
		Push:          pushService1,
		GPIO:          gpio1,
		Store:         store1,
		Interval:      30 * time.Millisecond,
		WarnThreshold: 3,
		CritThreshold: 10,
		CooldownSecs:  1, // 1 Sekunde Cooldown (kurz für Test)
	})
	if err != nil {
		t.Fatalf("Failed to create runner1: %v", err)
	}

	ctx1, cancel1 := context.WithTimeout(context.Background(), 400*time.Millisecond)
	defer cancel1()

	done1 := make(chan error, 1)
	go func() {
		done1 <- runner1.Run(ctx1)
	}()

	time.Sleep(350 * time.Millisecond)
	runner1.Stop()
	<-done1
	store1.Close()

	// Warte, bis Cooldown abgelaufen ist
	time.Sleep(2 * time.Second)

	// Zweiter Lauf: Neustart mit abgelaufenem Cooldown
	store2, err := NewSQLiteStateStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store2: %v", err)
	}
	defer store2.Close()

	// Simuliere Recovery → dann wieder CRIT (neue Eskalation)
	fakeHealth2 := &fakeHealthChecker{
		results: []Result{
			{Success: true, Mode: "ping+https"}, // Recovery
			{Success: false, Mode: "ping+https", Error: "connection_failed"},
			{Success: false, Mode: "ping+https", Error: "connection_failed"},
			{Success: false, Mode: "ping+https", Error: "connection_failed"},
			{Success: false, Mode: "ping+https", Error: "connection_failed"},
			{Success: false, Mode: "ping+https", Error: "connection_failed"},
			{Success: false, Mode: "ping+https", Error: "connection_failed"},
			{Success: false, Mode: "ping+https", Error: "connection_failed"},
			{Success: false, Mode: "ping+https", Error: "connection_failed"},
			{Success: false, Mode: "ping+https", Error: "connection_failed"},
			{Success: false, Mode: "ping+https", Error: "connection_failed"}, // 10× → CRIT (neue Eskalation)
		},
	}
	counter2 := NewCounter()
	fakeAdapter2 := push.NewFakeAdapter()
	pushService2 := NewPushService(PushServiceConfig{
		Adapter: fakeAdapter2,
		Enabled: true,
	})
	gpio2 := NewNoOpGPIO()

	runner2, err := NewRunner(RunnerConfig{
		Health:        fakeHealth2,
		Counter:       counter2,
		Push:          pushService2,
		GPIO:          gpio2,
		Store:         store2,
		Interval:      30 * time.Millisecond,
		WarnThreshold: 3,
		CritThreshold: 10,
		CooldownSecs:  1, // 1 Sekunde Cooldown
	})
	if err != nil {
		t.Fatalf("Failed to create runner2: %v", err)
	}

	ctx2, cancel2 := context.WithTimeout(context.Background(), 400*time.Millisecond)
	defer cancel2()

	done2 := make(chan error, 1)
	go func() {
		done2 <- runner2.Run(ctx2)
	}()

	time.Sleep(380 * time.Millisecond)
	runner2.Stop()
	<-done2

	// Prüfe, dass Push gesendet wurde (Cooldown abgelaufen, neue Eskalation)
	messages2 := fakeAdapter2.GetMessages()
	critFound := false
	for _, msg := range messages2 {
		if msg.Severity == rules.SeverityCrit {
			critFound = true
		}
	}
	if !critFound {
		t.Error("Expected CRIT push (cooldown expired, new escalation), got none")
	}
}

// TestPhase2_Integration_StableInfo_NoSave testet:
// INFO stabil → Keine Änderungen → Kein Save
func TestPhase2_Integration_StableInfo_NoSave(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "watcher_state.db")

	store, err := NewSQLiteStateStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Initial State: INFO, FailCount=0
	initialState := PersistedState{
		FailCount:       0,
		CurrentSeverity: rules.SeverityInfo,
		LastEscalation:  time.Unix(0, 0),
	}
	if err := store.Save(initialState); err != nil {
		t.Fatalf("Failed to save initial state: %v", err)
	}

	// Health immer Success (stabil INFO)
	fakeHealth := &fakeHealthChecker{
		results: []Result{
			{Success: true, Mode: "ping+https"},
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
		Store:         store,
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
	runner.Stop()
	<-done

	// Prüfe, dass State unverändert ist (kein Save bei stabiler INFO)
	finalState, err := store.Load()
	if err != nil {
		t.Fatalf("Failed to load final state: %v", err)
	}

	// State sollte unverändert sein (kein Save bei stabiler INFO)
	if finalState.FailCount != initialState.FailCount {
		t.Errorf("Expected FailCount unchanged (%d), got %d", initialState.FailCount, finalState.FailCount)
	}
	if finalState.CurrentSeverity != initialState.CurrentSeverity {
		t.Errorf("Expected Severity unchanged (%v), got %v", initialState.CurrentSeverity, finalState.CurrentSeverity)
	}
	if !finalState.LastEscalation.Equal(initialState.LastEscalation) {
		t.Errorf("Expected LastEscalation unchanged (%v), got %v", initialState.LastEscalation, finalState.LastEscalation)
	}
}

// TestPhase2_Integration_Flapping_WithCooldown testet:
// Eskalation → Cooldown aktiv → Mehrere Failures → Kein weiterer Push
func TestPhase2_Integration_Flapping_WithCooldown(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "watcher_state.db")

	store, err := NewSQLiteStateStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Eskalation zu CRIT
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
			{Success: false, Mode: "ping+https", Error: "connection_failed"}, // 12× → CRIT (kein Push, Cooldown)
			{Success: false, Mode: "ping+https", Error: "connection_failed"}, // 13× → CRIT (kein Push, Cooldown)
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
		CooldownSecs:  2, // 2 Sekunden Cooldown (kurz für Test)
	})
	if err != nil {
		t.Fatalf("Failed to create runner: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- runner.Run(ctx)
	}()

	time.Sleep(450 * time.Millisecond)
	runner.Stop()
	<-done

	// Prüfe, dass nur EIN Push gesendet wurde (bei erster CRIT-Eskalation)
	// Weitere CRIT-Zustände sollten durch Cooldown blockiert sein
	messages := fakeAdapter.GetMessages()
	critCount := 0
	for _, msg := range messages {
		if msg.Severity == rules.SeverityCrit {
			critCount++
		}
	}
	if critCount > 1 {
		t.Errorf("Expected at most 1 CRIT push (cooldown blocks subsequent), got %d", critCount)
	}
	if critCount == 0 {
		t.Error("Expected 1 CRIT push (first escalation), got none")
	}
}

// TestPhase2_Integration_NoPanic_DBError testet:
// DB-Fehler → Kein Panic → Weiterlaufen mit Default-State
func TestPhase2_Integration_NoPanic_DBError(t *testing.T) {
	// Verwendet ungültigen DB-Pfad (sollte Fehler beim Laden geben)
	invalidPath := "/invalid/path/that/does/not/exist/watcher_state.db"

	store, err := NewSQLiteStateStore(invalidPath)
	// NewSQLiteStateStore sollte das Verzeichnis erstellen, also sollte kein Fehler sein
	// Aber wir können einen Store mit geschlossener DB simulieren
	if err != nil {
		// Erwartet: Verzeichnis kann nicht erstellt werden (auf manchen Systemen)
		// In diesem Fall überspringen wir den Test
		t.Skipf("Cannot create store at invalid path (expected on some systems): %v", err)
	}
	defer store.Close()

	// Schließe die DB, um einen Fehler beim Laden zu simulieren
	store.Close()

	// Versuche, einen Runner mit geschlossener DB zu erstellen
	// Dies sollte keinen Panic verursachen
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

	// Runner sollte auch mit geschlossener DB erstellt werden können
	// (Load-Fehler werden ignoriert)
	runner, err := NewRunner(RunnerConfig{
		Health:        fakeHealth,
		Counter:       counter,
		Push:          pushService,
		GPIO:          gpio,
		Store:         store, // Geschlossene DB
		Interval:      50 * time.Millisecond,
		WarnThreshold: 3,
		CritThreshold: 10,
		CooldownSecs:  600,
	})
	if err != nil {
		t.Fatalf("Failed to create runner (should handle DB errors gracefully): %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		// Sollte keinen Panic verursachen, auch wenn DB-Fehler auftreten
		done <- runner.Run(ctx)
	}()

	time.Sleep(80 * time.Millisecond)
	runner.Stop()
	<-done

	// Test erfolgreich, wenn kein Panic aufgetreten ist
}
