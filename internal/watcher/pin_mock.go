package watcher

import (
	"fmt"
	"sync"
)

// MockPin ist eine Mock-Implementierung des Pin-Interfaces für Tests.
// Phase 1.5: Ermöglicht Tests ohne Hardware.
type MockPin struct {
	pinNumber int
	state     bool // true = HIGH, false = LOW
	mu        sync.Mutex
}

// NewMockPin erstellt einen neuen MockPin.
// Phase 1.5: Wird für Tests verwendet, später durch echte periph.io Pins ersetzt.
func NewMockPin(pinNumber int) (Pin, error) {
	if pinNumber < 0 {
		return nil, fmt.Errorf("invalid pin number: %d", pinNumber)
	}

	return &MockPin{
		pinNumber: pinNumber,
		state:     false, // Initial LOW
	}, nil
}

// High setzt den Pin auf HIGH (3.3V).
func (m *MockPin) High() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.state = true
	return nil
}

// Low setzt den Pin auf LOW (GND).
func (m *MockPin) Low() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.state = false
	return nil
}

// Close schließt den Pin und gibt Ressourcen frei.
func (m *MockPin) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Pin auf LOW setzen
	m.state = false
	return nil
}

// GetState gibt den aktuellen Zustand des Pins zurück (für Tests).
func (m *MockPin) GetState() bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.state
}

// GetPinNumber gibt die Pin-Nummer zurück (für Tests).
func (m *MockPin) GetPinNumber() int {
	return m.pinNumber
}
