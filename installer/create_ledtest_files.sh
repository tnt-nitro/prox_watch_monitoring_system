#!/bin/bash
# Erstellt die notwendigen Dateien für das LED-Test-Tool auf dem Raspberry Pi
# 
# Verwendung: bash installer/create_ledtest_files.sh

set -euo pipefail

PROJECT_DIR="$(pwd)"

if [ ! -f "$PROJECT_DIR/go.mod" ]; then
    echo "Fehler: Bitte im Projekt-Verzeichnis ausführen"
    exit 1
fi

echo "Erstelle Verzeichnisse und Dateien für LED-Test-Tool..."

# Verzeichnisse erstellen
mkdir -p cmd/ledtest
mkdir -p internal/gpio
mkdir -p internal/neopixel
mkdir -p internal/ui

# cmd/ledtest/main.go
cat > cmd/ledtest/main.go << 'EOF'
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
EOF

# internal/gpio/gpio.go
cat > internal/gpio/gpio.go << 'EOF'
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
EOF

# internal/gpio/pin.go
cat > internal/gpio/pin.go << 'EOF'
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
EOF

echo "✓ Dateien erstellt"
echo ""
echo "Nächste Schritte:"
echo "  1. go get github.com/rpi-ws281x/rpi-ws281x-go"
echo "  2. go mod tidy"
echo "  3. go build -tags raspberry ./cmd/ledtest"
