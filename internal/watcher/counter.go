package watcher

// Counter verwaltet Fehlzähler für Health-Checks.
// Phase 1: Minimal, In-Memory, keine Persistenz
// WARN → ≥3 Fehlversuche (aus Config)
// CRIT → ≥10 Fehlversuche (aus Config)
// Siehe docs/17_watcher_counter_severity.md für vollständige Spezifikation.
type Counter interface {
	Increment() error
	GetCount() int
	Reset() error
}

// counter ist die In-Memory-Implementierung des Counter-Interfaces.
// Single-Thread, keine Nebenläufigkeit, keine Persistenz.
type counter struct {
	failCount int
}

// NewCounter erstellt einen neuen Counter mit initialem Wert 0.
func NewCounter() Counter {
	return &counter{
		failCount: 0,
	}
}

// Increment erhöht den Fehlzähler um 1 und gibt den neuen Wert zurück.
// Thread-safe durch Single-Thread-Design (kein Mutex erforderlich).
func (c *counter) Increment() error {
	c.failCount++
	return nil
}

// GetCount gibt den aktuellen Fehlzähler zurück.
func (c *counter) GetCount() int {
	return c.failCount
}

// Reset setzt den Fehlzähler auf 0 zurück.
func (c *counter) Reset() error {
	c.failCount = 0
	return nil
}
