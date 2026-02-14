package neopixel

// Strip steuert einen WS2812 NeoPixel-Strip (8 LEDs)
type Strip struct {
	leds    int
	pin     int
	strip   interface{} // Wird von Build-Tag-spezifischen Implementierungen verwendet
}

// Color repräsentiert eine RGB-Farbe
type Color struct {
	R, G, B uint8
}

// Standardfarben
var (
	ColorRed   = Color{R: 255, G: 0, B: 0}
	ColorGreen = Color{R: 0, G: 255, B: 0}
	ColorBlue  = Color{R: 0, G: 0, B: 255}
	ColorOff   = Color{R: 0, G: 0, B: 0}
)

// NewStrip erstellt einen neuen NeoPixel-Strip
// Implementierung abhängig vom Build-Tag (raspberry oder !raspberry)
// Siehe: strip_raspberry.go (Hardware) und strip_mock.go (Mock)

// SetLED setzt eine einzelne LED auf eine Farbe
// Implementierung abhängig vom Build-Tag
// Siehe: strip_raspberry.go (Hardware) und strip_mock.go (Mock)

// Render sendet die LED-Daten an die Hardware
// Implementierung abhängig vom Build-Tag
// Siehe: strip_raspberry.go (Hardware) und strip_mock.go (Mock)

// SetAll setzt alle LEDs auf eine Farbe
func (s *Strip) SetAll(color Color) error {
	for i := 0; i < s.leds; i++ {
		if err := s.SetLED(i, color); err != nil {
			return err
		}
	}
	return nil
}

// Clear schaltet alle LEDs aus
func (s *Strip) Clear() error {
	return s.SetAll(ColorOff)
}

// Close schließt den Strip
// Implementierung abhängig vom Build-Tag
// Siehe: strip_raspberry.go (Hardware) und strip_mock.go (Mock)
