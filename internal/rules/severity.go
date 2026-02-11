package rules

// Severity represents the severity level of an event.
type Severity int

const (
	// SeverityInfo represents informational events (no push).
	SeverityInfo Severity = iota
	// SeverityWarn represents warning events (push optional).
	SeverityWarn
	// SeverityCrit represents critical events (push always).
	SeverityCrit
)

// String returns the string representation of the severity level.
func (s Severity) String() string {
	switch s {
	case SeverityInfo:
		return "INFO"
	case SeverityWarn:
		return "WARN"
	case SeverityCrit:
		return "CRIT"
	default:
		return "UNKNOWN"
	}
}
