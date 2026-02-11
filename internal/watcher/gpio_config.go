package watcher

import (
	"errors"
	"fmt"
	"time"
)

// GPIOBackend beschreibt den Backend-Typ für GPIO.
// Phase 1.5: "noop" oder "raspberry".
type GPIOBackend string

const (
	GPIOBackendNoOp      GPIOBackend = "noop"
	GPIOBackendRaspberry GPIOBackend = "raspberry"
)

// GPIOConfig enthält die Konfiguration für GPIO (Hardware).
// Phase 1.5: Erweitert für Raspberry Pi Hardware-GPIO.
type GPIOConfig struct {
	Enabled bool

	// Backend wählt die Implementierung aus: "noop" oder "raspberry".
	Backend GPIOBackend

	// Pins (BCM-Nummerierung)
	LEDPinGreen  int
	LEDPinYellow int
	LEDPinRed    int
	BeeperPin    int

	// Beeper-Konfiguration
	BeeperDayOnly bool
	// BeeperWindowStart/BeeperWindowEnd im Format "HH:MM" (lokale Zeit)
	BeeperWindowStart string
	BeeperWindowEnd   string
	// Maximale Beep-Dauer in Millisekunden (z.B. 1000)
	BeeperMaxDurationMs int
}

// DefaultGPIOConfig gibt eine Standard-GPIO-Konfiguration zurück.
// Phase 1.5: Pin-Layout für Raspberry Pi.
func DefaultGPIOConfig() GPIOConfig {
	return GPIOConfig{
		Enabled:            false,              // Phase 1: Default false (NoOp)
		Backend:            GPIOBackendNoOp,    // Standard: NoOp
		LEDPinGreen:        17,                 // BCM Pin 17
		LEDPinYellow:       27,                 // BCM Pin 27
		LEDPinRed:          22,                 // BCM Pin 22
		BeeperPin:          23,                 // BCM Pin 23
		BeeperDayOnly:      true,
		BeeperWindowStart:  "08:00",
		BeeperWindowEnd:    "22:00",
		BeeperMaxDurationMs: 1000,
	}
}

// Validate prüft die GPIO-Konfiguration auf Konsistenz.
// Guards:
// - Wenn Enabled=false → immer NoOp (Backend wird ignoriert)
// - Wenn Backend=raspberry:
//   - Pins müssen eindeutig sein
//   - BeeperPin darf nicht LED-Pin sein
//   - BeeperWindow muss gültig sein (HH:MM)
func (c GPIOConfig) Validate() error {
	// Disabled → immer gültig, erzwingt NoOp zur Laufzeit.
	if !c.Enabled {
		return nil
	}

	// Enabled: Backend muss gesetzt sein, Default = raspberry wenn leer.
	if c.Backend == "" {
		c.Backend = GPIOBackendRaspberry
	}

	switch c.Backend {
	case GPIOBackendNoOp:
		// Immer gültig – NoOp benötigt keine Pins.
		return nil
	case GPIOBackendRaspberry:
		// Pins müssen sinnvoll gesetzt sein (>0).
		if c.LEDPinGreen <= 0 || c.LEDPinYellow <= 0 || c.LEDPinRed <= 0 || c.BeeperPin <= 0 {
			return errors.New("gpio: all pins must be > 0 for raspberry backend")
		}

		// Pins müssen eindeutig sein.
		if c.LEDPinGreen == c.LEDPinYellow ||
			c.LEDPinGreen == c.LEDPinRed ||
			c.LEDPinYellow == c.LEDPinRed {
			return errors.New("gpio: LED pins must be distinct for raspberry backend")
		}

		// BeeperPin darf nicht LED-Pin sein.
		if c.BeeperPin == c.LEDPinGreen ||
			c.BeeperPin == c.LEDPinYellow ||
			c.BeeperPin == c.LEDPinRed {
			return errors.New("gpio: beeper pin must not equal any LED pin for raspberry backend")
		}

		// Beeper-Fenster validieren (optional, aber empfohlen).
		if err := validateTimeHM(c.BeeperWindowStart); err != nil {
			return fmt.Errorf("gpio: invalid beeper_window_start: %w", err)
		}
		if err := validateTimeHM(c.BeeperWindowEnd); err != nil {
			return fmt.Errorf("gpio: invalid beeper_window_end: %w", err)
		}

		// BeeperMaxDurationMs muss positiv und sinnvoll sein.
		if c.BeeperMaxDurationMs <= 0 || c.BeeperMaxDurationMs > 10000 {
			return errors.New("gpio: beeper_max_ms must be between 1 and 10000")
		}

		return nil
	default:
		return fmt.Errorf("gpio: unknown backend %q", c.Backend)
	}
}

// validateTimeHM prüft, ob eine Zeit im Format "HH:MM" gültig ist.
func validateTimeHM(v string) error {
	if v == "" {
		return errors.New("empty time")
	}
	_, err := time.Parse("15:04", v)
	return err
}

// NewGPIOFromConfig erstellt ein GPIO-Backend basierend auf der Konfiguration.
// Regeln:
// - Enabled=false → immer NoOpGPIO
// - Backend=noop → NoOpGPIO
// - Backend=raspberry → (Platzhalter, Implementierung folgt in Phase 1.5)
func NewGPIOFromConfig(cfg GPIOConfig) (GPIO, error) {
	// Disabled → NoOp
	if !cfg.Enabled {
		return NewNoOpGPIO(), nil
	}

	backend := cfg.Backend
	if backend == "" {
		backend = GPIOBackendRaspberry
	}

	switch backend {
	case GPIOBackendNoOp:
		return NewNoOpGPIO(), nil
	case GPIOBackendRaspberry:
		return NewRaspberryGPIO(cfg)
	default:
		return nil, fmt.Errorf("gpio: unknown backend %q", backend)
	}
}
