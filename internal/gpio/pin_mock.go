//go:build !raspberry
// +build !raspberry

package gpio

import (
	"fmt"
)

// mockPin ist die Mock-Implementierung für Entwicklung ohne Hardware
type mockPin struct {
	pinNum int
	mode   PinMode
	state  PinState
}

// NewPin erstellt einen neuen GPIO-Pin (Mock für Entwicklung)
func NewPin(pinNum int, mode PinMode) (Pin, error) {
	return &mockPin{
		pinNum: pinNum,
		mode:   mode,
		state:  PinStateLow,
	}, nil
}

func (p *mockPin) High() error {
	if p.mode != PinModeOutput {
		return fmt.Errorf("pin %d is not in output mode", p.pinNum)
	}
	p.state = PinStateHigh
	fmt.Printf("[MOCK] GPIO%d = HIGH\n", p.pinNum)
	return nil
}

func (p *mockPin) Low() error {
	if p.mode != PinModeOutput {
		return fmt.Errorf("pin %d is not in output mode", p.pinNum)
	}
	p.state = PinStateLow
	fmt.Printf("[MOCK] GPIO%d = LOW\n", p.pinNum)
	return nil
}

func (p *mockPin) Read() (PinState, error) {
	if p.mode == PinModeOutput {
		return p.state, nil
	}
	// Mock: Simuliere nicht gedrückten Button (HIGH wegen Pull-Up)
	return PinStateHigh, nil
}

func (p *mockPin) Close() error {
	if p.mode == PinModeOutput {
		return p.Low()
	}
	return nil
}
