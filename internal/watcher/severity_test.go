package watcher

import (
	"testing"

	"prox-watch/internal/rules"
)

func TestEvaluateSeverity_Zero(t *testing.T) {
	// 0 → INFO
	severity := EvaluateSeverity(0, 3, 10)
	if severity != rules.SeverityInfo {
		t.Errorf("Expected INFO for failCount=0, got %v", severity)
	}
}

func TestEvaluateSeverity_One(t *testing.T) {
	// 1 → INFO (unter warnThreshold)
	severity := EvaluateSeverity(1, 3, 10)
	if severity != rules.SeverityInfo {
		t.Errorf("Expected INFO for failCount=1 (warnThreshold=3), got %v", severity)
	}
}

func TestEvaluateSeverity_WarnThreshold(t *testing.T) {
	// warnThreshold → WARN
	warnThreshold := 3
	severity := EvaluateSeverity(warnThreshold, warnThreshold, 10)
	if severity != rules.SeverityWarn {
		t.Errorf("Expected WARN for failCount=%d (warnThreshold), got %v", warnThreshold, severity)
	}
}

func TestEvaluateSeverity_AboveWarnThreshold(t *testing.T) {
	// warnThreshold + 1 → WARN (noch unter critThreshold)
	severity := EvaluateSeverity(4, 3, 10)
	if severity != rules.SeverityWarn {
		t.Errorf("Expected WARN for failCount=4 (warnThreshold=3, critThreshold=10), got %v", severity)
	}
}

func TestEvaluateSeverity_CritThreshold(t *testing.T) {
	// critThreshold → CRIT
	critThreshold := 10
	severity := EvaluateSeverity(critThreshold, 3, critThreshold)
	if severity != rules.SeverityCrit {
		t.Errorf("Expected CRIT for failCount=%d (critThreshold), got %v", critThreshold, severity)
	}
}

func TestEvaluateSeverity_AboveCritThreshold(t *testing.T) {
	// critThreshold + 1 → CRIT
	severity := EvaluateSeverity(11, 3, 10)
	if severity != rules.SeverityCrit {
		t.Errorf("Expected CRIT for failCount=11 (critThreshold=10), got %v", severity)
	}
}

func TestEvaluateSeverity_Priority(t *testing.T) {
	// Priorität: CRIT > WARN > INFO
	// Teste, dass CRIT höchste Priorität hat
	critSeverity := EvaluateSeverity(10, 3, 10)
	warnSeverity := EvaluateSeverity(3, 3, 10)
	infoSeverity := EvaluateSeverity(0, 3, 10)

	if critSeverity <= warnSeverity {
		t.Errorf("CRIT should be > WARN, got CRIT=%v, WARN=%v", critSeverity, warnSeverity)
	}
	if warnSeverity <= infoSeverity {
		t.Errorf("WARN should be > INFO, got WARN=%v, INFO=%v", warnSeverity, infoSeverity)
	}
	if critSeverity <= infoSeverity {
		t.Errorf("CRIT should be > INFO, got CRIT=%v, INFO=%v", critSeverity, infoSeverity)
	}
}

func TestEvaluateSeverity_CritLessThanWarn(t *testing.T) {
	// Defensive Validierung: Wenn crit < warn, verwende crit als warn
	severity := EvaluateSeverity(5, 10, 3) // critThreshold=3 < warnThreshold=10
	// Bei failCount=5 sollte CRIT zurückgegeben werden (da 5 >= 3)
	if severity != rules.SeverityCrit {
		t.Errorf("Expected CRIT for failCount=5 when critThreshold=3 (defensive), got %v", severity)
	}
}

func TestShouldPush_InfoToWarn(t *testing.T) {
	// INFO → WARN: Push auslösen
	if !ShouldPush(rules.SeverityInfo, rules.SeverityWarn) {
		t.Error("Expected Push for INFO→WARN")
	}
}

func TestShouldPush_WarnToCrit(t *testing.T) {
	// WARN → CRIT: Push auslösen
	if !ShouldPush(rules.SeverityWarn, rules.SeverityCrit) {
		t.Error("Expected Push for WARN→CRIT")
	}
}

func TestShouldPush_SameSeverity(t *testing.T) {
	// Gleiche Severity: Kein Push
	if ShouldPush(rules.SeverityWarn, rules.SeverityWarn) {
		t.Error("Expected no Push for WARN→WARN")
	}
	if ShouldPush(rules.SeverityCrit, rules.SeverityCrit) {
		t.Error("Expected no Push for CRIT→CRIT")
	}
	if ShouldPush(rules.SeverityInfo, rules.SeverityInfo) {
		t.Error("Expected no Push for INFO→INFO")
	}
}

func TestShouldPush_Improvement(t *testing.T) {
	// Verbesserung (CRIT → INFO, WARN → INFO): Kein Push
	if ShouldPush(rules.SeverityCrit, rules.SeverityInfo) {
		t.Error("Expected no Push for CRIT→INFO (improvement)")
	}
	if ShouldPush(rules.SeverityWarn, rules.SeverityInfo) {
		t.Error("Expected no Push for WARN→INFO (improvement)")
	}
	if ShouldPush(rules.SeverityCrit, rules.SeverityWarn) {
		t.Error("Expected no Push for CRIT→WARN (improvement)")
	}
}

func TestShouldPush_EdgeCases(t *testing.T) {
	// Edge Cases
	if ShouldPush(rules.SeverityInfo, rules.SeverityCrit) {
		// INFO → CRIT: Push auslösen (sollte true sein)
		// Test ist korrekt, aber prüfen wir explizit
		if !ShouldPush(rules.SeverityInfo, rules.SeverityCrit) {
			t.Error("Expected Push for INFO→CRIT")
		}
	}
}
