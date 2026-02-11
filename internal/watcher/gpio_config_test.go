package watcher

import "testing"

func TestDefaultGPIOConfig(t *testing.T) {
	cfg := DefaultGPIOConfig()

	if cfg.Enabled {
		t.Errorf("Expected Enabled=false by default, got true")
	}
	if cfg.Backend != GPIOBackendNoOp {
		t.Errorf("Expected Backend=%q by default, got %q", GPIOBackendNoOp, cfg.Backend)
	}
	if cfg.LEDPinGreen != 17 || cfg.LEDPinYellow != 27 || cfg.LEDPinRed != 22 || cfg.BeeperPin != 23 {
		t.Errorf("Unexpected default pin layout: green=%d, yellow=%d, red=%d, beeper=%d",
			cfg.LEDPinGreen, cfg.LEDPinYellow, cfg.LEDPinRed, cfg.BeeperPin)
	}
	if !cfg.BeeperDayOnly {
		t.Errorf("Expected BeeperDayOnly=true by default")
	}
	if cfg.BeeperWindowStart != "08:00" || cfg.BeeperWindowEnd != "22:00" {
		t.Errorf("Unexpected default beeper window: %s-%s", cfg.BeeperWindowStart, cfg.BeeperWindowEnd)
	}
	if cfg.BeeperMaxDurationMs != 1000 {
		t.Errorf("Expected BeeperMaxDurationMs=1000, got %d", cfg.BeeperMaxDurationMs)
	}
}

func TestGPIOConfig_Validate_DisabledForcesNoOp(t *testing.T) {
	cfg := DefaultGPIOConfig()
	cfg.Enabled = false
	cfg.Backend = GPIOBackendRaspberry

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() for disabled config returned error: %v", err)
	}

	gpio, err := NewGPIOFromConfig(cfg)
	if err != nil {
		t.Fatalf("NewGPIOFromConfig() returned error for disabled config: %v", err)
	}
	if _, ok := gpio.(*NoOpGPIO); !ok {
		t.Errorf("Expected NoOpGPIO when Enabled=false, got %T", gpio)
	}
}

func TestGPIOConfig_Validate_RaspberryPinsDistinct(t *testing.T) {
	cfg := DefaultGPIOConfig()
	cfg.Enabled = true
	cfg.Backend = GPIOBackendRaspberry

	// Gültige Konfiguration
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() for valid raspberry config returned error: %v", err)
	}

	// Duplikate bei LED-Pins → Fehler
	cfgBad := cfg
	cfgBad.LEDPinYellow = cfgBad.LEDPinGreen
	if err := cfgBad.Validate(); err == nil {
		t.Errorf("Expected error for duplicate LED pins, got nil")
	}

	// BeeperPin gleich LED-Pin → Fehler
	cfgBad = cfg
	cfgBad.BeeperPin = cfgBad.LEDPinRed
	if err := cfgBad.Validate(); err == nil {
		t.Errorf("Expected error for beeper pin equal to LED pin, got nil")
	}
}

func TestGPIOConfig_Validate_InvalidBackend(t *testing.T) {
	cfg := DefaultGPIOConfig()
	cfg.Enabled = true
	cfg.Backend = GPIOBackend("invalid")

	if err := cfg.Validate(); err == nil {
		t.Errorf("Expected error for invalid backend, got nil")
	}
}

func TestGPIOConfig_Validate_InvalidBeeperWindow(t *testing.T) {
	cfg := DefaultGPIOConfig()
	cfg.Enabled = true
	cfg.Backend = GPIOBackendRaspberry

	cfg.BeeperWindowStart = "25:00"
	if err := cfg.Validate(); err == nil {
		t.Errorf("Expected error for invalid BeeperWindowStart, got nil")
	}

	cfg = DefaultGPIOConfig()
	cfg.Enabled = true
	cfg.Backend = GPIOBackendRaspberry
	cfg.BeeperWindowEnd = "99:99"
	if err := cfg.Validate(); err == nil {
		t.Errorf("Expected error for invalid BeeperWindowEnd, got nil")
	}
}

func TestGPIOConfig_Validate_InvalidBeeperDuration(t *testing.T) {
	cfg := DefaultGPIOConfig()
	cfg.Enabled = true
	cfg.Backend = GPIOBackendRaspberry
	cfg.BeeperMaxDurationMs = 0

	if err := cfg.Validate(); err == nil {
		t.Errorf("Expected error for BeeperMaxDurationMs <= 0, got nil")
	}

	cfg.BeeperMaxDurationMs = 20000
	if err := cfg.Validate(); err == nil {
		t.Errorf("Expected error for BeeperMaxDurationMs > 10000, got nil")
	}
}

func TestNewGPIOFromConfig_BackendSelection(t *testing.T) {
	// Enabled=false → NoOp
	cfg := DefaultGPIOConfig()
	cfg.Enabled = false
	gpio, err := NewGPIOFromConfig(cfg)
	if err != nil {
		t.Fatalf("NewGPIOFromConfig() error for disabled config: %v", err)
	}
	if _, ok := gpio.(*NoOpGPIO); !ok {
		t.Errorf("Expected NoOpGPIO when Enabled=false, got %T", gpio)
	}

	// Enabled=true + backend=noop → NoOp
	cfg = DefaultGPIOConfig()
	cfg.Enabled = true
	cfg.Backend = GPIOBackendNoOp
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() failed for noop backend: %v", err)
	}
	gpio, err = NewGPIOFromConfig(cfg)
	if err != nil {
		t.Fatalf("NewGPIOFromConfig() error for noop backend: %v", err)
	}
	if _, ok := gpio.(*NoOpGPIO); !ok {
		t.Errorf("Expected NoOpGPIO for backend=noop, got %T", gpio)
	}

	// Enabled=true + backend=raspberry → aktuell nicht implementiert (Fehler erwartet)
	cfg = DefaultGPIOConfig()
	cfg.Enabled = true
	cfg.Backend = GPIOBackendRaspberry
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() failed for raspberry backend: %v", err)
	}
	if _, err := NewGPIOFromConfig(cfg); err == nil {
		t.Errorf("Expected error for raspberry backend (not implemented yet), got nil")
	}
}

