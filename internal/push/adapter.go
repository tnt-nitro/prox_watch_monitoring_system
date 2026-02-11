package push

import (
	"context"
	"time"

	"prox-watch/internal/rules"
)

// Adapter is the interface for push notification adapters.
type Adapter interface {
	// Send sends a message to a topic.
	Send(ctx context.Context, topic string, message Message) error

	// GetTopic returns the topic for a severity level.
	GetTopic(severity rules.Severity) string
}

// Message represents a push notification message (metadata only).
type Message struct {
	EventID   string
	Severity  rules.Severity
	Timestamp time.Time
	// No log text
	// No IPs
	// No hostnames
}
