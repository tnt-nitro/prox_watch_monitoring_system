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

// TestPhase3_Integration_INFO_To_CRIT_AttemptOnce testet:
// INFO → CRIT → Attempt 1×
func TestPhase3_Integration_INFO_To_CRIT_AttemptOnce(t *testing.T) {
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

	// Health-Checker: Erfolg → Fehler (INFO → CRIT)
	fakeHealth := &fakeHealthChecker{
		results: []Result{
			{Success: true},  // INFO
			{Success: false}, // 1
			{Success: false}, // 2
			{Success: false}, // 3 → WARN
			{Success: false}, // 4
			{Success: false}, // 5
			{Success: false}, // 6
			{Success: false}, // 7
			{Success: false}, // 8
			{Success: false}, // 9
			{Success: false}, // 10 → CRIT (Edge-Trigger)
		},
	}

	counter := NewCounter()
	fakeAdapter := push.NewFakeAdapter()
	pushService := NewPushService(PushServiceConfig{
		Adapter: fakeAdapter,
		Enabled: true,
	})
	gpio := NewNoOpGPIO()

	// Power-Cycle-Konfiguration
	powerCycleCfg := DefaultPowerCycleConfig()
	powerCycleCfg.Enabled = true
	powerCycleCfg.ArmFilePath = armPath
	powerCycleCfg.RequireManualArm = true
	powerCycleCfg.RelayMode = RelayModeCutPowerOnInactive
	powerCycleCfg.MinDowntimeSeconds = 1
	powerCycleCfg.MaxAttempts = 1

	// Erstelle MockPin für Power-Cycle
	mockPin, err := NewMockPin(powerCycleCfg.GPIOPin)
	if err != nil {
		t.Fatalf("Failed to create mock pin: %v", err)
	}

	// Erstelle PowerCycler
	powerCycler, err := NewGPIORelayPowerCycler(powerCycleCfg, store, mockPin)
	if err != nil {
		t.Fatalf("Failed to create power cycler: %v", err)
	}

	runner, err := NewRunner(RunnerConfig{
		Health:        fakeHealth,
		Counter:       counter,
		Push:          pushService,
		GPIO:          gpio,
		PowerCycler:   powerCycler,
		Store:         store,
		Interval:      30 * time.Millisecond,
		WarnThreshold: 3,
		CritThreshold: 10,
		CooldownSecs:  600,
		PowerCycleCfg: powerCycleCfg,
	})
	if err != nil {
		t.Fatalf("Failed to create runner: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond) // 10+ Intervalle
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- runner.Run(ctx)
	}()

	<-done

	// Prüfe, dass genau 1 Attempt durchgeführt wurde
	state, err := store.Load()
	if err != nil {
		t.Fatalf("Failed to load state: %v", err)
	}
	if state.PowerAttempts != 1 {
		t.Errorf("Expected PowerAttempts=1 (only one attempt on INFO→CRIT edge), got %d", state.PowerAttempts)
	}

	// Prüfe, dass ARM-Datei entfernt wurde
	if _, err := os.Stat(armPath); err == nil {
		t.Error("Expected arm file to be removed after attempt, but it still exists")
	}
}

// TestPhase3_Integration_CRIT_Persistent_NoFurtherAttempt testet:
// CRIT dauerhaft → kein weiterer Attempt
func TestPhase3_Integration_CRIT_Persistent_NoFurtherAttempt(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "watcher_state.db")
	armPath := filepath.Join(tmpDir, "arm_powercycle")

	store, err := NewSQLiteStateStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Setze State auf CRIT (ohne Edge-Trigger)
	state := PersistedState{
		FailCount:        10,
		CurrentSeverity:  rules.SeverityCrit,
		LastEscalation:   time.Now(),
		PowerAttempts:    0,
		LastPowerAttempt: time.Unix(0, 0),
	}
	if err := store.Save(state); err != nil {
		t.Fatalf("Failed to save state: %v", err)
	}

	// Health-Checker: Dauerhaft CRIT (kein Edge)
	fakeHealth := &fakeHealthChecker{
		results: []Result{
			{Success: false}, // CRIT (weiterhin)
			{Success: false}, // CRIT (weiterhin)
			{Success: false}, // CRIT (weiterhin)
		},
	}

	counter := NewCounter()
	// Setze Counter auf 10 (CRIT)
	for i := 0; i < 10; i++ {
		counter.Increment()
	}

	fakeAdapter := push.NewFakeAdapter()
	pushService := NewPushService(PushServiceConfig{
		Adapter: fakeAdapter,
		Enabled: true,
	})
	gpio := NewNoOpGPIO()

	powerCycleCfg := DefaultPowerCycleConfig()
	powerCycleCfg.Enabled = true
	powerCycleCfg.ArmFilePath = armPath
	powerCycleCfg.RequireManualArm = true
	powerCycleCfg.RelayMode = RelayModeCutPowerOnInactive
	powerCycleCfg.MinDowntimeSeconds = 1
	powerCycleCfg.MaxAttempts = 1

	mockPin, err := NewMockPin(powerCycleCfg.GPIOPin)
	if err != nil {
		t.Fatalf("Failed to create mock pin: %v", err)
	}

	powerCycler, err := NewGPIORelayPowerCycler(powerCycleCfg, store, mockPin)
	if err != nil {
		t.Fatalf("Failed to create power cycler: %v", err)
	}

	runner, err := NewRunner(RunnerConfig{
		Health:        fakeHealth,
		Counter:       counter,
		Push:          pushService,
		GPIO:          gpio,
		PowerCycler:   powerCycler,
		Store:         store,
		Interval:      30 * time.Millisecond,
		WarnThreshold: 3,
		CritThreshold: 10,
		CooldownSecs:  600,
		PowerCycleCfg: powerCycleCfg,
	})
	if err != nil {
		t.Fatalf("Failed to create runner: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond) // 3+ Intervalle
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- runner.Run(ctx)
	}()

	<-done

	// Prüfe, dass KEIN Attempt durchgeführt wurde (kein Edge-Trigger)
	state, err = store.Load()
	if err != nil {
		t.Fatalf("Failed to load state: %v", err)
	}
	if state.PowerAttempts != 0 {
		t.Errorf("Expected PowerAttempts=0 (no attempt without edge trigger), got %d", state.PowerAttempts)
	}
}

// TestPhase3_Integration_CRIT_WARN_CRIT_NewAttempt testet:
// CRIT → WARN → CRIT → neuer Attempt möglich (wenn ARM + Retry ok)
func TestPhase3_Integration_CRIT_WARN_CRIT_NewAttempt(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "watcher_state.db")
	armPath := filepath.Join(tmpDir, "arm_powercycle")

	store, err := NewSQLiteStateStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Erste Sequenz: INFO → CRIT (erster Attempt)
	// Setze State nach erstem Attempt
	state := PersistedState{
		FailCount:        10,
		CurrentSeverity:  rules.SeverityCrit,
		LastEscalation:   time.Now().Add(-2000 * time.Second), // Vor 2000 Sekunden (Retry abgelaufen)
		PowerAttempts:    1,                                   // Erster Attempt bereits durchgeführt
		LastPowerAttempt: time.Now().Add(-2000 * time.Second), // Vor 2000 Sekunden
	}
	if err := store.Save(state); err != nil {
		t.Fatalf("Failed to save state: %v", err)
	}

	// Health-Checker: CRIT → WARN → CRIT (neue Eskalation)
	fakeHealth := &fakeHealthChecker{
		results: []Result{
			{Success: true},  // WARN (Recovery)
			{Success: false}, // 1
			{Success: false}, // 2
			{Success: false}, // 3 → WARN
			{Success: false}, // 4
			{Success: false}, // 5
			{Success: false}, // 6
			{Success: false}, // 7
			{Success: false}, // 8
			{Success: false}, // 9
			{Success: false}, // 10 → CRIT (neue Eskalation)
		},
	}

	counter := NewCounter()
	fakeAdapter := push.NewFakeAdapter()
	pushService := NewPushService(PushServiceConfig{
		Adapter: fakeAdapter,
		Enabled: true,
	})
	gpio := NewNoOpGPIO()

	powerCycleCfg := DefaultPowerCycleConfig()
	powerCycleCfg.Enabled = true
	powerCycleCfg.ArmFilePath = armPath
	powerCycleCfg.RequireManualArm = true
	powerCycleCfg.RelayMode = RelayModeCutPowerOnInactive
	powerCycleCfg.MinDowntimeSeconds = 1
	powerCycleCfg.MaxAttempts = 2 // Erlaube 2 Attempts
	powerCycleCfg.RetryAfterSeconds = 900 // 15 Minuten (abgelaufen)

	// Erstelle ARM-Datei für zweiten Attempt
	if err := os.WriteFile(armPath, []byte(""), 0600); err != nil {
		t.Fatalf("Failed to create arm file: %v", err)
	}

	mockPin, err := NewMockPin(powerCycleCfg.GPIOPin)
	if err != nil {
		t.Fatalf("Failed to create mock pin: %v", err)
	}

	powerCycler, err := NewGPIORelayPowerCycler(powerCycleCfg, store, mockPin)
	if err != nil {
		t.Fatalf("Failed to create power cycler: %v", err)
	}

	runner, err := NewRunner(RunnerConfig{
		Health:        fakeHealth,
		Counter:       counter,
		Push:          pushService,
		GPIO:          gpio,
		PowerCycler:   powerCycler,
		Store:         store,
		Interval:      30 * time.Millisecond,
		WarnThreshold: 3,
		CritThreshold: 10,
		CooldownSecs:  600,
		PowerCycleCfg: powerCycleCfg,
	})
	if err != nil {
		t.Fatalf("Failed to create runner: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond) // 10+ Intervalle
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- runner.Run(ctx)
	}()

	<-done

	// Prüfe, dass zweiter Attempt durchgeführt wurde (WARN → CRIT Edge)
	state, err = store.Load()
	if err != nil {
		t.Fatalf("Failed to load state: %v", err)
	}
	if state.PowerAttempts != 2 {
		t.Errorf("Expected PowerAttempts=2 (second attempt on WARN→CRIT edge), got %d", state.PowerAttempts)
	}
}

// TestPhase3_Integration_Restart_CRIT_NoAttempt testet:
// Restart mit gespeichertem CRIT → kein Attempt (ohne Edge)
func TestPhase3_Integration_Restart_CRIT_NoAttempt(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "watcher_state.db")
	armPath := filepath.Join(tmpDir, "arm_powercycle")

	// Erster Lauf: CRIT erreicht, aber kein Attempt (kein ARM)
	store1, err := NewSQLiteStateStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store1: %v", err)
	}
	defer store1.Close()

	// Setze State auf CRIT (ohne Attempt)
	state1 := PersistedState{
		FailCount:        10,
		CurrentSeverity:  rules.SeverityCrit,
		LastEscalation:   time.Now(),
		PowerAttempts:    0,
		LastPowerAttempt: time.Unix(0, 0),
	}
	if err := store1.Save(state1); err != nil {
		t.Fatalf("Failed to save state1: %v", err)
	}

	// Zweiter Lauf (Restart): CRIT weiterhin, aber kein Edge-Trigger
	store2, err := NewSQLiteStateStore(dbPath) // Gleiche DB
	if err != nil {
		t.Fatalf("Failed to create store2: %v", err)
	}
	defer store2.Close()

	// Erstelle ARM-Datei (würde normalerweise Attempt erlauben)
	if err := os.WriteFile(armPath, []byte(""), 0600); err != nil {
		t.Fatalf("Failed to create arm file: %v", err)
	}

	fakeHealth := &fakeHealthChecker{
		results: []Result{
			{Success: false}, // CRIT (weiterhin, kein Edge)
		},
	}

	counter := NewCounter()
	// Setze Counter auf 10 (CRIT)
	for i := 0; i < 10; i++ {
		counter.Increment()
	}

	fakeAdapter := push.NewFakeAdapter()
	pushService := NewPushService(PushServiceConfig{
		Adapter: fakeAdapter,
		Enabled: true,
	})
	gpio := NewNoOpGPIO()

	powerCycleCfg := DefaultPowerCycleConfig()
	powerCycleCfg.Enabled = true
	powerCycleCfg.ArmFilePath = armPath
	powerCycleCfg.RequireManualArm = true
	powerCycleCfg.RelayMode = RelayModeCutPowerOnInactive
	powerCycleCfg.MinDowntimeSeconds = 1
	powerCycleCfg.MaxAttempts = 1

	mockPin, err := NewMockPin(powerCycleCfg.GPIOPin)
	if err != nil {
		t.Fatalf("Failed to create mock pin: %v", err)
	}

	powerCycler, err := NewGPIORelayPowerCycler(powerCycleCfg, store2, mockPin)
	if err != nil {
		t.Fatalf("Failed to create power cycler: %v", err)
	}

	runner, err := NewRunner(RunnerConfig{
		Health:        fakeHealth,
		Counter:       counter,
		Push:          pushService,
		GPIO:          gpio,
		PowerCycler:   powerCycler,
		Store:         store2,
		Interval:      30 * time.Millisecond,
		WarnThreshold: 3,
		CritThreshold: 10,
		CooldownSecs:  600,
		PowerCycleCfg: powerCycleCfg,
	})
	if err != nil {
		t.Fatalf("Failed to create runner: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond) // Ein Intervall
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- runner.Run(ctx)
	}()

	<-done

	// Prüfe, dass KEIN Attempt durchgeführt wurde (kein Edge-Trigger)
	state2, err := store2.Load()
	if err != nil {
		t.Fatalf("Failed to load state2: %v", err)
	}
	if state2.PowerAttempts != 0 {
		t.Errorf("Expected PowerAttempts=0 (no attempt without edge trigger on restart), got %d", state2.PowerAttempts)
	}

	// Prüfe, dass ARM-Datei NICHT entfernt wurde (kein Attempt)
	if _, err := os.Stat(armPath); err != nil {
		t.Error("Expected arm file to remain (no attempt), but it was removed")
	}
}
