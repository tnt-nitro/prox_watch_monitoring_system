package ui

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// MenuItem repräsentiert ein Menü-Element
type MenuItem struct {
	Label string
	Action func() error
}

// Menu ist das Terminal-UI-Menü
type Menu struct {
	title      string
	items      []MenuItem
	cursor     int
	gpioMatrix *GPIOMatrix
}

// NewMenu erstellt ein neues Menü
func NewMenu() *Menu {
	overview := NewGPIOOverview()
	matrix := NewGPIOMatrix()
	
	return &Menu{
		title:      "LED & BEEPER TEST TOOL",
		gpioMatrix: matrix,
		items: []MenuItem{
			{Label: "LED Test", Action: func() error {
				// LED-Test ausführen (Terminal-Handling wird in RunLEDTest() gemacht)
				return RunLEDTest()
			}},
			{Label: "Beeper Test", Action: func() error {
				// TODO: Beeper-Test implementieren
				fmt.Println("\n  Beeper-Test wird implementiert...")
				fmt.Println("  Drücke Enter zum Fortfahren...")
				var buf [1]byte
				os.Stdin.Read(buf[:])
				return nil
			}},
			{Label: "GPIO Übersicht", Action: func() error {
				overview.Render()
				fmt.Println("\n  Drücke Enter zum Fortfahren...")
				var buf [1]byte
				os.Stdin.Read(buf[:])
				return nil
			}},
			{Label: "Exit", Action: func() error {
				fmt.Println("Auf Wiedersehen!")
				os.Exit(0)
				return nil
			}},
		},
		cursor: 0,
	}
}

