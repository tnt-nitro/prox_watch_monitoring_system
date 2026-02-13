package state

import (
	"time"
)

// Event represents a tracked event with its counter state.
// Severity is stored as int to avoid import cycles.
// Use SeverityToInt() and IntToSeverity() helpers for conversion.
type Event struct {
	EventID   string
	Severity  int // Stored as int (rules.Severity), converted when needed
	Count     int
	FirstSeen time.Time
	LastSeen  time.Time
}

// CountState represents the counting state for an event.
type CountState struct {
	EventID    string
	Count      int
	FirstSeen  time.Time
	LastSeen   time.Time
	WindowEnds time.Time
}

// CooldownState represents a cooldown period for an event.
type CooldownState struct {
	EventID       string
	CooldownUntil time.Time
}

// AckState represents an acknowledgment state for an event.
type AckState struct {
	EventID  string
	AckUntil time.Time
}
