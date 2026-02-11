package push

import (
	"context"
	"sync"

	"prox-watch/internal/rules"
)

// FakeAdapter is a fake adapter for testing.
type FakeAdapter struct {
	messages []Message
	mu       sync.Mutex
	topics   map[rules.Severity]string
}

// NewFakeAdapter creates a new fake adapter for testing.
func NewFakeAdapter() *FakeAdapter {
	return &FakeAdapter{
		messages: make([]Message, 0),
		topics: map[rules.Severity]string{
			rules.SeverityInfo: "prox-watch-info",
			rules.SeverityWarn: "prox-watch-warn",
			rules.SeverityCrit: "prox-watch-crit",
		},
	}
}

// Send records a message (for testing).
func (f *FakeAdapter) Send(ctx context.Context, topic string, message Message) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.messages = append(f.messages, message)
	return nil
}

// GetTopic returns the topic for a severity level.
func (f *FakeAdapter) GetTopic(severity rules.Severity) string {
	return f.topics[severity]
}

// GetMessages returns all sent messages (for testing).
func (f *FakeAdapter) GetMessages() []Message {
	f.mu.Lock()
	defer f.mu.Unlock()

	messages := make([]Message, len(f.messages))
	copy(messages, f.messages)
	return messages
}

// Clear clears all recorded messages (for testing).
func (f *FakeAdapter) Clear() {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.messages = f.messages[:0]
}
