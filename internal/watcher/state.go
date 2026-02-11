package watcher

import "prox-watch/internal/rules"

// State verwaltet den aktuellen Zustand des Watchers.
// Phase 1: In-Memory, keine Persistenz
// Siehe docs/17_watcher_counter_severity.md für vollständige Spezifikation.
type State struct {
	CurrentSeverity rules.Severity
	FailCount       int
	WarnThreshold   int
	CritThreshold   int
}

// NewState erstellt einen neuen State mit Default-Werten.
func NewState() *State {
	return &State{
		CurrentSeverity: rules.SeverityInfo,
		FailCount:       0,
		WarnThreshold:   3,  // Default
		CritThreshold:   10, // Default
	}
}

// NewStateWithThresholds erstellt einen neuen State mit konfigurierten Thresholds.
func NewStateWithThresholds(warnThreshold, critThreshold int) *State {
	return &State{
		CurrentSeverity: rules.SeverityInfo,
		FailCount:       0,
		WarnThreshold:   warnThreshold,
		CritThreshold:   critThreshold,
	}
}
