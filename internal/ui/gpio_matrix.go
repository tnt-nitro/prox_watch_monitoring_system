package ui

import (
	"fmt"
	"strings"

	"prox-watch/internal/gpio"
)

// GPIOMatrix zeigt eine dauerhaft sichtbare GPIO-Pin-Matrix
type GPIOMatrix struct {
	neopixelPin int
	beeperPin   int
	buttonPin   int
}

// NewGPIOMatrix erstellt eine neue GPIO-Matrix
func NewGPIOMatrix() *GPIOMatrix {
	return &GPIOMatrix{
		neopixelPin: gpio.NeoPixelPin,
		beeperPin:   gpio.BeeperPin,
		buttonPin:   gpio.ButtonPin,
	}
}

// PinInfo enthält Informationen zu einem Pin
type PinInfo struct {
	PinNum      int
	GPIONum     int
	Function    string
	Color       string  // Hintergrundfarbe für Pin-Nummer
	CableColor  string  // Textfarbe für Kabel
	IsUsed      bool
	Usage       string
}

// GetPinInfo gibt Informationen zu allen Pins zurück
func (g *GPIOMatrix) GetPinInfo() []PinInfo {
	// Farben basierend auf Matrix Voice Pinout
	// Pin-Nummern: Hintergrund in Farben, Text weiß
	// Kabel: Text in Farben (nicht Hintergrund)
	const (
		// Hintergrundfarben für Pin-Nummern
		orangeBg   = "\033[48;5;208m\033[37m" // Orange Hintergrund, weißer Text
		redBg      = "\033[48;5;196m\033[37m" // Rot Hintergrund, weißer Text
		blackBg    = "\033[48;5;232m\033[37m" // Schwarz Hintergrund, weißer Text
		pinkBg     = "\033[48;5;213m\033[37m" // Pink Hintergrund, weißer Text
		greenBg    = "\033[48;5;46m\033[37m"  // Grün Hintergrund, weißer Text
		blueBg     = "\033[48;5;21m\033[37m"  // Blau Hintergrund, weißer Text
		yellowBg   = "\033[48;5;226m\033[37m" // Gelb Hintergrund, weißer Text
		
		// Textfarben für Kabel
		orangeText = "\033[38;5;208m" // Orange Text
		redText    = "\033[38;5;196m" // Rot Text
		grayText   = "\033[38;5;240m" // Dunkles Grau Text (für GND)
		pinkText   = "\033[38;5;213m" // Pink Text
		greenText  = "\033[38;5;46m"  // Grün Text
		blueText   = "\033[38;5;21m"  // Blau Text
		yellowText = "\033[38;5;226m" // Gelb Text
		
		reset    = "\033[0m"
	)

	pins := []PinInfo{
		// Pin 1-2
		{1, 0, "3.3V", orangeBg, orangeText, false, ""},
		{2, 0, "5V", redBg, redText, true, "NeoPixel 5V"},
		// Pin 3-4
		{3, 2, "GPIO2", pinkBg, pinkText, false, ""},
		{4, 0, "5V", redBg, redText, false, ""},
		// Pin 5-6
		{5, 3, "GPIO3", pinkBg, pinkText, false, ""},
		{6, 0, "GND", blackBg, grayText, true, "NeoPixel GND"},
		// Pin 7-8
		{7, 4, "GPIO4", greenBg, greenText, false, ""},
		{8, 14, "GPIO14", blueBg, blueText, false, ""},
		// Pin 9-10
		{9, 0, "GND", blackBg, grayText, false, ""},
		{10, 15, "GPIO15", blueBg, blueText, false, ""},
		// Pin 11-12
		{11, 17, "GPIO17", greenBg, greenText, false, ""},
		{12, 18, "GPIO18", greenBg, greenText, true, "NeoPixel DIN"},
		// Pin 13-14
		{13, 27, "GPIO27", greenBg, greenText, false, ""},
		{14, 0, "GND", blackBg, grayText, true, "Beeper GND"},
		// Pin 15-16
		{15, 22, "GPIO22", greenBg, greenText, false, ""},
		{16, 23, "GPIO23", greenBg, greenText, true, "Button NO"},
		// Pin 17-18
		{17, 0, "3.3V", orangeBg, orangeText, false, ""},
		{18, 24, "GPIO24", greenBg, greenText, true, "Beeper +"},
		// Pin 19-20
		{19, 10, "GPIO10", blueBg, blueText, false, ""},
		{20, 0, "GND", blackBg, grayText, true, "Button GND"},
		// Pin 21-22
		{21, 9, "GPIO9", blueBg, blueText, false, ""},
		{22, 25, "GPIO25", greenBg, greenText, false, ""},
		// Pin 23-24
		{23, 11, "GPIO11", blueBg, blueText, false, ""},
		{24, 8, "GPIO8", blueBg, blueText, false, ""},
		// Pin 25-26
		{25, 0, "GND", blackBg, grayText, false, ""},
		{26, 7, "GPIO7", blueBg, blueText, false, ""},
		// Pin 27-28
		{27, 0, "ID_SD", yellowBg, yellowText, false, ""},
		{28, 0, "ID_SC", yellowBg, yellowText, false, ""},
		// Pin 29-30
		{29, 5, "GPIO5", greenBg, greenText, false, ""},
		{30, 0, "GND", blackBg, grayText, false, ""},
		// Pin 31-32
		{31, 6, "GPIO6", greenBg, greenText, false, ""},
		{32, 12, "GPIO12", greenBg, greenText, false, ""},
		// Pin 33-34
		{33, 13, "GPIO13", greenBg, greenText, false, ""},
		{34, 0, "GND", blackBg, grayText, false, ""},
		// Pin 35-36
		{35, 19, "GPIO19", greenBg, greenText, false, ""},
		{36, 16, "GPIO16", greenBg, greenText, false, ""},
		// Pin 37-38
		{37, 26, "GPIO26", greenBg, greenText, false, ""},
		{38, 20, "GPIO20", greenBg, greenText, false, ""},
		// Pin 39-40
		{39, 0, "GND", blackBg, grayText, false, ""},
		{40, 21, "GPIO21", greenBg, greenText, false, ""},
	}

	// Markiere verwendete Pins
	for i := range pins {
		if pins[i].PinNum == 12 && pins[i].GPIONum == 18 {
			pins[i].IsUsed = true
			pins[i].Usage = "NeoPixel DIN"
		}
		if pins[i].PinNum == 18 && pins[i].GPIONum == 24 {
			pins[i].IsUsed = true
			pins[i].Usage = "Beeper +"
		}
		if pins[i].PinNum == 16 && pins[i].GPIONum == 23 {
			pins[i].IsUsed = true
			pins[i].Usage = "Button NO"
		}
	}

	return pins
}

