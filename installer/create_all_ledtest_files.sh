#!/bin/bash
# Erstellt ALLE notwendigen Dateien für das LED-Test-Tool auf dem Raspberry Pi
# 
# Verwendung: 
#   cd ~/prox_watch_monitoring_system
#   bash installer/create_all_ledtest_files.sh

set -euo pipefail

PROJECT_DIR="$(pwd)"

if [ ! -f "$PROJECT_DIR/go.mod" ]; then
    echo "Fehler: Bitte im Projekt-Verzeichnis ausführen"
    echo "Beispiel: cd ~/prox_watch_monitoring_system && bash installer/create_all_ledtest_files.sh"
    exit 1
fi

echo "Erstelle Verzeichnisse und Dateien für LED-Test-Tool..."

# Verzeichnisse erstellen
mkdir -p cmd/ledtest
mkdir -p internal/gpio
mkdir -p internal/neopixel
mkdir -p internal/ui

# cmd/ledtest/main.go
cat > cmd/ledtest/main.go << 'MAINEOF'
package main

import (
	"fmt"
	"os"

	"prox-watch/internal/ui"
)

func main() {
	// Terminal UI starten
	menu := ui.NewMenu()
	if err := menu.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Fehler: %v\n", err)
		os.Exit(1)
	}
}
MAINEOF

# internal/gpio/gpio.go
cat > internal/gpio/gpio.go << 'GPIOEOF'
package gpio

// GPIO-Pin-Konstanten (BCM-Nummerierung)
const (
	// NeoPixel (WS2812) - GPIO18 (Pin 12)
	NeoPixelPin = 18

	// Beeper (KY-012) - GPIO24 (Pin 18)
	BeeperPin = 24

	// Button - GPIO23 (Pin 16)
	ButtonPin = 23
)
GPIOEOF

# internal/gpio/pin.go
cat > internal/gpio/pin.go << 'PINEOF'
package gpio

// PinMode definiert den Modus eines GPIO-Pins
type PinMode int

const (
	PinModeInput PinMode = iota
	PinModeInputPullUp
	PinModeInputPullDown
	PinModeOutput
)

// PinState definiert den Zustand eines GPIO-Pins
type PinState int

const (
	PinStateLow PinState = iota
	PinStateHigh
)

// Pin ist die Schnittstelle für GPIO-Pin-Zugriff
type Pin interface {
	// High setzt den Pin auf HIGH
	High() error

	// Low setzt den Pin auf LOW
	Low() error

	// Read liest den Pin-Zustand
	Read() (PinState, error)

	// Close schließt den Pin
	Close() error
}
PINEOF

# internal/gpio/beeper.go
cat > internal/gpio/beeper.go << 'BEEPEREOF'
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
BEEPEREOF

# internal/gpio/button.go
cat > internal/gpio/button.go << 'BUTTONEOF'
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
BUTTONEOF

# internal/gpio/init.go
cat > internal/gpio/init.go << 'INITEOF'
package gpio

import (
	"sync"

	"periph.io/x/host/v3"
)

var (
	initOnce sync.Once
	initErr  error
)

// Init initialisiert periph.io (nur einmal, thread-safe)
func Init() error {
	initOnce.Do(func() {
		_, initErr = host.Init()
	})
	return initErr
}
INITEOF

# internal/gpio/init_mock.go
cat > internal/gpio/init_mock.go << 'INITMOCKEOF'
//go:build !raspberry
// +build !raspberry

package gpio

// Init ist eine No-Op für Mock-Builds
func Init() error {
	return nil
}
INITMOCKEOF

# internal/gpio/pin_raspberry.go
cat > internal/gpio/pin_raspberry.go << 'PINRASPBERRYEOF'
//go:build raspberry
// +build raspberry

package gpio

import (
	"fmt"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
)

// raspberryPin ist die Raspberry Pi Hardware-Implementierung
type raspberryPin struct {
	pin gpio.PinIO
}

