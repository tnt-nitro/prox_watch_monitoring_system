package pattern

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"prox-watch/internal/journal"
	"prox-watch/internal/rules"
)

func TestPatternMatcher_Match(t *testing.T) {
	matcher := NewPatternMatcher()

	// Create test patterns
	patterns := []Pattern{
		{
			PatternID: "host.kernel.error",
			Source:    "kernel",
			MatchType: MatchTypeEvent,
			Severity:  rules.SeverityCrit,
			CountRule: CountRule{Threshold: 1, Window: 0},
		},
		{
			PatternID: "host.systemd.warning",
			Source:    "systemd",
			MatchType: MatchTypeEvent,
			Severity:  rules.SeverityWarn,
			CountRule: CountRule{Threshold: 3, Window: 10 * time.Minute},
		},
	}

	// Manually register patterns
	for _, pattern := range patterns {
		matcher.registry.patterns[pattern.PatternID] = pattern
	}

	ctx := context.Background()
	now := time.Now()

	tests := []struct {
		name     string
		entry    journal.Entry
		expected *MatchResult
	}{
		{
			name: "match_crit_kernel",
			entry: journal.Entry{
				Priority:  6, // ERROR
				Source:    "kernel",
				Timestamp: now,
			},
			expected: &MatchResult{
				EventID:   "host.kernel.error",
				Severity:  rules.SeverityCrit,
				PatternID: "host.kernel.error",
			},
		},
		{
			name: "match_warn_systemd",
			entry: journal.Entry{
				Priority:  4, // WARNING
				Source:    "systemd",
				Timestamp: now,
			},
			expected: &MatchResult{
				EventID:   "host.systemd.warning",
				Severity:  rules.SeverityWarn,
				PatternID: "host.systemd.warning",
			},
		},
		{
			name: "no_match",
			entry: journal.Entry{
				Priority:  2, // INFO
				Source:    "unknown",
				Timestamp: now,
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := matcher.Match(ctx, tt.entry)
			if err != nil {
				t.Fatalf("Match() error = %v", err)
			}

			if tt.expected == nil {
				if result != nil {
					t.Errorf("Match() = %v, want nil", result)
				}
				return
			}

			if result == nil {
				t.Fatalf("Match() = nil, want %v", tt.expected)
			}

			if result.EventID != tt.expected.EventID {
				t.Errorf("Match() EventID = %s, want %s", result.EventID, tt.expected.EventID)
			}
			if result.Severity != tt.expected.Severity {
				t.Errorf("Match() Severity = %v, want %v", result.Severity, tt.expected.Severity)
			}
			if result.PatternID != tt.expected.PatternID {
				t.Errorf("Match() PatternID = %s, want %s", result.PatternID, tt.expected.PatternID)
			}
		})
	}
}

func TestPatternMatcher_GenerateEventID(t *testing.T) {
	matcher := NewPatternMatcher()

	tests := []struct {
		name     string
		pattern  Pattern
		entry    journal.Entry
		expected string
	}{
		{
			name: "pattern_id_format",
			pattern: Pattern{
				PatternID: "host.network.link_down",
			},
			entry: journal.Entry{
				Source: "kernel",
			},
			expected: "host.network.link_down",
		},
		{
			name: "generated_from_source",
			pattern: Pattern{
				PatternID: "test",
			},
			entry: journal.Entry{
				Source:   "kernel",
				Priority: 6,
			},
			expected: "host.kernel.error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matcher.generateEventID(tt.pattern, tt.entry)
			if result != tt.expected {
				t.Errorf("generateEventID() = %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestRegistry_LoadPatterns(t *testing.T) {
	// Create temporary pattern file
	tmpDir := t.TempDir()
	patternFile := filepath.Join(tmpDir, "patterns.yaml")

	patterns := []Pattern{
		{
			PatternID: "host.kernel.error",
			Source:    "kernel",
			MatchType: MatchTypeEvent,
			Severity:  rules.SeverityCrit,
		},
	}

	data := `- pattern_id: host.kernel.error
  source: kernel
  match_type: 2
  severity: 2
`

	if err := os.WriteFile(patternFile, []byte(data), 0644); err != nil {
		t.Fatalf("Failed to write pattern file: %v", err)
	}

	registry := NewRegistry()
	if err := registry.LoadPatterns(patternFile); err != nil {
		t.Fatalf("LoadPatterns() error = %v", err)
	}

	pattern, exists := registry.GetPattern("host.kernel.error")
	if !exists {
		t.Error("GetPattern() = false, want true")
	}
	if pattern.PatternID != "host.kernel.error" {
		t.Errorf("GetPattern() PatternID = %s, want host.kernel.error", pattern.PatternID)
	}
}

func TestRegistry_DuplicatePattern(t *testing.T) {
	registry := NewRegistry()

	pattern := Pattern{
		PatternID: "test.pattern",
		Source:    "test",
		MatchType: MatchTypeEvent,
		Severity:  rules.SeverityInfo,
	}

	// Add first pattern
	registry.patterns[pattern.PatternID] = pattern

	// Try to add duplicate
	err := registry.validatePattern(pattern)
	if err != nil {
		t.Fatalf("validatePattern() error = %v", err)
	}

	// Check duplicate detection
	if _, exists := registry.patterns[pattern.PatternID]; !exists {
		t.Error("Pattern should exist")
	}
}

func TestPatternMatcher_ContextCancellation(t *testing.T) {
	matcher := NewPatternMatcher()
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel context immediately
	cancel()

	entry := journal.Entry{
		Priority:  6,
		Source:    "kernel",
		Timestamp: time.Now(),
	}

	result, err := matcher.Match(ctx, entry)
	if err != context.Canceled {
		t.Errorf("Match() error = %v, want context.Canceled", err)
	}
	if result != nil {
		t.Errorf("Match() result = %v, want nil", result)
	}
}
