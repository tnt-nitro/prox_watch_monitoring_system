package watcher

import (
	"context"
	"prox-watch/internal/push"
	"prox-watch/internal/rules"
	"testing"
	"time"
)

func TestPushService_SendIfEscalation_InfoNoPush(t *testing.T) {
	// INFO → kein Push
	fakeAdapter := push.NewFakeAdapter()
	service := NewPushService(PushServiceConfig{
		Adapter: fakeAdapter,
		Enabled: true,
	})

	ctx := context.Background()
	err := service.SendIfEscalation(ctx, rules.SeverityInfo, rules.SeverityInfo)
	if err != nil {
		t.Fatalf("SendIfEscalation() failed: %v", err)
	}

	// Kein Push sollte gesendet worden sein
	messages := fakeAdapter.GetMessages()
	if len(messages) != 0 {
		t.Errorf("Expected no push for INFO→INFO, got %d messages", len(messages))
	}
}

func TestPushService_SendIfEscalation_InfoToWarn(t *testing.T) {
	// INFO→WARN → Push
	fakeAdapter := push.NewFakeAdapter()
	service := NewPushService(PushServiceConfig{
		Adapter: fakeAdapter,
		Enabled: true,
	})

	ctx := context.Background()
	err := service.SendIfEscalation(ctx, rules.SeverityInfo, rules.SeverityWarn)
	if err != nil {
		t.Fatalf("SendIfEscalation() failed: %v", err)
	}

	// Push sollte gesendet worden sein
	messages := fakeAdapter.GetMessages()
	if len(messages) != 1 {
		t.Fatalf("Expected 1 push for INFO→WARN, got %d messages", len(messages))
	}

	msg := messages[0]
	if msg.EventID != "external.availability.down" {
		t.Errorf("Expected EventID 'external.availability.down', got '%s'", msg.EventID)
	}
	if msg.Severity != rules.SeverityWarn {
		t.Errorf("Expected Severity WARN, got %v", msg.Severity)
	}
}

func TestPushService_SendIfEscalation_WarnToCrit(t *testing.T) {
	// WARN→CRIT → Push
	fakeAdapter := push.NewFakeAdapter()
	service := NewPushService(PushServiceConfig{
		Adapter: fakeAdapter,
		Enabled: true,
	})

	ctx := context.Background()
	err := service.SendIfEscalation(ctx, rules.SeverityWarn, rules.SeverityCrit)
	if err != nil {
		t.Fatalf("SendIfEscalation() failed: %v", err)
	}

	// Push sollte gesendet worden sein
	messages := fakeAdapter.GetMessages()
	if len(messages) != 1 {
		t.Fatalf("Expected 1 push for WARN→CRIT, got %d messages", len(messages))
	}

	msg := messages[0]
	if msg.EventID != "external.availability.down" {
		t.Errorf("Expected EventID 'external.availability.down', got '%s'", msg.EventID)
	}
	if msg.Severity != rules.SeverityCrit {
		t.Errorf("Expected Severity CRIT, got %v", msg.Severity)
	}
}

func TestPushService_SendIfEscalation_WarnToWarn(t *testing.T) {
	// WARN→WARN → kein Push (kein Statuswechsel)
	fakeAdapter := push.NewFakeAdapter()
	service := NewPushService(PushServiceConfig{
		Adapter: fakeAdapter,
		Enabled: true,
	})

	ctx := context.Background()
	err := service.SendIfEscalation(ctx, rules.SeverityWarn, rules.SeverityWarn)
	if err != nil {
		t.Fatalf("SendIfEscalation() failed: %v", err)
	}

	// Kein Push sollte gesendet worden sein
	messages := fakeAdapter.GetMessages()
	if len(messages) != 0 {
		t.Errorf("Expected no push for WARN→WARN, got %d messages", len(messages))
	}
}

func TestPushService_SendIfEscalation_CritToInfo(t *testing.T) {
	// CRIT→INFO → kein Push (Verbesserung)
	fakeAdapter := push.NewFakeAdapter()
	service := NewPushService(PushServiceConfig{
		Adapter: fakeAdapter,
		Enabled: true,
	})

	ctx := context.Background()
	err := service.SendIfEscalation(ctx, rules.SeverityCrit, rules.SeverityInfo)
	if err != nil {
		t.Fatalf("SendIfEscalation() failed: %v", err)
	}

	// Kein Push sollte gesendet worden sein
	messages := fakeAdapter.GetMessages()
	if len(messages) != 0 {
		t.Errorf("Expected no push for CRIT→INFO (improvement), got %d messages", len(messages))
	}
}

