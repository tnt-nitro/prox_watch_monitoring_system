package gpio

import (
	"fmt"
	"time"
)

// Beeper steuert den aktiven Buzzer (KY-012)
type Beeper struct {
	pin Pin
}

// NewBeeper erstellt einen neuen Beeper
func NewBeeper() (*Beeper, error) {
	// Initialisiere periph.io (nur bei Raspberry Pi Build)
	if err := Init(); err != nil {
		return nil, fmt.Errorf("failed to initialize GPIO: %w", err)
	}

	pin, err := NewPin(BeeperPin, PinModeOutput)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize beeper pin: %w", err)
	}

	return &Beeper{
		pin: pin,
	}, nil
}

// Beep spielt einen einzelnen Ton (Dauer in Millisekunden)
func (b *Beeper) Beep(durationMs int) error {
	if err := b.pin.High(); err != nil {
		return err
	}
	time.Sleep(time.Duration(durationMs) * time.Millisecond)
	if err := b.pin.Low(); err != nil {
		return err
	}
	return nil
}

// BeepShort spielt 3 kurze Töne
func (b *Beeper) BeepShort() error {
	for i := 0; i < 3; i++ {
		if err := b.Beep(200); err != nil {
			return err
		}
		time.Sleep(100 * time.Millisecond)
	}
	return nil
}

// BeepSOS spielt SOS-Morse-Code (· · · – – – · · ·)
func (b *Beeper) BeepSOS() error {
	// S: · · ·
	for i := 0; i < 3; i++ {
		if err := b.Beep(200); err != nil {
			return err
		}
		time.Sleep(200 * time.Millisecond)
	}

	// O: – – –
	for i := 0; i < 3; i++ {
		if err := b.Beep(600); err != nil {
			return err
		}
		time.Sleep(200 * time.Millisecond)
	}

	// S: · · ·
	for i := 0; i < 3; i++ {
		if err := b.Beep(200); err != nil {
			return err
		}
		if i < 2 {
			time.Sleep(200 * time.Millisecond)
		}
	}

	return nil
}

// Close schließt den Beeper
func (b *Beeper) Close() error {
	if err := b.pin.Low(); err != nil {
		return err
	}
	return b.pin.Close()
}
