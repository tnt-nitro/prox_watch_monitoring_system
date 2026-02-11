//go:build !raspberry
// +build !raspberry

package watcher

import (
	"fmt"
	"sync/atomic"
	"time"

	"prox-watch/internal/rules"
)

// RaspberryGPIO ist die Hardware-Implementierung des GPIO-Interfaces für Raspberry Pi.
// Phase 1.5: Verwendet periph.io für GPIO-Zugriff.
// Siehe docs/22_gpio_hardware_architecture.md für vollständige Spezifikation.
type RaspberryGPIO struct {
	greenPin  Pin
	yellowPin Pin
	redPin    Pin
	beeperPin Pin
	config    GPIOConfig
	beepActive int32 // Atomic flag: 1 = aktiv, 0 = inaktiv
}

// NewRaspberryGPIO erstellt einen neuen RaspberryGPIO.
// Initialisiert periph.io, lädt Pins über BCM-ID und konfiguriert sie als Output.
// Alle Pins werden initial auf LOW gesetzt.
func NewRaspberryGPIO(cfg GPIOConfig) (GPIO, error) {
	// Validierung: Pins müssen eindeutig sein
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid pin configuration: %w", err)
	}

	// TODO: periph.io Initialisierung
	// host.Init() - Initialisiert periph.io Host
	// gpioreg.ByName() - Lädt Pin über BCM-ID (z.B. "GPIO17")
	// pin.Out() - Konfiguriert Pin als Output
	// pin.Low() - Setzt Pin initial auf LOW

	// Platzhalter: MockPins für Tests (später durch echte periph.io Pins ersetzt)
	greenPin, err := newMockPin(cfg.LEDPinGreen)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize green pin: %w", err)
	}

	yellowPin, err := newMockPin(cfg.LEDPinYellow)
	if err != nil {
		greenPin.Close()
		return nil, fmt.Errorf("failed to initialize yellow pin: %w", err)
	}

	redPin, err := newMockPin(cfg.LEDPinRed)
	if err != nil {
		greenPin.Close()
		yellowPin.Close()
		return nil, fmt.Errorf("failed to initialize red pin: %w", err)
	}

	beeperPin, err := newMockPin(cfg.BeeperPin)
	if err != nil {
		greenPin.Close()
		yellowPin.Close()
		redPin.Close()
		return nil, fmt.Errorf("failed to initialize beeper pin: %w", err)
	}

	// Alle Pins initial auf LOW setzen
	if err := greenPin.Low(); err != nil {
		cleanupPins(greenPin, yellowPin, redPin, beeperPin)
		return nil, fmt.Errorf("failed to set green pin low: %w", err)
	}

	if err := yellowPin.Low(); err != nil {
		cleanupPins(greenPin, yellowPin, redPin, beeperPin)
		return nil, fmt.Errorf("failed to set yellow pin low: %w", err)
	}

	if err := redPin.Low(); err != nil {
		cleanupPins(greenPin, yellowPin, redPin, beeperPin)
		return nil, fmt.Errorf("failed to set red pin low: %w", err)
	}

	if err := beeperPin.Low(); err != nil {
		cleanupPins(greenPin, yellowPin, redPin, beeperPin)
		return nil, fmt.Errorf("failed to set beeper pin low: %w", err)
	}

	return &RaspberryGPIO{
		greenPin:   greenPin,
		yellowPin:  yellowPin,
		redPin:     redPin,
		beeperPin:  beeperPin,
		config:     cfg,
		beepActive: 0, // Initial: inaktiv
	}, nil
}

// SetLED setzt die LED-Farbe basierend auf Severity.
// Nur eine LED ist gleichzeitig aktiv (sauberer Wechsel).
func (r *RaspberryGPIO) SetLED(sev rules.Severity) error {
	// Alle LEDs zunächst ausschalten
	if err := r.greenPin.Low(); err != nil {
		return fmt.Errorf("failed to set green pin low: %w", err)
	}
	if err := r.yellowPin.Low(); err != nil {
		return fmt.Errorf("failed to set yellow pin low: %w", err)
	}
	if err := r.redPin.Low(); err != nil {
		return fmt.Errorf("failed to set red pin low: %w", err)
	}

	// Gewünschte LED aktivieren
	switch sev {
	case rules.SeverityInfo:
		return r.greenPin.High()
	case rules.SeverityWarn:
		return r.yellowPin.High()
	case rules.SeverityCrit:
		return r.redPin.High()
	default:
		// Unbekannte Severity → alle LEDs aus
		return nil
	}
}

// Beep aktiviert den Beeper für die konfigurierte Dauer.
// Phase 1.5: Asynchron (Goroutine), max. 1000ms, keine Leaks, Concurrency-Schutz.
// Beeper nur wenn: Eskalation zu CRIT, Zeitfenster erfüllt (wenn aktiviert), nicht bereits aktiv.
func (r *RaspberryGPIO) Beep() error {
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
		if err := r.beeperPin.High(); err != nil {
			// Fehler beim Aktivieren: Ignorieren (kein Log mit sensiblen Daten)
			return
		}

		// Warten (max. 1000ms)
		time.Sleep(duration)

		// Beeper deaktivieren
		_ = r.beeperPin.Low()
	}()

	return nil
}

// Close schließt alle GPIO-Pins und gibt Ressourcen frei.
func (r *RaspberryGPIO) Close() error {
	var errs []error

	// Alle Pins auf LOW setzen (LEDs aus, Beeper aus)
	if err := r.greenPin.Low(); err != nil {
		errs = append(errs, fmt.Errorf("green pin low: %w", err))
	}
	if err := r.yellowPin.Low(); err != nil {
		errs = append(errs, fmt.Errorf("yellow pin low: %w", err))
	}
	if err := r.redPin.Low(); err != nil {
		errs = append(errs, fmt.Errorf("red pin low: %w", err))
	}
	if err := r.beeperPin.Low(); err != nil {
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


// cleanupPins schließt alle Pins bei Fehler während Initialisierung.
func cleanupPins(pins ...Pin) {
	for _, pin := range pins {
		if pin != nil {
			_ = pin.Close()
		}
	}
}

