package ui

import (
	"fmt"

	"prox-watch/internal/gpio"
)

// GPIOOverview zeigt eine tabellarische GPIO-Übersicht
type GPIOOverview struct {
	neopixelPin int
	beeperPin   int
	buttonPin   int
}

// NewGPIOOverview erstellt eine neue GPIO-Übersicht
func NewGPIOOverview() *GPIOOverview {
	return &GPIOOverview{
		neopixelPin: gpio.NeoPixelPin,
		beeperPin:   gpio.BeeperPin,
		buttonPin:   gpio.ButtonPin,
	}
}

// Render zeigt die GPIO-Übersicht mit tabellarischer Darstellung
func (g *GPIOOverview) Render() {
	// ANSI Farbcodes
	const (
		reset    = "\033[0m"
		orange   = "\033[38;5;208m" // Orange
		red      = "\033[31m"       // Rot
		pink     = "\033[38;5;213m" // Pink
		black    = "\033[90m"       // Dunkelgrau (für Schwarz)
		green    = "\033[32m"       // Grün
		purple   = "\033[35m"       // Lila
		darkGray = "\033[90m"       // Dunkelgrau
		bold     = "\033[1m"
	)

	// GPIO-Pin-Farbzuordnung
	pinColors := map[int]string{
		2:  orange, 3: orange,
		4:  red, 17: red,
		27: pink, 22: pink,
		23: black, // Button
		24: green, // Beeper
		18: purple, // NeoPixel
	}

	fmt.Print("\033[2J\033[H") // Clear screen
	fmt.Println()
	fmt.Printf("  %sGPIO PIN BELEGUNG - RASPBERRY PI 3B+%s\n", bold, reset)
	fmt.Println()
	fmt.Println("  ┌─────┬──────┬──────────────────┬─────────┬──────────────┐")
	fmt.Println("  │ Pin │ GPIO │ Funktion         │ Kabel   │ Status       │")
	fmt.Println("  ├─────┼──────┼──────────────────┼─────────┼──────────────┤")

	// Wichtige Pins anzeigen
	importantPins := []struct {
		pin      int
		gpio     int
		label    string
		hasCable bool
		status   string
	}{
		{2, 0, "5V", true, "POWER"},
		{6, 0, "GND", true, "GROUND"},
		{12, 18, "NeoPixel DIN", true, "ACTIVE"},
		{14, 0, "GND", true, "GROUND"},
		{16, 23, "Button NO", true, "READY"},
		{18, 24, "Beeper +", true, "READY"},
		{20, 0, "GND", true, "GROUND"},
	}

	for _, p := range importantPins {
		color := darkGray
		if p.gpio > 0 {
			if c, ok := pinColors[p.gpio]; ok {
				color = c
			}
		}

		cable := "─────"
		if !p.hasCable {
			cable = "     "
		}

		gpioStr := fmt.Sprintf("%d", p.gpio)
		if p.gpio == 0 {
			gpioStr = "─"
		}

		fmt.Printf("  │ %2d  │ %s%4s%s │ %-16s │ %s │ %-12s │\n",
			p.pin,
			color, gpioStr, reset,
			p.label,
			cable,
			p.status,
		)
	}

	fmt.Println("  └─────┴──────┴──────────────────┴─────────┴──────────────┘")
	fmt.Println()
	fmt.Printf("  %sLegende:%s\n", bold, reset)
	fmt.Printf("  %sOrange%s = GPIO 2, 3\n", orange, reset)
	fmt.Printf("  %sRot%s = GPIO 4, 17\n", red, reset)
	fmt.Printf("  %sPink%s = GPIO 27, 22\n", pink, reset)
	fmt.Printf("  %sSchwarz%s = GPIO 23 (Button)\n", black, reset)
	fmt.Printf("  %sGrün%s = GPIO 24 (Beeper)\n", green, reset)
	fmt.Printf("  %sLila%s = GPIO 18 (NeoPixel)\n", purple, reset)
	fmt.Printf("  %sDunkelgrau%s = GND, 5V, 3.3V\n", darkGray, reset)
	fmt.Println()
}

// RenderLEDStrip zeigt die 8 NeoPixel LEDs
func (g *GPIOOverview) RenderLEDStrip(ledStates []bool) {
	const (
		reset  = "\033[0m"
		red    = "\033[31m"
		green  = "\033[32m"
		blue   = "\033[34m"
		yellow = "\033[33m"
		off    = "\033[90m"
		bold   = "\033[1m"
	)

	fmt.Print("\033[2J\033[H") // Clear screen
	fmt.Println()
	fmt.Printf("  %sNEO PIXEL STRIP (8 LEDs)%s\n", bold, reset)
	fmt.Println()
	
	// LED-Anzeige zentriert
	fmt.Print("  ")
	for i := 0; i < 8; i++ {
		if i < len(ledStates) && ledStates[i] {
			// LED ist an - zeige in Farbe
			color := yellow // Standard: Gelb
			if i == 0 {
				color = red
			} else if i == 1 {
				color = green
			} else if i == 2 {
				color = blue
			}
			fmt.Printf("%s[*]%s ", color, reset)
		} else {
			// LED ist aus
			fmt.Printf("%s[ ]%s ", off, reset)
		}
	}
	fmt.Println()
	
	// Zahlen zentriert unter den LEDs
	fmt.Print("  ")
	for i := 0; i < 8; i++ {
		fmt.Printf(" %d  ", i)
	}
	fmt.Println()
	fmt.Println()
}

// RenderCompact zeigt eine kompakte Übersicht
func (g *GPIOOverview) RenderCompact() {
	const (
		reset    = "\033[0m"
		orange   = "\033[38;5;208m"
		red      = "\033[31m"
		pink     = "\033[38;5;213m"
		black    = "\033[90m"
		green    = "\033[32m"
		purple   = "\033[35m"
		darkGray = "\033[90m"
		bold     = "\033[1m"
	)

	fmt.Print("\033[2J\033[H")
	fmt.Println()
	fmt.Printf("  %sGPIO ÜBERSICHT%s\n", bold, reset)
	fmt.Println()
	fmt.Println("  ┌─────┬──────┬─────────────────┬─────────┐")
	fmt.Println("  │ Pin │ GPIO │ Funktion        │ Status  │")
	fmt.Println("  ├─────┼──────┼─────────────────┼─────────┤")

	// NeoPixel
	fmt.Printf("  │ %2d  │ %s%4d%s │ NeoPixel DIN    │ %sACTIVE%s │\n",
		12, purple, g.neopixelPin, reset, green, reset)

	// Beeper
	fmt.Printf("  │ %2d  │ %s%4d%s │ Beeper +        │ %sREADY%s  │\n",
		18, green, g.beeperPin, reset, green, reset)

	// Button
	fmt.Printf("  │ %2d  │ %s%4d%s │ Button NO       │ %sREADY%s  │\n",
		16, black, g.buttonPin, reset, green, reset)

	fmt.Println("  └─────┴──────┴─────────────────┴─────────┘")
	fmt.Println()
}
