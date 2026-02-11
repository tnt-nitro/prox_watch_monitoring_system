package watcher

import "prox-watch/internal/rules"

// GPIO steuert LED und Beeper.
// Phase 1: NoOp-Implementierung (MVP)
// Siehe docs/18_watcher_gpio.md für vollständige Spezifikation.
type GPIO interface {
	SetLED(sev rules.Severity) error
	Beep() error
	Close() error
}

// NoOpGPIO ist eine No-Operation-Implementierung des GPIO-Interfaces.
// Immer erfolgreich, keine Hardware, keine Nebenwirkungen.
// Phase 1: Standard-Implementierung (wenn gpio.enabled: false).
type NoOpGPIO struct{}

// NewNoOpGPIO erstellt einen neuen NoOpGPIO.
func NewNoOpGPIO() GPIO {
	return &NoOpGPIO{}
}

// SetLED setzt die LED-Farbe basierend auf Severity.
// NoOp: Immer erfolgreich, keine Hardware-Zugriffe.
func (g *NoOpGPIO) SetLED(sev rules.Severity) error {
	// No-op: Immer erfolgreich
	return nil
}

// Beep aktiviert den Beeper.
// NoOp: Immer erfolgreich, keine Hardware-Zugriffe.
func (g *NoOpGPIO) Beep() error {
	// No-op: Immer erfolgreich
	return nil
}

// Close schließt GPIO-Ressourcen.
// NoOp: Immer erfolgreich, keine Ressourcen zu schließen.
func (g *NoOpGPIO) Close() error {
	// No-op: Immer erfolgreich
	return nil
}

// TODO: Hardware-Implementierung (später, optional)
// Phase 1: Nur NoOp-GPIO (immer erfolgreich, keine Hardware)
// Später (optional): Raspberry Pi GPIO (periph.io/periph)
