package pattern

import (
	"context"
	"fmt"
	"strings"

	"prox-watch/internal/journal"
	"prox-watch/internal/rules"
)

// Matcher is the interface for pattern matching.
type Matcher interface {
	// Match checks an entry against patterns and returns a match result if found.
	Match(ctx context.Context, entry journal.Entry) (*MatchResult, error)

	// LoadPatterns loads pattern definitions from a file.
	LoadPatterns(path string) error
}

// PatternMatcher implements the Matcher interface.
type PatternMatcher struct {
	registry *Registry
}

// NewPatternMatcher creates a new pattern matcher.
func NewPatternMatcher() *PatternMatcher {
	return &PatternMatcher{
		registry: NewRegistry(),
	}
}

// LoadPatterns loads patterns from a file.
func (m *PatternMatcher) LoadPatterns(path string) error {
	return m.registry.LoadPatterns(path)
}

// Match checks an entry against all patterns and returns the first match.
func (m *PatternMatcher) Match(ctx context.Context, entry journal.Entry) (*MatchResult, error) {
	patterns := m.registry.GetAllPatterns()

	for _, pattern := range patterns {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Check if pattern matches entry
		if m.matches(pattern, entry) {
			eventID := m.generateEventID(pattern, entry)
			return &MatchResult{
				EventID:   eventID,
				Severity:  pattern.Severity,
				PatternID: pattern.PatternID,
			}, nil
		}
	}

	// No match found
	return nil, nil
}

// matches checks if a pattern matches an entry.
func (m *PatternMatcher) matches(pattern Pattern, entry journal.Entry) bool {
	// Check source match
	if pattern.Source != "" && pattern.Source != entry.Source {
		return false
	}

	// Match based on type
	switch pattern.MatchType {
	case MatchTypeEvent:
		// Event-based matching (e.g., priority-based)
		// For MVP, we match based on priority ranges
		// This is abstract - no real log content
		return m.matchEvent(pattern, entry)

	case MatchTypeKeyword:
		// Keyword matching would require log text
		// For MVP with metadata-only entries, we skip this
		// In real implementation, this would match against log content
		return false

	case MatchTypeRegex:
		// Regex matching would require log text and local regex mapping
		// For MVP with metadata-only entries, we skip this
		// In real implementation, this would use local regex patterns
		return false

	default:
		return false
	}
}

// matchEvent matches based on event characteristics (priority, source).
func (m *PatternMatcher) matchEvent(pattern Pattern, entry journal.Entry) bool {
	// Abstract matching based on priority
	// Priority 0-3: INFO
	// Priority 4: WARNING
	// Priority 5-7: ERROR/CRITICAL

	switch pattern.Severity {
	case rules.SeverityCrit:
		// Match critical priorities (5-7)
		return entry.Priority >= 5 && entry.Priority <= 7
	case rules.SeverityWarn:
		// Match warning priority (4)
		return entry.Priority == 4
	case rules.SeverityInfo:
		// Match info priorities (0-3)
		return entry.Priority >= 0 && entry.Priority <= 3
	default:
		return false
	}
}

// generateEventID generates an event ID from pattern and entry.
// Format: <DOMAIN>.<CATEGORY>.<EVENT>
func (m *PatternMatcher) generateEventID(pattern Pattern, entry journal.Entry) string {
	// Extract domain and category from pattern ID
	// Pattern ID format: <domain>.<category>.<event>
	parts := strings.Split(pattern.PatternID, ".")
	if len(parts) >= 3 {
		// Use pattern ID as-is if it follows the format
		return pattern.PatternID
	}

	// Generate from source and priority if pattern ID is not in correct format
	domain := "host"
	category := m.sourceToCategory(entry.Source)
	event := m.priorityToEvent(entry.Priority)

	return fmt.Sprintf("%s.%s.%s", domain, category, event)
}

// sourceToCategory converts a journal source to a category.
func (m *PatternMatcher) sourceToCategory(source string) string {
	switch source {
	case "systemd", "systemd-journald":
		return "systemd"
	case "kernel":
		return "kernel"
	case "cron":
		return "scheduler"
	default:
		return "unknown"
	}
}

// priorityToEvent converts a syslog priority to an event name.
func (m *PatternMatcher) priorityToEvent(priority int) string {
	switch {
	case priority >= 5 && priority <= 7:
		return "error"
	case priority == 4:
		return "warning"
	default:
		return "info"
	}
}