// Run startet das Menü
func (m *Menu) Run() error {
	// Terminal in Raw-Modus setzen
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("failed to set terminal to raw mode: %w", err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	for {
		m.render()
		key := make([]byte, 3)
		n, err := os.Stdin.Read(key)
		if err != nil {
			return err
		}

		if n == 1 {
			switch key[0] {
			case 27: // ESC
				return nil
			case 13, 10: // Enter
				if m.cursor < len(m.items) {
					if m.items[m.cursor].Action != nil {
						if err := m.items[m.cursor].Action(); err != nil {
							return err
						}
					}
				}
			case 'q':
				return nil
			}
		} else if n == 3 && key[0] == 27 && key[1] == '[' {
			switch key[2] {
			case 'A': // Pfeil nach oben
				if m.cursor > 0 {
					m.cursor--
				}
			case 'B': // Pfeil nach unten
				if m.cursor < len(m.items)-1 {
					m.cursor++
				}
			}
		}
	}
}

// render zeichnet das Menü mit Split-Screen (GPIO-Matrix rechts)
func (m *Menu) render() {
	// Clear screen
	fmt.Print("\033[2J\033[H")

	const (
		reset = "\033[0m"
		bold  = "\033[1m"
	)

	// Linke Seite: Menü (positioniert)
	menuCol := 2
	menuRow := 2
	
	fmt.Printf("\033[%d;%dH", menuRow, menuCol)
	fmt.Printf("%s%s%s", bold, m.title, reset)

	// Menü-Items (untereinander)
	menuItemRow := menuRow + 2
	for i, item := range m.items {
		fmt.Printf("\033[%d;%dH", menuItemRow+i, menuCol)
		if i == m.cursor {
			fmt.Printf("%s>%s %s%s%s", "\033[7m", reset, bold, item.Label, reset)
		} else {
			fmt.Printf("  %s", item.Label)
		}
	}

	// Rechte Seite: GPIO-Matrix (dauerhaft sichtbar)
	// Position: Spalte 50, Zeile 2
	matrixCol := 50
	matrixRow := 2

	pins := m.gpioMatrix.GetPinInfo()
	const white = "\033[37m"

	// Titel der Matrix
	fmt.Printf("\033[%d;%dH", matrixRow, matrixCol)
	fmt.Printf("%sGPIO PIN MATRIX%s", bold, reset)

	// Matrix zeichnen (Format: Bedeutung Pinnummer | Pinnummer Bedeutung)
	// Separate Spalten für NeoPixel, Beeper, Button
	// Abstand: 2 Zeichen zwischen den Spalten (kompakt)
	// Rechter Pin endet bei: matrixCol + 7 + 1 + 2 + 3 + 2 + 1 + 7 = matrixCol + 23
	neopixelCol := matrixCol + 26  // Nach dem rechten Pin + 3 Zeichen Abstand (für sichtbares Kabel)
	// Längste NeoPixel-Verwendung: "[NeoPixel DIN]" = 15 Zeichen
	beeperCol := neopixelCol + 15 + 2  // Direkt nach längstem NeoPixel-Eintrag + 2 Zeichen
	// Längste Beeper-Verwendung: "[Beeper GND]" = 12 Zeichen  
	buttonCol := beeperCol + 12 + 2    // Direkt nach längstem Beeper-Eintrag + 2 Zeichen

	// Spalten-Header
	fmt.Printf("\033[%d;%dH", matrixRow, neopixelCol)
	fmt.Printf("%sNeoPixel%s", bold, reset)
	fmt.Printf("\033[%d;%dH", matrixRow, beeperCol)
	fmt.Printf("%sBeeper%s", bold, reset)
	fmt.Printf("\033[%d;%dH", matrixRow, buttonCol)
	fmt.Printf("%sButton%s", bold, reset)

	rowOffset := matrixRow + 2
	for i := 0; i < len(pins); i += 2 {
		leftPin := pins[i]
		rightPin := pins[i+1]

		fmt.Printf("\033[%d;%dH", rowOffset+(i/2), matrixCol)

		// Linker Pin: Bedeutung rechtsbündig [Farbe]Pinnummer[reset]
		leftPinEndCol := matrixCol + 7 + 1 + 2 // Function (7) + Space (1) + PinNum (2)
		fmt.Printf("%s%7s%s %s%2d%s",
			white,
			leftPin.Function,
			reset,
			leftPin.Color,
			leftPin.PinNum,
			reset,
		)

		// Trennlinie
		fmt.Printf(" │ ")

		// Rechter Pin: [Farbe]Pinnummer[reset] Bedeutung
		// Endposition: nach dem rechten Pin (PinNum + Space + Function)
		rightPinStartCol := matrixCol + 7 + 1 + 2 + 3 // Linker Pin + Trennlinie " │ "
		rightPinEndCol := rightPinStartCol + 2 + 1 + 7 // PinNum (2) + Space (1) + Function (7)
		fmt.Printf("%s%2d%s %s%-7s%s",
			rightPin.Color,
			rightPin.PinNum,
			reset,
			white,
			rightPin.Function,
			reset,
		)
		
		// Debug: Prüfe ob Endpositionen korrekt sind
		// fmt.Printf("\033[%d;%dH", rowOffset+(i/2), 1)
		// fmt.Printf("L:%d R:%d", leftPinEndCol, rightPinEndCol)

		// Linker Pin: Kabel in Pin-Farbe bis zur entsprechenden Spalte
		if leftPin.IsUsed {
			col := neopixelCol
			if leftPin.Usage == "Beeper +" || leftPin.Usage == "Beeper GND" {
				col = beeperCol
			} else if leftPin.Usage == "Button NO" || leftPin.Usage == "Button GND" {
				col = buttonCol
			}
			
			// Kabel-Länge berechnen
			cableLength := col - leftPinEndCol - 1 // -1 für das "["
			if cableLength > 0 {
				fmt.Printf("\033[%d;%dH", rowOffset+(i/2), leftPinEndCol)
				// Kabel in Pin-Textfarbe (nicht Hintergrund)
				fmt.Printf("%s%s%s[%s%s%s]%s",
					leftPin.CableColor,
					strings.Repeat("-", cableLength),
					reset,
					"\033[33m",
					leftPin.Usage,
					reset,
					reset,
				)
			}
		}

		// Rechter Pin: Kabel in Pin-Farbe bis zur entsprechenden Spalte
		if rightPin.IsUsed {
			col := neopixelCol
			if rightPin.Usage == "Beeper +" || rightPin.Usage == "Beeper GND" {
				col = beeperCol
			} else if rightPin.Usage == "Button NO" || rightPin.Usage == "Button GND" {
				col = buttonCol
			}
			
			// Kabel-Länge berechnen
			cableLength := col - rightPinEndCol - 1 // -1 für das "["
			if cableLength > 0 {
				fmt.Printf("\033[%d;%dH", rowOffset+(i/2), rightPinEndCol)
				// Kabel in Pin-Textfarbe (nicht Hintergrund)
				fmt.Printf("%s%s%s[%s%s%s]%s",
					rightPin.CableColor,
					strings.Repeat("-", cableLength),
					reset,
					"\033[33m",
					rightPin.Usage,
					reset,
					reset,
				)
			}
		}
	}

	// Cursor zurück zum Menü (erste Menüzeile)
	fmt.Printf("\033[%d;%dH", menuRow+2, menuCol)
}
