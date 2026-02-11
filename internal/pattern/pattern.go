package pattern

import (
	"time"

	"prox-watch/internal/rules"
)

// Pattern represents a pattern definition (metadata only).
type Pattern struct {
	PatternID string      `yaml:"pattern_id"`
	Source    string      `yaml:"source"`
	MatchType MatchType   `yaml:"match_type"`
	Severity  rules.Severity `yaml:"severity"`
	CountRule CountRule   `yaml:"count_rule,omitempty"`
}

// MatchType represents the type of pattern matching.
type MatchType int

const (
	// MatchTypeKeyword matches against keywords.
	MatchTypeKeyword MatchType = iota
	// MatchTypeRegex matches against regex (local only, not in repo).
	MatchTypeRegex
	// MatchTypeEvent matches against system events.
	MatchTypeEvent
)

// CountRule defines counting rules for a pattern.
type CountRule struct {
	Threshold int           `yaml:"threshold"`
	Window    time.Duration `yaml:"window"`
}

// MatchResult represents the result of pattern matching.
type MatchResult struct {
	EventID   string
	Severity  rules.Severity
	PatternID string
}
