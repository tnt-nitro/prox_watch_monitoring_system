//go:build raspberry
// +build raspberry

package watcher

import (
	"fmt"
	"sync/atomic"
	"time"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/host/v3"
	"prox-watch/internal/rules"
)

// NewRaspberryGPIO erstellt einen neuen RaspberryGPIO mit echter Hardware-Anbindung.
// Phase 1.5: Verwendet periph.io für GPIO-Zugriff auf Raspberry Pi.
// Siehe docs/22_gpio_hardware_architecture.md für vollständige Spezifikation.
func NewRaspberryGPIO(cfg GPIOConfig) (GPIO, error) {
	// Validierung: Pins müssen eindeutig sein
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid pin configuration: %w", err)
	}

	// periph.io Host initialisieren
	if _, err := host.Init(); err != nil {
		return nil, fmt.Errorf("failed to initialize periph.io host: %w", err)
	}

	// Pins über BCM-ID laden
	greenPin, err := loadPin(cfg.LEDPinGreen, "green")
	if err != nil {
		return nil, fmt.Errorf("failed to load green pin: %w", err)
	}

	yellowPin, err := loadPin(cfg.LEDPinYellow, "yellow")
	if err != nil {
		greenPin.Close()
		return nil, fmt.Errorf("failed to load yellow pin: %w", err)
	}

	redPin, err := loadPin(cfg.LEDPinRed, "red")
	if err != nil {
		greenPin.Close()
		yellowPin.Close()
		return nil, fmt.Errorf("failed to load red pin: %w", err)
	}

	beeperPin, err := loadPin(cfg.BeeperPin, "beeper")
	if err != nil {
		greenPin.Close()
		yellowPin.Close()
		redPin.Close()
		return nil, fmt.Errorf("failed to load beeper pin: %w", err)
	}

	// Alle Pins initial auf LOW setzen
	if err := greenPin.Out(gpio.Low); err != nil {
		cleanupPinsHW(greenPin, yellowPin, redPin, beeperPin)
		return nil, fmt.Errorf("failed to set green pin low: %w", err)
	}

	if err := yellowPin.Out(gpio.Low); err != nil {
		cleanupPinsHW(greenPin, yellowPin, redPin, beeperPin)
		return nil, fmt.Errorf("failed to set yellow pin low: %w", err)
	}

	if err := redPin.Out(gpio.Low); err != nil {
		cleanupPinsHW(greenPin, yellowPin, redPin, beeperPin)
		return nil, fmt.Errorf("failed to set red pin low: %w", err)
	}

	if err := beeperPin.Out(gpio.Low); err != nil {
		cleanupPinsHW(greenPin, yellowPin, redPin, beeperPin)
		return nil, fmt.Errorf("failed to set beeper pin low: %w", err)
	}

	return &raspberryGPIOHW{
		greenPin:   greenPin,
		yellowPin:  yellowPin,
		redPin:     redPin,
		beeperPin:  beeperPin,
		config:     cfg,
		beepActive: 0, // Initial: inaktiv
	}, nil
}

// raspberryGPIOHW ist die Hardware-Implementierung des GPIO-Interfaces für Raspberry Pi.
type raspberryGPIOHW struct {
	greenPin  gpio.PinIO
	yellowPin gpio.PinIO
	redPin    gpio.PinIO
	beeperPin gpio.PinIO
	config    GPIOConfig
	beepActive int32 // Atomic flag: 1 = aktiv, 0 = inaktiv
}

// SetLED setzt die LED-Farbe basierend auf Severity.
// Nur eine LED ist gleichzeitig aktiv (sauberer Wechsel).
func (r *raspberryGPIOHW) SetLED(sev rules.Severity) error {
	// Alle LEDs zunächst ausschalten
	if err := r.greenPin.Out(gpio.Low); err != nil {
		return fmt.Errorf("failed to set green pin low: %w", err)
	}
	if err := r.yellowPin.Out(gpio.Low); err != nil {
		return fmt.Errorf("failed to set yellow pin low: %w", err)
	}
	if err := r.redPin.Out(gpio.Low); err != nil {
		return fmt.Errorf("failed to set red pin low: %w", err)
	}

	// Gewünschte LED aktivieren
	switch sev {
	case rules.SeverityInfo:
		return r.greenPin.Out(gpio.High)
	case rules.SeverityWarn:
		return r.yellowPin.Out(gpio.High)
	case rules.SeverityCrit:
		return r.redPin.Out(gpio.High)
	default:
		// Unbekannte Severity → alle LEDs aus
		return nil
	}
}

