package watcher

import "prox-watch/internal/rules"

// EvaluateSeverity bestimmt die Severity basierend auf FailCount und Thresholds.
// Phase 1: Minimal, keine Zeitfenster, keine Persistenz.
// Siehe docs/17_watcher_counter_severity.md für vollständige Spezifikation.
//
// Regeln:
//   - failCount == 0 → INFO
//   - failCount >= critThreshold → CRIT
//   - failCount >= warnThreshold → WARN
//   - else → INFO
//
// Priorität: CRIT > WARN > INFO
func EvaluateSeverity(
	failCount int,
	warnThreshold int,
	critThreshold int,
) rules.Severity {
	// Validierung: critThreshold sollte >= warnThreshold sein
	// (wird in Config-Validierung geprüft, hier defensiv)
	if critThreshold < warnThreshold {
		// Fallback: Wenn crit < warn, verwende crit als warn
		warnThreshold = critThreshold
	}

	// failCount == 0 → INFO
	if failCount == 0 {
		return rules.SeverityInfo
	}

	// failCount >= critThreshold → CRIT
	if failCount >= critThreshold {
		return rules.SeverityCrit
	}

	// failCount >= warnThreshold → WARN
	if failCount >= warnThreshold {
		return rules.SeverityWarn
	}

	// else → INFO
	return rules.SeverityInfo
}

// ShouldPush bestimmt, ob ein Push ausgelöst werden soll.
// Push nur bei Statuswechsel nach oben (newSeverity > currentSeverity).
//
// Push-Bedingungen:
//   - newSeverity > currentSeverity → Push
//   - newSeverity == currentSeverity → Kein Push
//   - newSeverity < currentSeverity → Kein Push (Verbesserung)
func ShouldPush(currentSeverity, newSeverity rules.Severity) bool {
	return newSeverity > currentSeverity
}
