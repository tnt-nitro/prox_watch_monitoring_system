//go:build !raspberry
// +build !raspberry

package neopixel

import (
	"fmt"
)

// NewStrip erstellt einen neuen NeoPixel-Strip (Mock für Entwicklung)
func NewStrip(leds int, pin int) (*Strip, error) {
	if leds <= 0 || leds > 1024 {
		return nil, fmt.Errorf("invalid LED count: %d (must be 1-1024)", leds)
	}

	fmt.Printf("[MOCK] NeoPixel Strip initialisiert: %d LEDs, Pin %d\n", leds, pin)
	return &Strip{
		leds:  leds,
		pin:   pin,
		strip: nil,
	}, nil
}

// SetLED setzt eine einzelne LED auf eine Farbe (Mock)
func (s *Strip) SetLED(index int, color Color) error {
	if index < 0 || index >= s.leds {
		return fmt.Errorf("LED index out of range: %d (max: %d)", index, s.leds-1)
	}

	fmt.Printf("[MOCK] LED %d → RGB(%d, %d, %d)\n", index, color.R, color.G, color.B)
	return nil
}

// Render sendet die LED-Daten an die Hardware (Mock)
func (s *Strip) Render() error {
	fmt.Println("[MOCK] Render() - LEDs würden jetzt aktualisiert")
	return nil
}

// Close schließt den Strip (Mock)
func (s *Strip) Close() error {
	fmt.Println("[MOCK] NeoPixel Strip geschlossen")
	return nil
}