// NewPin erstellt einen neuen GPIO-Pin (Raspberry Pi Hardware)
func NewPin(pinNum int, mode PinMode) (Pin, error) {
	pinName := fmt.Sprintf("GPIO%d", pinNum)
	pin := gpioreg.ByName(pinName)
	if pin == nil {
		return nil, fmt.Errorf("pin %s not found", pinName)
	}

	switch mode {
	case PinModeOutput:
		if err := pin.Out(gpio.Low); err != nil {
			return nil, fmt.Errorf("failed to set pin %s to output: %w", pinName, err)
		}
	case PinModeInput:
		if err := pin.In(gpio.Float, gpio.NoEdge); err != nil {
			return nil, fmt.Errorf("failed to set pin %s to input: %w", pinName, err)
		}
	case PinModeInputPullUp:
		if err := pin.In(gpio.PullUp, gpio.NoEdge); err != nil {
			return nil, fmt.Errorf("failed to set pin %s to input (pull-up): %w", pinName, err)
		}
	case PinModeInputPullDown:
		if err := pin.In(gpio.PullDown, gpio.NoEdge); err != nil {
			return nil, fmt.Errorf("failed to set pin %s to input (pull-down): %w", pinName, err)
		}
	default:
		return nil, fmt.Errorf("unsupported pin mode: %d", mode)
	}

	return &raspberryPin{pin: pin}, nil
}

func (p *raspberryPin) High() error {
	return p.pin.Out(gpio.High)
}

func (p *raspberryPin) Low() error {
	return p.pin.Out(gpio.Low)
}

func (p *raspberryPin) Read() (PinState, error) {
	level := p.pin.Read()
	if level == gpio.High {
		return PinStateHigh, nil
	}
	return PinStateLow, nil
}

func (p *raspberryPin) Close() error {
	// periph.io Pins müssen nicht explizit geschlossen werden
	// Aber wir setzen den Pin auf LOW für Sicherheit
	return p.pin.Out(gpio.Low)
}
PINRASPBERRYEOF

# internal/gpio/pin_mock.go
cat > internal/gpio/pin_mock.go << 'PINMOCKEOF'
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
PINMOCKEOF

# internal/neopixel/strip.go
cat > internal/neopixel/strip.go << 'STRIPEOF'
package neopixel

import (
	"fmt"
)

// Strip steuert einen WS2812 NeoPixel-Strip (8 LEDs)
type Strip struct {
	leds    int
	pin     int
	strip   *WS2812Strip
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
func NewStrip(leds int, pin int) (*Strip, error) {
	// TODO: WS2812-Bibliothek initialisieren
	// Für jetzt: Placeholder
	return &Strip{
		leds: leds,
		pin:  pin,
	}, nil
}

// SetLED setzt eine einzelne LED auf eine Farbe
func (s *Strip) SetLED(index int, color Color) error {
	if index < 0 || index >= s.leds {
		return fmt.Errorf("LED index out of range: %d (max: %d)", index, s.leds-1)
	}
	// TODO: WS2812-Bibliothek verwenden
	return nil
}

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
func (s *Strip) Close() error {
	if err := s.Clear(); err != nil {
		return err
	}
	// TODO: WS2812-Bibliothek schließen
	return nil
}

// WS2812Strip ist ein Platzhalter für die echte WS2812-Implementierung
type WS2812Strip struct {
	// Wird später mit github.com/rpi-ws281x/rpi-ws281x-go implementiert
}
STRIPEOF

# internal/ui/menu.go
cat > internal/ui/menu.go << 'MENUEOF'
package ui

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

// MenuItem repräsentiert ein Menü-Element
type MenuItem struct {
	Label string
	Action func() error
}

// Menu ist das Terminal-UI-Menü
type Menu struct {
	title  string
	items  []MenuItem
	cursor int
}

// NewMenu erstellt ein neues Menü
func NewMenu() *Menu {
	return &Menu{
		title: "LED & BEEPER TEST TOOL",
		items: []MenuItem{
			{Label: "LED Test", Action: nil}, // TODO: Implementieren
			{Label: "Beeper Test", Action: nil}, // TODO: Implementieren
			{Label: "GPIO Übersicht", Action: nil}, // TODO: Implementieren
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

// render zeichnet das Menü
func (m *Menu) render() {
	// Clear screen
	fmt.Print("\033[2J\033[H")

	// Titel
	fmt.Printf("\n  %s\n\n", m.title)

	// Menü-Items
	for i, item := range m.items {
		if i == m.cursor {
			fmt.Printf("  > %s\n", item.Label)
		} else {
			fmt.Printf("    %s\n", item.Label)
		}
	}

	fmt.Println()
}
MENUEOF

echo ""
echo "✓ Alle Dateien erstellt!"
echo ""
echo "Nächste Schritte:"
echo "  1. go get github.com/rpi-ws281x/rpi-ws281x-go"
echo "  2. go mod tidy"
echo "  3. go build -tags raspberry ./cmd/ledtest"
echo ""
