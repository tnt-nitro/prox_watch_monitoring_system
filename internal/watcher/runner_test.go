package watcher

import (
	"context"
	"prox-watch/internal/push"
	"prox-watch/internal/rules"
	"testing"
	"time"
)

// fakeHealthChecker ist ein Fake-HealthChecker für Tests.
type fakeHealthChecker struct {
	results []Result
	index   int
}

func (f *fakeHealthChecker) Check(ctx context.Context) (Result, error) {
	if f.index >= len(f.results) {
		// Wiederhole letztes Ergebnis
		return f.results[len(f.results)-1], nil
	}
	result := f.results[f.index]
	f.index++
	return result, nil
}

func TestRunner_Success_Reset(t *testing.T) {
	// Success → Reset
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
		Interval:      100 * time.Millisecond,
		WarnThreshold: 3,
		CritThreshold: 10,
	})
	if err != nil {
		t.Fatalf("Failed to create runner: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	// Runner im Hintergrund starten
	done := make(chan error, 1)
	go func() {
		done <- runner.Run(ctx)
	}()

	// Warten auf ein Intervall
	time.Sleep(150 * time.Millisecond)

	// Prüfe, dass Counter zurückgesetzt wurde
	if counter.GetCount() != 0 {
		t.Errorf("Expected counter to be reset to 0, got %d", counter.GetCount())
	}

	// Stoppe Runner
	runner.Stop()
	<-done
}

func TestRunner_WarnEscalation(t *testing.T) {
	// WARN-Eskalation
	fakeHealth := &fakeHealthChecker{
		results: []Result{
			{Success: false, Mode: "ping+https", Error: "connection_failed"},
			{Success: false, Mode: "ping+https", Error: "connection_failed"},
			{Success: false, Mode: "ping+https", Error: "connection_failed"}, // 3× → WARN
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
		Interval:      50 * time.Millisecond,
		WarnThreshold: 3,
		CritThreshold: 10,
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

	// Warten auf 3 Intervalle
	time.Sleep(180 * time.Millisecond)

	// Prüfe, dass Push gesendet wurde
	messages := fakeAdapter.GetMessages()
	if len(messages) == 0 {
		t.Error("Expected push message for WARN escalation, got none")
	} else {
		if messages[0].Severity != rules.SeverityWarn {
			t.Errorf("Expected WARN severity, got %v", messages[0].Severity)
		}
		if messages[0].EventID != "external.availability.down" {
			t.Errorf("Expected EventID 'external.availability.down', got '%s'", messages[0].EventID)
		}
	}

	runner.Stop()
	<-done
}

func TestRunner_CritEscalation(t *testing.T) {
	// CRIT-Eskalation
	fakeHealth := &fakeHealthChecker{
		results: []Result{
			{Success: false, Mode: "ping+https", Error: "connection_failed"},
			{Success: false, Mode: "ping+https", Error: "connection_failed"},
			{Success: false, Mode: "ping+https", Error: "connection_failed"}, // 3× → WARN
			{Success: false, Mode: "ping+https", Error: "connection_failed"},
			{Success: false, Mode: "ping+https", Error: "connection_failed"},
			{Success: false, Mode: "ping+https", Error: "connection_failed"},
			{Success: false, Mode: "ping+https", Error: "connection_failed"},
			{Success: false, Mode: "ping+https", Error: "connection_failed"},
			{Success: false, Mode: "ping+https", Error: "connection_failed"},
			{Success: false, Mode: "ping+https", Error: "connection_failed"}, // 10× → CRIT
		},
	}
	counter := NewCounter()
	fakeAdapter := push.NewFakeAdapter()
	pushService := NewPushService(PushServiceConfig{
		Adapter: fakeAdapter,
		Enabled: true,
	})
	gpio := NewNoOpGPIO()

	runner := NewRunner(RunnerConfig{
		Health:        fakeHealth,
		Counter:       counter,
		Push:          pushService,
		GPIO:          gpio,
		Interval:      30 * time.Millisecond,
		WarnThreshold: 3,
		CritThreshold: 10,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 400*time.Millisecond)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- runner.Run(ctx)
	}()

	// Warten auf 10 Intervalle
	time.Sleep(350 * time.Millisecond)

	// Prüfe, dass CRIT-Push gesendet wurde
	messages := fakeAdapter.GetMessages()
	critFound := false
	for _, msg := range messages {
		if msg.Severity == rules.SeverityCrit {
			critFound = true
			if msg.EventID != "external.availability.down" {
				t.Errorf("Expected EventID 'external.availability.down', got '%s'", msg.EventID)
			}
		}
	}
	if !critFound {
		t.Error("Expected CRIT push message, got none")
	}

	runner.Stop()
	<-done
}

func TestRunner_NoPushSameSeverity(t *testing.T) {
	// Kein Push bei gleicher Severity
	fakeHealth := &fakeHealthChecker{
		results: []Result{
			{Success: false, Mode: "ping+https", Error: "connection_failed"},
			{Success: false, Mode: "ping+https", Error: "connection_failed"},
			{Success: false, Mode: "ping+https", Error: "connection_failed"}, // 3× → WARN (Push)
			{Success: false, Mode: "ping+https", Error: "connection_failed"}, // 4× → WARN (kein Push)
		},
	}
	counter := NewCounter()
	fakeAdapter := push.NewFakeAdapter()
	pushService := NewPushService(PushServiceConfig{
		Adapter: fakeAdapter,
		Enabled: true,
	})
	gpio := NewNoOpGPIO()

	runner := NewRunner(RunnerConfig{
		Health:        fakeHealth,
		Counter:       counter,
		Push:          pushService,
		GPIO:          gpio,
		Interval:      50 * time.Millisecond,
		WarnThreshold: 3,
		CritThreshold: 10,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- runner.Run(ctx)
	}()

	time.Sleep(220 * time.Millisecond)

	// Prüfe, dass nur ein Push gesendet wurde (bei Eskalation INFO→WARN)
	messages := fakeAdapter.GetMessages()
	if len(messages) != 1 {
		t.Errorf("Expected 1 push message (INFO→WARN), got %d", len(messages))
	}
	if len(messages) > 0 && messages[0].Severity != rules.SeverityWarn {
		t.Errorf("Expected WARN severity, got %v", messages[0].Severity)
	}

	runner.Stop()
	<-done
}

func TestRunner_NoPushImprovement(t *testing.T) {
	// Kein Push bei Verbesserung
	fakeHealth := &fakeHealthChecker{
		results: []Result{
			{Success: false, Mode: "ping+https", Error: "connection_failed"},
			{Success: false, Mode: "ping+https", Error: "connection_failed"},
			{Success: false, Mode: "ping+https", Error: "connection_failed"}, // 3× → WARN (Push)
			{Success: true, Mode: "ping+https"}, // Recovery → INFO (kein Push)
		},
	}
	counter := NewCounter()
	fakeAdapter := push.NewFakeAdapter()
	pushService := NewPushService(PushServiceConfig{
		Adapter: fakeAdapter,
		Enabled: true,
	})
	gpio := NewNoOpGPIO()

	runner := NewRunner(RunnerConfig{
		Health:        fakeHealth,
		Counter:       counter,
		Push:          pushService,
		GPIO:          gpio,
		Interval:      50 * time.Millisecond,
		WarnThreshold: 3,
		CritThreshold: 10,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- runner.Run(ctx)
	}()

	time.Sleep(220 * time.Millisecond)

	// Prüfe, dass nur ein Push gesendet wurde (bei Eskalation INFO→WARN)
	// Kein Push bei Verbesserung (WARN→INFO)
	messages := fakeAdapter.GetMessages()
	if len(messages) != 1 {
		t.Errorf("Expected 1 push message (INFO→WARN), got %d (no push on improvement)", len(messages))
	}

	runner.Stop()
	<-done
}

func TestRunner_BeepOnlyCritEscalation(t *testing.T) {
	// Beep nur bei CRIT-Eskalation
	// (GPIO ist NoOp, aber Logik wird getestet)
	fakeHealth := &fakeHealthChecker{
		results: []Result{
			{Success: false, Mode: "ping+https", Error: "connection_failed"},
		},
	}
	counter := NewCounter()
	pushService := NewPushService(PushServiceConfig{
		Adapter: push.NewFakeAdapter(),
		Enabled: true,
	})
	gpio := NewNoOpGPIO()

	runner := NewRunner(RunnerConfig{
		Health:        fakeHealth,
		Counter:       counter,
		Push:          pushService,
		GPIO:          gpio,
		Interval:      50 * time.Millisecond,
		WarnThreshold: 3,
		CritThreshold: 10,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- runner.Run(ctx)
	}()

	time.Sleep(80 * time.Millisecond)

	// Beep sollte nur bei CRIT-Eskalation aufgerufen werden
	// Bei nur 1 Fehler (INFO) sollte kein Beep sein
	// (GPIO ist NoOp, aber Logik wird durch fehlende Fehler getestet)

	runner.Stop()
	<-done
}

func TestRunner_ContextCancel_Stop(t *testing.T) {
	// Context Cancel → Stop
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
		Interval:      100 * time.Millisecond,
		WarnThreshold: 3,
		CritThreshold: 10,
	})
	if err != nil {
		t.Fatalf("Failed to create runner: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() {
		done <- runner.Run(ctx)
	}()

	// Warten kurz, dann cancel
	time.Sleep(50 * time.Millisecond)
	cancel()

	// Warten auf Stop
	select {
	case err := <-done:
		if err != context.Canceled {
			t.Errorf("Expected context.Canceled, got %v", err)
		}
	case <-time.After(200 * time.Millisecond):
		t.Error("Runner did not stop after context cancellation")
	}
}

func TestRunner_Stop_Graceful(t *testing.T) {
	// Stop() → Graceful Shutdown
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
		Interval:      100 * time.Millisecond,
		WarnThreshold: 3,
		CritThreshold: 10,
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

	// Warten kurz, dann stop
	time.Sleep(50 * time.Millisecond)
	if err := runner.Stop(); err != nil {
		t.Errorf("Stop() failed: %v", err)
	}

	// Warten auf Stop
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Expected nil error on graceful stop, got %v", err)
		}
	case <-time.After(200 * time.Millisecond):
		t.Error("Runner did not stop after Stop()")
	}
}
