package rules

import (
	"time"

	"prox-watch/internal/config"
	"prox-watch/internal/state"
)

// Evaluator evaluates severity based on count state and rules.
type Evaluator struct {
	cfg *config.Config
}

// NewEvaluator creates a new severity evaluator.
func NewEvaluator(cfg *config.Config) *Evaluator {
	return &Evaluator{cfg: cfg}
}

// Evaluate evaluates the severity based on count state and time windows.
// Returns the highest severity that matches the criteria.
// Priority: CRIT > WARN > INFO
func (e *Evaluator) Evaluate(cs state.CountState, now time.Time, isHardError bool) Severity {
	// Hard error → immediately CRIT (highest priority)
	if isHardError {
		return SeverityCrit
	}

	// Check CRIT threshold and window (highest priority after hard error)
	critWindow := e.getWindow("crit")
	critThreshold := e.cfg.Rules.Thresholds.Crit
	if cs.Count >= critThreshold {
		windowEnd := cs.FirstSeen.Add(critWindow)
		if now.Before(windowEnd) || now.Equal(windowEnd) {
			return SeverityCrit
		}
		// Window expired, continue to check lower severities
	}

	// Check WARN threshold and window
	warnWindow := e.getWindow("warn")
	warnThreshold := e.cfg.Rules.Thresholds.Warn
	if cs.Count >= warnThreshold {
		windowEnd := cs.FirstSeen.Add(warnWindow)
		if now.Before(windowEnd) || now.Equal(windowEnd) {
			return SeverityWarn
		}
		// Window expired, continue to check lower severities
	}

	// Default: INFO
	// INFO threshold is always 1, no window
	if cs.Count >= 1 {
		return SeverityInfo
	}

	// No events yet
	return SeverityInfo
}

// getWindow returns the time window duration for a severity level.
func (e *Evaluator) getWindow(severity string) time.Duration {
	duration, err := e.cfg.Rules.GetWindowDuration(severity)
	if err != nil {
		// Fallback to defaults
		switch severity {
		case "warn":
			return 10 * time.Minute
		case "crit":
			return 15 * time.Minute
		default:
			return 0
		}
	}
	return duration
}

// GetThreshold returns the threshold for a severity level.
func (e *Evaluator) GetThreshold(severity Severity) int {
	switch severity {
	case SeverityInfo:
		return 1
	case SeverityWarn:
		return e.cfg.Rules.Thresholds.Warn
	case SeverityCrit:
		return e.cfg.Rules.Thresholds.Crit
	default:
		return 1
	}
}

// GetWindow returns the time window for a severity level.
func (e *Evaluator) GetWindow(severity Severity) time.Duration {
	switch severity {
	case SeverityInfo:
		return 0 // No window for INFO
	case SeverityWarn:
		return e.getWindow("warn")
	case SeverityCrit:
		return e.getWindow("crit")
	default:
		return 0
	}
}

// IsHardError checks if an event ID represents a hard error.
// Hard errors are events that should immediately trigger CRIT severity.
func (e *Evaluator) IsHardError(eventID string) bool {
	// Hard error patterns (abstract, no real data)
	hardErrorPatterns := []string{
		"host.network.link_down",
		"host.kernel.panic",
		"container.lifecycle.crash",
		"service.crash",
	}

	for _, pattern := range hardErrorPatterns {
		if eventID == pattern {
			return true
		}
	}

	return false
}
