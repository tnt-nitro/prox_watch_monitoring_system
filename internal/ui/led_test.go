package ui

import (
	"fmt"
	"os"
	"time"

	"golang.org/x/term"
	"prox-watch/internal/gpio"
	"prox-watch/internal/neopixel"
)

// RunLEDTest führt den LED-Test durch
func RunLEDTest() error {
	// Clear screen und Terminal zurücksetzen
	fmt.Print("\033[2J\033[H")
	fmt.Print("\033[?25h") // Cursor anzeigen
	
	// NeoPixel initialisieren
	fmt.Println("  Initialisiere NeoPixel...")
	fmt.Printf("  GPIO Pin: %d (BCM, Physical Pin 12)\n", gpio.NeoPixelPin)
	
	strip, err := neopixel.NewStrip(8, gpio.NeoPixelPin)
	if err != nil {
		fmt.Printf("\n  ❌ FEHLER bei Initialisierung: %v\n", err)
		fmt.Println("\n  Drücke Enter zum Fortfahren...")
		var buf [1]byte
		os.Stdin.Read(buf[:])
		return fmt.Errorf("failed to initialize NeoPixel: %w", err)
	}
	defer strip.Close()
	
	fmt.Println("  ✓ NeoPixel initialisiert")
	time.Sleep(500 * time.Millisecond)

	overview := NewGPIOOverview()

	// Test-Sequenz
	tests := []struct {
		name      string
		ledStates []bool
		colors    []neopixel.Color
		duration  time.Duration
	}{
		{
			name:      "LED 0 → Rot",
			ledStates: []bool{true, false, false, false, false, false, false, false},
			colors:    []neopixel.Color{neopixel.ColorRed, neopixel.ColorOff, neopixel.ColorOff, neopixel.ColorOff, neopixel.ColorOff, neopixel.ColorOff, neopixel.ColorOff, neopixel.ColorOff},
			duration:  3 * time.Second,
		},
		{
			name:      "LED 1 → Grün",
			ledStates: []bool{false, true, false, false, false, false, false, false},
			colors:    []neopixel.Color{neopixel.ColorOff, neopixel.ColorGreen, neopixel.ColorOff, neopixel.ColorOff, neopixel.ColorOff, neopixel.ColorOff, neopixel.ColorOff, neopixel.ColorOff},
			duration:  3 * time.Second,
		},
		{
			name:      "LED 2 → Blau",
			ledStates: []bool{false, false, true, false, false, false, false, false},
			colors:    []neopixel.Color{neopixel.ColorOff, neopixel.ColorOff, neopixel.ColorBlue, neopixel.ColorOff, neopixel.ColorOff, neopixel.ColorOff, neopixel.ColorOff, neopixel.ColorOff},
			duration:  3 * time.Second,
		},
		{
			name:      "Alle aus",
			ledStates: []bool{false, false, false, false, false, false, false, false},
			colors:    []neopixel.Color{neopixel.ColorOff, neopixel.ColorOff, neopixel.ColorOff, neopixel.ColorOff, neopixel.ColorOff, neopixel.ColorOff, neopixel.ColorOff, neopixel.ColorOff},
			duration:  2 * time.Second,
		},
	}

	// Tests durchführen
	for _, test := range tests {
		// Clear screen
		fmt.Print("\033[2J\033[H")

		// LEDs setzen
		for i, color := range test.colors {
			if err := strip.SetLED(i, color); err != nil {
				return fmt.Errorf("failed to set LED %d: %w", i, err)
			}
		}

		// Render (Hardware aktualisieren)
		if err := strip.Render(); err != nil {
			fmt.Print("\033[2J\033[H") // Clear screen
			fmt.Printf("  ❌ FEHLER beim Render: %v\n", err)
			fmt.Println("\n  Drücke Enter zum Fortfahren...")
			var buf [1]byte
			os.Stdin.Read(buf[:])
			return fmt.Errorf("failed to render LEDs: %w", err)
		}

		// GUI aktualisieren (nach erfolgreichem Render)
		overview.RenderLEDStrip(test.ledStates)

		// Test-Name anzeigen
		fmt.Printf("  %s%s%s\n", "\033[1m", test.name, "\033[0m")
		fmt.Println()
		fmt.Println("  Drücke Enter zum Fortfahren...")

		// Warten auf Enter oder Timeout
		done := make(chan bool, 1)
		go func() {
			var buf [1]byte
			os.Stdin.Read(buf[:])
			if buf[0] == 13 || buf[0] == 10 { // Enter
				done <- true
			}
		}()

		select {
		case <-done:
			// Enter gedrückt - weiter zum nächsten Test
		case <-time.After(test.duration):
			// Timeout - weiter zum nächsten Test
		}
	}

	// Lauflicht
	fmt.Print("\033[2J\033[H")
	fmt.Println("  Lauflicht...")
	fmt.Println("  Drücke Enter zum Beenden...")

	done := make(chan bool, 1)
	go func() {
		var buf [1]byte
		os.Stdin.Read(buf[:])
		if buf[0] == 13 || buf[0] == 10 { // Enter
			done <- true
		}
	}()

	ledIndex := 0
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			// Alle LEDs ausschalten
			strip.Clear()
			strip.Render()
			overview.RenderLEDStrip([]bool{false, false, false, false, false, false, false, false})
			return nil
		case <-ticker.C:
			// Alle LEDs ausschalten
			strip.Clear()

			// Aktuelle LED setzen
			var color neopixel.Color
			switch ledIndex % 3 {
			case 0:
				color = neopixel.ColorRed
			case 1:
				color = neopixel.ColorGreen
			case 2:
				color = neopixel.ColorBlue
			}

			strip.SetLED(ledIndex, color)
			strip.Render()

			// GUI aktualisieren
			ledStates := make([]bool, 8)
			ledStates[ledIndex] = true
			overview.RenderLEDStrip(ledStates)

			// Nächste LED
			ledIndex = (ledIndex + 1) % 8
		}
	}
}