// RenderMatrix zeigt die GPIO-Matrix im Split-Screen Format
func (g *GPIOMatrix) RenderMatrix(row, col int) {
	pins := g.GetPinInfo()
	
	// ANSI Farbcodes
	const (
		reset = "\033[0m"
		bold  = "\033[1m"
		white = "\033[37m"
	)

	// Positioniere Cursor (für Split-Screen)
	// row, col = Position des Matrices (z.B. oben rechts)
	
	// Render Matrix in 2 Spalten (wie Raspberry Pi Header)
	fmt.Printf("\033[%d;%dH", row, col)
	fmt.Printf("%sGPIO PIN MATRIX%s", bold, reset)
	
	rowOffset := row + 1
	
	for i := 0; i < len(pins); i += 2 {
		leftPin := pins[i]
		rightPin := pins[i+1]
		
		fmt.Printf("\033[%d;%dH", rowOffset+(i/2), col)
		
		// Linker Pin
		fmt.Printf("%s%2d%s %s%5s%s %s",
			leftPin.Color,
			leftPin.PinNum,
			reset,
			white,
			leftPin.Function,
			reset,
			leftPin.Color,
		)
		
		// Mittlerer Abstand
		fmt.Printf(" │ ")
		
		// Rechter Pin
		fmt.Printf("%s%5s%s %s%2d%s",
			rightPin.Color,
			rightPin.Function,
			reset,
			white,
			rightPin.PinNum,
			reset,
		)
		
		// Verwendungs-Info (wenn verwendet)
		if leftPin.IsUsed {
			fmt.Printf(" %s[%s]%s", "\033[33m", leftPin.Usage, reset)
		}
		if rightPin.IsUsed {
			fmt.Printf(" %s[%s]%s", "\033[33m", rightPin.Usage, reset)
		}
	}
}

// RenderCompactMatrix zeigt eine kompakte Matrix-Ansicht
func (g *GPIOMatrix) RenderCompactMatrix() string {
	pins := g.GetPinInfo()
	var sb strings.Builder
	
	const reset = "\033[0m"
	const white = "\033[37m"
	
	sb.WriteString("  GPIO PIN MATRIX\n")
	sb.WriteString("  ┌─────────────────────────────────────────┐\n")
	
	for i := 0; i < len(pins); i += 2 {
		leftPin := pins[i]
		rightPin := pins[i+1]
		
		sb.WriteString("  │ ")
		
		// Linker Pin
		sb.WriteString(leftPin.Color)
		sb.WriteString(fmt.Sprintf("%2d", leftPin.PinNum))
		sb.WriteString(reset)
		sb.WriteString(" ")
		sb.WriteString(white)
		sb.WriteString(fmt.Sprintf("%-5s", leftPin.Function))
		sb.WriteString(reset)
		
		// Mittlerer Abstand
		sb.WriteString(" │ ")
		
		// Rechter Pin
		sb.WriteString(white)
		sb.WriteString(fmt.Sprintf("%-5s", rightPin.Function))
		sb.WriteString(reset)
		sb.WriteString(" ")
		sb.WriteString(rightPin.Color)
		sb.WriteString(fmt.Sprintf("%2d", rightPin.PinNum))
		sb.WriteString(reset)
		
		sb.WriteString(" │")
		
		// Verwendungs-Info
		if leftPin.IsUsed {
			sb.WriteString(fmt.Sprintf(" %s[%s]%s", "\033[33m", leftPin.Usage, reset))
		}
		if rightPin.IsUsed {
			sb.WriteString(fmt.Sprintf(" %s[%s]%s", "\033[33m", rightPin.Usage, reset))
		}
		
		sb.WriteString("\n")
	}
	
	sb.WriteString("  └─────────────────────────────────────────┘\n")
	
	return sb.String()
}
