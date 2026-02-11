package push

import (
	"context"
	"testing"
	"time"

	"prox-watch/internal/config"
	"prox-watch/internal/rules"
)

func TestNtfyAdapter_GetTopic(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Alerts.Channel = "ntfy"
	cfg.Alerts.Ntfy.Server = "https://ntfy.sh"
	cfg.Alerts.Ntfy.Topics.Info = "prox-watch-info"
	cfg.Alerts.Ntfy.Topics.Warn = "prox-watch-warn"
	cfg.Alerts.Ntfy.Topics.Crit = "prox-watch-crit"

	adapter := NewNtfyAdapter(cfg, "")

	if adapter.GetTopic(rules.SeverityInfo) != "prox-watch-info" {
		t.Errorf("GetTopic(INFO) = %s, want prox-watch-info", adapter.GetTopic(rules.SeverityInfo))
	}
	if adapter.GetTopic(rules.SeverityWarn) != "prox-watch-warn" {
		t.Errorf("GetTopic(WARN) = %s, want prox-watch-warn", adapter.GetTopic(rules.SeverityWarn))
	}
	if adapter.GetTopic(rules.SeverityCrit) != "prox-watch-crit" {
		t.Errorf("GetTopic(CRIT) = %s, want prox-watch-crit", adapter.GetTopic(rules.SeverityCrit))
	}
}

func TestLocalOnlyAdapter_Send(t *testing.T) {
	adapter := NewLocalOnlyAdapter()

	message := Message{
		EventID:   "test.event",
		Severity:  rules.SeverityCrit,
		Timestamp: time.Now(),
	}

	ctx := context.Background()
	if err := adapter.Send(ctx, "test-topic", message); err != nil {
		t.Errorf("Send() error = %v, want nil", err)
	}
}

func TestFakeAdapter_Send(t *testing.T) {
	adapter := NewFakeAdapter()

	message := Message{
		EventID:   "test.event",
		Severity:  rules.SeverityCrit,
		Timestamp: time.Now(),
	}

	ctx := context.Background()
	if err := adapter.Send(ctx, "test-topic", message); err != nil {
		t.Errorf("Send() error = %v, want nil", err)
	}

	messages := adapter.GetMessages()
	if len(messages) != 1 {
		t.Errorf("GetMessages() len = %d, want 1", len(messages))
	}

	if messages[0].EventID != "test.event" {
		t.Errorf("GetMessages()[0].EventID = %s, want test.event", messages[0].EventID)
	}

	adapter.Clear()
	if len(adapter.GetMessages()) != 0 {
		t.Error("Clear() did not clear messages")
	}
}

func TestFakeAdapter_GetTopic(t *testing.T) {
	adapter := NewFakeAdapter()

	if adapter.GetTopic(rules.SeverityInfo) != "prox-watch-info" {
		t.Errorf("GetTopic(INFO) = %s, want prox-watch-info", adapter.GetTopic(rules.SeverityInfo))
	}
}
