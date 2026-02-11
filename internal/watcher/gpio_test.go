package watcher

import (
	"prox-watch/internal/rules"
	"testing"
)

func TestNoOpGPIO_SetLED(t *testing.T) {
	// SetLED → kein Fehler
	gpio := NewNoOpGPIO()

	// Teste alle Severity-Levels
	severities := []rules.Severity{
		rules.SeverityInfo,
		rules.SeverityWarn,
		rules.SeverityCrit,
	}

	for _, sev := range severities {
		if err := gpio.SetLED(sev); err != nil {
			t.Errorf("SetLED(%v) failed: %v", sev, err)
		}
	}
}

func TestNoOpGPIO_Beep(t *testing.T) {
	// Beep → kein Fehler
	gpio := NewNoOpGPIO()

	if err := gpio.Beep(); err != nil {
		t.Errorf("Beep() failed: %v", err)
	}
}

func TestNoOpGPIO_Close(t *testing.T) {
	// Close → kein Fehler
	gpio := NewNoOpGPIO()

	if err := gpio.Close(); err != nil {
		t.Errorf("Close() failed: %v", err)
	}
}

func TestNoOpGPIO_MultipleCalls(t *testing.T) {
	// Mehrfacher Aufruf → stabil
	gpio := NewNoOpGPIO()

	// SetLED mehrfach
	for i := 0; i < 10; i++ {
		if err := gpio.SetLED(rules.SeverityInfo); err != nil {
			t.Errorf("SetLED() failed on call %d: %v", i, err)
		}
		if err := gpio.SetLED(rules.SeverityWarn); err != nil {
			t.Errorf("SetLED() failed on call %d: %v", i, err)
		}
		if err := gpio.SetLED(rules.SeverityCrit); err != nil {
			t.Errorf("SetLED() failed on call %d: %v", i, err)
		}
	}

	// Beep mehrfach
	for i := 0; i < 10; i++ {
		if err := gpio.Beep(); err != nil {
			t.Errorf("Beep() failed on call %d: %v", i, err)
		}
	}

	// Close mehrfach
	for i := 0; i < 10; i++ {
		if err := gpio.Close(); err != nil {
			t.Errorf("Close() failed on call %d: %v", i, err)
		}
	}
}

func TestNoOpGPIO_NoState(t *testing.T) {
	// Prüfe, dass kein Zustand gespeichert wird
	gpio := NewNoOpGPIO()

	// SetLED sollte keinen Zustand ändern
	if err := gpio.SetLED(rules.SeverityInfo); err != nil {
		t.Fatalf("SetLED() failed: %v", err)
	}

	// Beep sollte funktionieren, unabhängig vom vorherigen SetLED
	if err := gpio.Beep(); err != nil {
		t.Errorf("Beep() failed after SetLED: %v", err)
	}

	// Close sollte funktionieren, unabhängig von vorherigen Aufrufen
	if err := gpio.Close(); err != nil {
		t.Errorf("Close() failed after SetLED/Beep: %v", err)
	}
}

func TestNoOpGPIO_AllMethods(t *testing.T) {
	// Alle Methoden sollten funktionieren
	gpio := NewNoOpGPIO()

	// SetLED
	if err := gpio.SetLED(rules.SeverityInfo); err != nil {
		t.Errorf("SetLED(INFO) failed: %v", err)
	}

	// Beep
	if err := gpio.Beep(); err != nil {
		t.Errorf("Beep() failed: %v", err)
	}

	// Close
	if err := gpio.Close(); err != nil {
		t.Errorf("Close() failed: %v", err)
	}
}

func TestNoOpGPIO_InterfaceCompliance(t *testing.T) {
	// Prüfe, dass NoOpGPIO das GPIO-Interface erfüllt
	var _ GPIO = &NoOpGPIO{}

	gpio := NewNoOpGPIO()
	var iface GPIO = gpio

	// Prüfe, dass alle Methoden aufgerufen werden können
	if err := iface.SetLED(rules.SeverityInfo); err != nil {
		t.Errorf("SetLED() via interface failed: %v", err)
	}

	if err := iface.Beep(); err != nil {
		t.Errorf("Beep() via interface failed: %v", err)
	}

	if err := iface.Close(); err != nil {
		t.Errorf("Close() via interface failed: %v", err)
	}
}