// Beep aktiviert den Beeper für die konfigurierte Dauer.
// Phase 1.5: Asynchron (Goroutine), max. 1000ms, keine Leaks, Concurrency-Schutz.
// Beeper nur wenn: Eskalation zu CRIT, Zeitfenster erfüllt (wenn aktiviert), nicht bereits aktiv.
func (r *raspberryGPIOHW) Beep() error {
	// Prüfe Tag-Zeitfenster (wenn aktiviert)
	if r.config.BeeperDayOnly {
		if !isDayTime(r.config.BeeperWindowStart, r.config.BeeperWindowEnd) {
			// Nacht: Kein Beep
			return nil
		}
	}

	// Concurrency-Schutz: Prüfe, ob Beeper bereits aktiv ist
	if !atomic.CompareAndSwapInt32(&r.beepActive, 0, 1) {
		// Beeper bereits aktiv → kein neues Beep
		return nil
	}

	// Beeper-Dauer (max. 1000ms)
	duration := time.Duration(r.config.BeeperMaxDurationMs) * time.Millisecond
	if duration > 1000*time.Millisecond {
		duration = 1000 * time.Millisecond
	}

	// Goroutine für Beeper (nicht blockierend)
	go func() {
		// Flag am Ende zurücksetzen (garantiert)
		defer atomic.StoreInt32(&r.beepActive, 0)

		// Beeper aktivieren
		if err := r.beeperPin.Out(gpio.High); err != nil {
			// Fehler beim Aktivieren: Ignorieren (kein Log mit sensiblen Daten)
			return
		}

		// Warten (max. 1000ms)
		time.Sleep(duration)

		// Beeper deaktivieren
		_ = r.beeperPin.Out(gpio.Low)
	}()

	return nil
}

// Close schließt alle GPIO-Pins und gibt Ressourcen frei.
func (r *raspberryGPIOHW) Close() error {
	var errs []error

	// Alle Pins auf LOW setzen (LEDs aus, Beeper aus)
	if err := r.greenPin.Out(gpio.Low); err != nil {
		errs = append(errs, fmt.Errorf("green pin low: %w", err))
	}
	if err := r.yellowPin.Out(gpio.Low); err != nil {
		errs = append(errs, fmt.Errorf("yellow pin low: %w", err))
	}
	if err := r.redPin.Out(gpio.Low); err != nil {
		errs = append(errs, fmt.Errorf("red pin low: %w", err))
	}
	if err := r.beeperPin.Out(gpio.Low); err != nil {
		errs = append(errs, fmt.Errorf("beeper pin low: %w", err))
	}

	// Alle Pins schließen
	if err := r.greenPin.Close(); err != nil {
		errs = append(errs, fmt.Errorf("green pin close: %w", err))
	}
	if err := r.yellowPin.Close(); err != nil {
		errs = append(errs, fmt.Errorf("yellow pin close: %w", err))
	}
	if err := r.redPin.Close(); err != nil {
		errs = append(errs, fmt.Errorf("red pin close: %w", err))
	}
	if err := r.beeperPin.Close(); err != nil {
		errs = append(errs, fmt.Errorf("beeper pin close: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors during close: %v", errs)
	}

	return nil
}

// loadPin lädt einen GPIO-Pin über BCM-ID und konfiguriert ihn als Output.
func loadPin(bcmPin int, name string) (gpio.PinIO, error) {
	// BCM-Pin-Name: "GPIO17", "GPIO27", etc.
	pinName := fmt.Sprintf("GPIO%d", bcmPin)

	pin := gpioreg.ByName(pinName)
	if pin == nil {
		return nil, fmt.Errorf("pin %s (%s) not found", pinName, name)
	}

	// Pin als Output konfigurieren (initial LOW)
	// Out() konfiguriert den Pin und setzt ihn auf den angegebenen Level
	if err := pin.Out(gpio.Low); err != nil {
		return nil, fmt.Errorf("failed to configure pin %s as output: %w", pinName, err)
	}

	return pin, nil
}

// cleanupPinsHW schließt alle Pins bei Fehler während Initialisierung.
func cleanupPinsHW(pins ...gpio.PinIO) {
	for _, pin := range pins {
		if pin != nil {
			_ = pin.Close()
		}
	}
}
