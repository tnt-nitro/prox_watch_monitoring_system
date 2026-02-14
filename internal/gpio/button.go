package gpio

import (
	"fmt"
)

// Button liest den Taster (GPIO23, intern Pull-Up, gedrückt = LOW)
type Button struct {
	pin Pin
}

// NewButton erstellt einen neuen Button
func NewButton() (*Button, error) {
	// Initialisiere periph.io (nur bei Raspberry Pi Build)
	if err := Init(); err != nil {
		return nil, fmt.Errorf("failed to initialize GPIO: %w", err)
	}

	pin, err := NewPin(ButtonPin, PinModeInputPullUp)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize button pin: %w", err)
	}

	return &Button{
		pin: pin,
	}, nil
}

// IsPressed prüft, ob der Button gedrückt ist (LOW = gedrückt)
func (b *Button) IsPressed() (bool, error) {
	state, err := b.pin.Read()
	if err != nil {
		return false, err
	}
	// LOW = gedrückt (wegen Pull-Up)
	return state == PinStateLow, nil
}

// Close schließt den Button
func (b *Button) Close() error {
	return b.pin.Close()
}
