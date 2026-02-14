//go:build raspberry
// +build raspberry

package neopixel

import (
	"fmt"

	ws2811 "github.com/rpi-ws281x/rpi-ws281x-go"
)

// NewStrip erstellt einen neuen NeoPixel-Strip mit echter WS2812-Hardware
func NewStrip(leds int, pin int) (*Strip, error) {
	if leds <= 0 || leds > 1024 {
		return nil, fmt.Errorf("invalid LED count: %d (must be 1-1024)", leds)
	}

	// WS2812-Optionen konfigurieren
	opt := ws2811.DefaultOptions
	opt.Channels[0].GpioPin = pin
	opt.Channels[0].LedCount = leds
	opt.Channels[0].Brightness = 255
	// WS2812 verwendet GRB Format (nicht RGB!)
	opt.Channels[0].StripeType = ws2811.WS2811StripGRB

	// WS2812 initialisieren
	err := ws2811.Init(&opt)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize WS2812 (Pin %d, %d LEDs): %w", pin, leds, err)
	}

	return &Strip{
		leds:  leds,
		pin:   pin,
		strip: nil, // Nicht benötigt, da ws2811 global ist
	}, nil
}

// SetLED setzt eine einzelne LED auf eine Farbe (GRB Format!)
func (s *Strip) SetLED(index int, color Color) error {
	if index < 0 || index >= s.leds {
		return fmt.Errorf("LED index out of range: %d (max: %d)", index, s.leds-1)
	}

	// WS2812 verwendet GRB Format (nicht RGB!)
	// Format: 0x00GGRRBB
	grbColor := uint32(color.G)<<16 | uint32(color.R)<<8 | uint32(color.B)
	ws2811.SetLed(index, grbColor)
	return nil
}

// Render sendet die LED-Daten an die Hardware
func (s *Strip) Render() error {
	err := ws2811.Render()
	if err != nil {
		return fmt.Errorf("failed to render LEDs: %w", err)
	}
	return nil
}

// Close schließt den Strip und gibt Ressourcen frei
func (s *Strip) Close() error {
	// Alle LEDs ausschalten
	if err := s.Clear(); err != nil {
		return err
	}
	if err := s.Render(); err != nil {
		return err
	}
	
	// WS2812-Bibliothek schließen
	ws2811.Fini()
	return nil
}
