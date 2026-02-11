package watcher

// Pin ist die Abstraktion für einen GPIO-Pin.
// Ermöglicht Tests ohne Hardware und Austauschbarkeit der GPIO-Bibliothek.
type Pin interface {
	// High setzt den Pin auf HIGH (3.3V).
	High() error

	// Low setzt den Pin auf LOW (GND).
	Low() error

	// Close schließt den Pin und gibt Ressourcen frei.
	Close() error
}

// TODO: Implementierung folgt in Phase 1.5
// Pin-Abstraktion ermöglicht:
// - Tests ohne Hardware (MockPin)
// - Austauschbarkeit (periph.io, sysfs, etc.)
// - Saubere Trennung von Hardware-Logik