func TestPushService_SendIfEscalation_Disabled(t *testing.T) {
	// Push deaktiviert → kein Push
	fakeAdapter := push.NewFakeAdapter()
	service := NewPushService(PushServiceConfig{
		Adapter: fakeAdapter,
		Enabled: false,
	})

	ctx := context.Background()
	err := service.SendIfEscalation(ctx, rules.SeverityInfo, rules.SeverityWarn)
	if err != nil {
		t.Fatalf("SendIfEscalation() failed: %v", err)
	}

	// Kein Push sollte gesendet worden sein
	messages := fakeAdapter.GetMessages()
	if len(messages) != 0 {
		t.Errorf("Expected no push when disabled, got %d messages", len(messages))
	}
}

func TestPushService_SendIfEscalation_EventID(t *testing.T) {
	// Prüfe, dass Event-ID korrekt ist
	fakeAdapter := push.NewFakeAdapter()
	service := NewPushService(PushServiceConfig{
		Adapter: fakeAdapter,
		Enabled: true,
	})

	ctx := context.Background()
	err := service.SendIfEscalation(ctx, rules.SeverityInfo, rules.SeverityWarn)
	if err != nil {
		t.Fatalf("SendIfEscalation() failed: %v", err)
	}

	messages := fakeAdapter.GetMessages()
	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}

	msg := messages[0]
	if msg.EventID != "external.availability.down" {
		t.Errorf("Expected EventID 'external.availability.down', got '%s'", msg.EventID)
	}
}

func TestPushService_SendIfEscalation_Timestamp(t *testing.T) {
	// Prüfe, dass Timestamp gesetzt ist
	fakeAdapter := push.NewFakeAdapter()
	service := NewPushService(PushServiceConfig{
		Adapter: fakeAdapter,
		Enabled: true,
	})

	before := time.Now().UTC()
	ctx := context.Background()
	err := service.SendIfEscalation(ctx, rules.SeverityInfo, rules.SeverityWarn)
	after := time.Now().UTC()

	if err != nil {
		t.Fatalf("SendIfEscalation() failed: %v", err)
	}

	messages := fakeAdapter.GetMessages()
	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}

	msg := messages[0]
	if msg.At.Before(before) || msg.At.After(after) {
		t.Errorf("Timestamp should be between %v and %v, got %v", before, after, msg.At)
	}
}

func TestPushService_SendIfEscalation_NoHostInMessage(t *testing.T) {
	// Prüfe, dass keine Host-Informationen in der Message sind
	fakeAdapter := push.NewFakeAdapter()
	service := NewPushService(PushServiceConfig{
		Adapter: fakeAdapter,
		Enabled: true,
	})

	ctx := context.Background()
	err := service.SendIfEscalation(ctx, rules.SeverityInfo, rules.SeverityWarn)
	if err != nil {
		t.Fatalf("SendIfEscalation() failed: %v", err)
	}

	messages := fakeAdapter.GetMessages()
	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}

	msg := messages[0]
	// EventID sollte keine Host-Informationen enthalten
	if msg.EventID == "" || len(msg.EventID) > 100 {
		t.Errorf("EventID should be set and reasonable length, got '%s'", msg.EventID)
	}
	// Prüfe, dass EventID keine IP/Host-ähnlichen Muster enthält
	// (einfache Heuristik)
	if containsIPPattern(msg.EventID) {
		t.Errorf("EventID should not contain IP-like patterns, got '%s'", msg.EventID)
	}
}

// containsIPPattern ist eine einfache Heuristik, um IP-ähnliche Muster zu erkennen.
func containsIPPattern(s string) bool {
	// Sehr einfache Prüfung: Enthält Zahlen mit Punkten (wie IPs)
	// Dies ist keine vollständige Validierung, nur eine Heuristik für Tests
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

func TestPushService_IsEnabled(t *testing.T) {
	// Prüfe IsEnabled()
	serviceEnabled := NewPushService(PushServiceConfig{
		Adapter: push.NewFakeAdapter(),
		Enabled: true,
	})
	if !serviceEnabled.IsEnabled() {
		t.Error("Expected IsEnabled() to return true")
	}

	serviceDisabled := NewPushService(PushServiceConfig{
		Adapter: push.NewFakeAdapter(),
		Enabled: false,
	})
	if serviceDisabled.IsEnabled() {
		t.Error("Expected IsEnabled() to return false")
	}
}
