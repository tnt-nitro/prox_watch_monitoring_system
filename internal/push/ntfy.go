package push

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"prox-watch/internal/config"
	"prox-watch/internal/rules"
)

// NtfyAdapter implements the Adapter interface using ntfy.
type NtfyAdapter struct {
	server string
	topics TopicConfig
	client *http.Client
	token  string // Optional authentication token
}

// TopicConfig contains topic names for different severity levels.
type TopicConfig struct {
	Info string
	Warn string
	Crit string
}

// NewNtfyAdapter creates a new ntfy adapter.
func NewNtfyAdapter(cfg *config.Config, token string) *NtfyAdapter {
	server := cfg.Alerts.Ntfy.Server
	if server == "" {
		server = "https://ntfy.sh" // Default ntfy server
	}

	topics := TopicConfig{
		Info: cfg.Alerts.Ntfy.Topics.Info,
		Warn: cfg.Alerts.Ntfy.Topics.Warn,
		Crit: cfg.Alerts.Ntfy.Topics.Crit,
	}

	return &NtfyAdapter{
		server: server,
		topics: topics,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		token: token,
	}
}

// Send sends a message to an ntfy topic.
func (n *NtfyAdapter) Send(ctx context.Context, topic string, message Message) error {
	// Build URL
	url := fmt.Sprintf("%s/%s", n.server, topic)

	// Create message payload (metadata only)
	payload := map[string]interface{}{
		"event_id":  message.EventID,
		"severity":  message.Severity.String(),
		"timestamp": message.Timestamp.Format(time.RFC3339),
		// No log text
		// No IPs
		// No hostnames
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Title", fmt.Sprintf("prox-watch: %s", message.Severity.String()))
	req.Header.Set("Priority", n.getPriority(message.Severity))

	// Add authentication token if provided
	if n.token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", n.token))
	}

	// Send request
	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// GetTopic returns the topic for a severity level.
func (n *NtfyAdapter) GetTopic(severity rules.Severity) string {
	switch severity {
	case rules.SeverityInfo:
		return n.topics.Info
	case rules.SeverityWarn:
		return n.topics.Warn
	case rules.SeverityCrit:
		return n.topics.Crit
	default:
		return n.topics.Info
	}
}

// getPriority returns the ntfy priority level for a severity.
func (n *NtfyAdapter) getPriority(severity rules.Severity) string {
	switch severity {
	case rules.SeverityCrit:
		return "5" // Urgent
	case rules.SeverityWarn:
		return "3" // Default
	case rules.SeverityInfo:
		return "1" // Min
	default:
		return "3"
	}
}

// LocalOnlyAdapter is a no-op adapter for local-only mode.
type LocalOnlyAdapter struct{}

// NewLocalOnlyAdapter creates a local-only adapter (no-op).
func NewLocalOnlyAdapter() *LocalOnlyAdapter {
	return &LocalOnlyAdapter{}
}

// Send does nothing (local-only mode).
func (l *LocalOnlyAdapter) Send(ctx context.Context, topic string, message Message) error {
	// No-op: local-only mode doesn't send push notifications
	return nil
}

// GetTopic returns an empty topic (not used in local-only mode).
func (l *LocalOnlyAdapter) GetTopic(severity rules.Severity) string {
	return ""
}
