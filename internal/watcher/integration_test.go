package watcher

import (
	"context"
	"prox-watch/internal/push"
	"prox-watch/internal/rules"
	"testing"
	"time"
)

// fakeHealthChecker ist ein Fake-HealthChecker für Integration-Tests.
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

// TestIntegration_Scenario1_StableOK testet das Szenario: Stabil OK
// Health immer Success, kein Push, Counter bleibt 0
func TestIntegration_Scenario1_StableOK(t *testing.T) {
	// Health immer Success
	fakeHealth := &fakeHealthChecker{
		results: []Result{
			{Success: true, Mode: "ping+https"},
			{Success: true, Mode: "ping+https"},
			{Success: true, Mode: "ping+https"},
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

	// Prüfungen
	if counter.GetCount() != 0 {
		t.Errorf("Expected counter to be 0 (stable OK), got %d", counter.GetCount())
	}

	messages := fakeAdapter.GetMessages()
	if len(messages) != 0 {
		t.Errorf("Expected no push messages (stable OK), got %d", len(messages))
	}

	runner.Stop()
	<-done
}

// TestIntegration_Scenario2_WARN testet das Szenario: WARN
// 3 Failures, Eskalation → WARN, 1 Push
func TestIntegration_Scenario2_WARN(t *testing.T) {
	// 3 Failures → WARN
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

	// Prüfungen
	if counter.GetCount() != 3 {
		t.Errorf("Expected counter to be 3, got %d", counter.GetCount())
	}

	messages := fakeAdapter.GetMessages()
	if len(messages) != 1 {
		t.Fatalf("Expected 1 push message (WARN escalation), got %d", len(messages))
	}

	msg := messages[0]
	if msg.Severity != rules.SeverityWarn {
		t.Errorf("Expected WARN severity, got %v", msg.Severity)
	}
	if msg.EventID != "external.availability.down" {
		t.Errorf("Expected EventID 'external.availability.down', got '%s'", msg.EventID)
	}

	runner.Stop()
	<-done
}

// TestIntegration_Scenario3_CRIT testet das Szenario: CRIT
// 10 Failures, Eskalation → CRIT, Push + Beep
func TestIntegration_Scenario3_CRIT(t *testing.T) {
	// 10 Failures → CRIT
	results := make([]Result, 10)
	for i := 0; i < 10; i++ {
		results[i] = Result{Success: false, Mode: "ping+https", Error: "connection_failed"}
	}

	fakeHealth := &fakeHealthChecker{
		results: results,
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
		Interval:      30 * time.Millisecond,
		WarnThreshold: 3,
		CritThreshold: 10,
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

	// Warten auf 10 Intervalle
	time.Sleep(350 * time.Millisecond)

	// Prüfungen
	if counter.GetCount() != 10 {
		t.Errorf("Expected counter to be 10, got %d", counter.GetCount())
	}

	messages := fakeAdapter.GetMessages()
	// Erwartung: 2 Pushes (INFO→WARN bei 3, WARN→CRIT bei 10)
	if len(messages) < 2 {
		t.Errorf("Expected at least 2 push messages (WARN + CRIT escalation), got %d", len(messages))
	}

	// Prüfe, dass CRIT-Push vorhanden ist
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

	// Beep wird bei CRIT-Eskalation aufgerufen (GPIO ist NoOp, aber Logik wird getestet)
	// (Keine explizite Prüfung möglich, da NoOpGPIO keine State hat)

	runner.Stop()
	<-done
}

// TestIntegration_Scenario4_Recovery testet das Szenario: Recovery
// CRIT → Success, kein Push, LED zurück INFO
func TestIntegration_Scenario4_Recovery(t *testing.T) {
	// 10 Failures → CRIT, dann Recovery
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
			{Success: false, Mode: "ping+https", Error: "connection_failed"}, // 10× → CRIT
			{Success: true, Mode: "ping+https"}, // Recovery → INFO
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
		Interval:      30 * time.Millisecond,
		WarnThreshold: 3,
		CritThreshold: 10,
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

	// Warten auf 11 Intervalle
	time.Sleep(380 * time.Millisecond)

	// Prüfungen
	// Counter sollte zurückgesetzt sein (nach Recovery)
	if counter.GetCount() != 0 {
		t.Errorf("Expected counter to be 0 after recovery, got %d", counter.GetCount())
	}

	messages := fakeAdapter.GetMessages()
	// Erwartung: 2 Pushes (INFO→WARN bei 3, WARN→CRIT bei 10)
	// Kein Push bei Recovery (CRIT→INFO)
	if len(messages) != 2 {
		t.Errorf("Expected 2 push messages (WARN + CRIT escalation, no push on recovery), got %d", len(messages))
	}

	// Prüfe, dass kein Push bei Recovery gesendet wurde
	// (Alle Messages sollten WARN oder CRIT sein, kein INFO)
	for _, msg := range messages {
		if msg.Severity == rules.SeverityInfo {
			t.Error("Expected no INFO push message (recovery should not trigger push)")
		}
	}

	runner.Stop()
	<-done
}

// TestIntegration_Scenario5_Flapping testet das Szenario: Flattern
// Fail, Success, Fail - kein falsches Push-Spam
func TestIntegration_Scenario5_Flapping(t *testing.T) {
	// Flattern: Fail, Success, Fail, Success, Fail
	fakeHealth := &fakeHealthChecker{
		results: []Result{
			{Success: false, Mode: "ping+https", Error: "connection_failed"},
			{Success: true, Mode: "ping+https"},
			{Success: false, Mode: "ping+https", Error: "connection_failed"},
			{Success: true, Mode: "ping+https"},
			{Success: false, Mode: "ping+https", Error: "connection_failed"},
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

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- runner.Run(ctx)
	}()

	// Warten auf 5 Intervalle
	time.Sleep(280 * time.Millisecond)

	// Prüfungen
	// Counter sollte am Ende 1 sein (letzter Failure, vorherige wurden zurückgesetzt)
	if counter.GetCount() != 1 {
		t.Errorf("Expected counter to be 1 (last failure after flapping), got %d", counter.GetCount())
	}

	messages := fakeAdapter.GetMessages()
	// Erwartung: Kein Push (keine Eskalation, da Counter immer wieder zurückgesetzt wird)
	if len(messages) != 0 {
		t.Errorf("Expected no push messages (flapping should not trigger push), got %d", len(messages))
	}

	runner.Stop()
	<-done
}

// TestIntegration_Determinism testet Determinismus
// Gleiche Inputs → gleiche Outputs
func TestIntegration_Determinism(t *testing.T) {
	// Gleiche Health-Check-Ergebnisse
	results := []Result{
		{Success: false, Mode: "ping+https", Error: "connection_failed"},
		{Success: false, Mode: "ping+https", Error: "connection_failed"},
		{Success: false, Mode: "ping+https", Error: "connection_failed"}, // 3× → WARN
	}

	// Erster Lauf
	fakeHealth1 := &fakeHealthChecker{
		results: results,
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
		Interval:      50 * time.Millisecond,
		WarnThreshold: 3,
		CritThreshold: 10,
	})
	if err != nil {
		t.Fatalf("Failed to create runner1: %v", err)
	}

	ctx1, cancel1 := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel1()

	done1 := make(chan error, 1)
	go func() {
		done1 <- runner1.Run(ctx1)
	}()

	time.Sleep(180 * time.Millisecond)
	runner1.Stop()
	<-done1

	// Zweiter Lauf (gleiche Inputs)
	fakeHealth2 := &fakeHealthChecker{
		results: results,
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
		Interval:      50 * time.Millisecond,
		WarnThreshold: 3,
		CritThreshold: 10,
	})
	if err != nil {
		t.Fatalf("Failed to create runner2: %v", err)
	}

	ctx2, cancel2 := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel2()

	done2 := make(chan error, 1)
	go func() {
		done2 <- runner2.Run(ctx2)
	}()

	time.Sleep(180 * time.Millisecond)
	runner2.Stop()
	<-done2

	// Prüfe Determinismus
	if counter1.GetCount() != counter2.GetCount() {
		t.Errorf("Expected deterministic counter: %d == %d", counter1.GetCount(), counter2.GetCount())
	}

	messages1 := fakeAdapter1.GetMessages()
	messages2 := fakeAdapter2.GetMessages()
	if len(messages1) != len(messages2) {
		t.Errorf("Expected deterministic push count: %d == %d", len(messages1), len(messages2))
	}

	if len(messages1) > 0 && len(messages2) > 0 {
		if messages1[0].Severity != messages2[0].Severity {
			t.Errorf("Expected deterministic severity: %v == %v", messages1[0].Severity, messages2[0].Severity)
		}
	}
}

// TestIntegration_NoHostInOutput testet, dass keine Host/IP im Testoutput sind
func TestIntegration_NoHostInOutput(t *testing.T) {
	fakeHealth := &fakeHealthChecker{
		results: []Result{
			{Success: false, Mode: "ping+https", Error: "connection_failed"},
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

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- runner.Run(ctx)
	}()

	time.Sleep(80 * time.Millisecond)
	runner.Stop()
	<-done

	// Prüfe, dass keine Host/IP in Messages sind
	messages := fakeAdapter.GetMessages()
	for _, msg := range messages {
		// EventID sollte keine IP-ähnlichen Muster enthalten
		if containsIPPattern(msg.EventID) {
			t.Errorf("EventID should not contain IP-like patterns: %s", msg.EventID)
		}
	}
}

// containsIPPattern ist eine einfache Heuristik, um IP-ähnliche Muster zu erkennen.
func containsIPPattern(s string) bool {
	// Sehr einfache Prüfung: Enthält Zahlen mit Punkten (wie IPs)
	hasDigits := false
	hasDots := false
	for _, r := range s {
		if r >= '0' && r <= '9' {
			hasDigits = true
		}
		if r == '.' {
			hasDots = true
		}
	}
	return hasDigits && hasDots
}
