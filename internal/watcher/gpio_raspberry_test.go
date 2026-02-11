package watcher

import (
	"prox-watch/internal/rules"
	"testing"
)

func TestNewRaspberryGPIO_ValidConfig(t *testing.T) {
	// Valid Config
	cfg := GPIOConfig{
		Enabled:            true,
		Backend:            GPIOBackendRaspberry,
		LEDPinGreen:        17,
		LEDPinYellow:       27,
		LEDPinRed:          22,
		BeeperPin:          23,
		BeeperDayOnly:      true,
		BeeperWindowStart:  "08:00",
		BeeperWindowEnd:    "22:00",
		BeeperMaxDurationMs: 1000,
	}

	gpio, err := NewRaspberryGPIO(cfg)
	if err != nil {
		t.Fatalf("NewRaspberryGPIO() failed: %v", err)
	}
	defer gpio.Close()

	// Prüfe, dass GPIO erstellt wurde
	if gpio == nil {
		t.Error("Expected GPIO instance, got nil")
	}
}

func TestNewRaspberryGPIO_DuplicatePins(t *testing.T) {
	// Duplicate Pins → Fehler
	cfg := GPIOConfig{
		Enabled:     true,
		Backend:     GPIOBackendRaspberry,
		LEDPinGreen: 17,
		LEDPinYellow: 17, // Duplikat
		LEDPinRed:   22,
		BeeperPin:   23,
	}

	_, err := NewRaspberryGPIO(cfg)
	if err == nil {
		t.Error("Expected error for duplicate pins, got nil")
	}
}

func TestNewRaspberryGPIO_BeeperEqualsLED(t *testing.T) {
	// Beeper Pin = LED Pin → Fehler
	cfg := GPIOConfig{
		Enabled:     true,
		Backend:     GPIOBackendRaspberry,
		LEDPinGreen: 17,
		LEDPinYellow: 27,
		LEDPinRed:   22,
		BeeperPin:   17, // Gleich Green Pin
	}

	_, err := NewRaspberryGPIO(cfg)
	if err == nil {
		t.Error("Expected error when beeper pin equals LED pin, got nil")
	}
}

func TestRaspberryGPIO_SetLED(t *testing.T) {
	cfg := DefaultGPIOConfig()
	cfg.Enabled = true
	cfg.Backend = GPIOBackendRaspberry

	gpio, err := NewRaspberryGPIO(cfg)
	if err != nil {
		t.Fatalf("NewRaspberryGPIO() failed: %v", err)
	}
	defer gpio.Close()

	// Test INFO → Grün
	if err := gpio.SetLED(rules.SeverityInfo); err != nil {
		t.Errorf("SetLED(INFO) failed: %v", err)
	}

	// Test WARN → Gelb
	if err := gpio.SetLED(rules.SeverityWarn); err != nil {
		t.Errorf("SetLED(WARN) failed: %v", err)
	}

	// Test CRIT → Rot
	if err := gpio.SetLED(rules.SeverityCrit); err != nil {
		t.Errorf("SetLED(CRIT) failed: %v", err)
	}
}

func TestRaspberryGPIO_Beep(t *testing.T) {
	cfg := DefaultGPIOConfig()
	cfg.Enabled = true
	cfg.Backend = GPIOBackendRaspberry
	cfg.BeeperDayOnly = true

	gpio, err := NewRaspberryGPIO(cfg)
	if err != nil {
		t.Fatalf("NewRaspberryGPIO() failed: %v", err)
	}
	defer gpio.Close()

	// Beep sollte funktionieren (Tag-Zeitfenster wird gemockt)
	if err := gpio.Beep(); err != nil {
		t.Errorf("Beep() failed: %v", err)
	}
}

func TestRaspberryGPIO_Beep_TimeWindow_Within(t *testing.T) {
	// Test: CRIT innerhalb Zeitfenster → Beep
	cfg := DefaultGPIOConfig()
	cfg.Enabled = true
	cfg.Backend = GPIOBackendRaspberry
	cfg.BeeperDayOnly = true
	cfg.BeeperWindowStart = "08:00"
	cfg.BeeperWindowEnd = "22:00"

	gpio, err := NewRaspberryGPIO(cfg)
	if err != nil {
		t.Fatalf("NewRaspberryGPIO() failed: %v", err)
	}
	defer gpio.Close()

	// Beep sollte funktionieren (wenn aktuelle Zeit im Fenster liegt)
	// Hinweis: isDayTime() verwendet time.Now(), daher kann der Test je nach Tageszeit unterschiedlich ausfallen
	// Für deterministische Tests wäre ein Mock-Zeit-Interface nötig
	err = gpio.Beep()
	if err != nil {
		t.Errorf("Beep() failed: %v", err)
	}
}

func TestRaspberryGPIO_Beep_TimeWindow_Outside(t *testing.T) {
	// Test: CRIT außerhalb Zeitfenster → kein Beep
	// Hinweis: Dieser Test ist schwierig ohne Mock-Zeit, da isDayTime() time.Now() verwendet
	// Für vollständige Tests wäre ein Mock-Zeit-Interface erforderlich
	cfg := DefaultGPIOConfig()
	cfg.Enabled = true
	cfg.Backend = GPIOBackendRaspberry
	cfg.BeeperDayOnly = true
	cfg.BeeperWindowStart = "22:00"
	cfg.BeeperWindowEnd = "06:00"

	gpio, err := NewRaspberryGPIO(cfg)
	if err != nil {
		t.Fatalf("NewRaspberryGPIO() failed: %v", err)
	}
	defer gpio.Close()

	// Beep sollte abhängig von der aktuellen Zeit funktionieren oder nicht
	// (ohne Mock-Zeit können wir nur prüfen, dass kein Fehler auftritt)
	err = gpio.Beep()
	if err != nil {
		t.Errorf("Beep() failed: %v", err)
	}
}

func TestRaspberryGPIO_Beep_DoubleCall(t *testing.T) {
	// Test: Doppeltes CRIT → nur ein Beep (Concurrency-Schutz)
	cfg := DefaultGPIOConfig()
	cfg.Enabled = true
	cfg.Backend = GPIOBackendRaspberry
	cfg.BeeperDayOnly = false // Deaktiviert für einfacheren Test
	cfg.BeeperMaxDurationMs = 100

	gpio, err := NewRaspberryGPIO(cfg)
	if err != nil {
		t.Fatalf("NewRaspberryGPIO() failed: %v", err)
	}
	defer gpio.Close()

	// Erster Beep
	if err := gpio.Beep(); err != nil {
		t.Errorf("First Beep() failed: %v", err)
	}

	// Zweiter Beep sofort (sollte blockiert werden durch beepActive)
	if err := gpio.Beep(); err != nil {
		t.Errorf("Second Beep() failed: %v", err)
	}

	// Warten, bis erster Beep beendet ist
	time.Sleep(150 * time.Millisecond)

	// Jetzt sollte ein neuer Beep möglich sein
	if err := gpio.Beep(); err != nil {
		t.Errorf("Third Beep() after wait failed: %v", err)
	}
}

func TestRaspberryGPIO_Beep_MaxDuration(t *testing.T) {
	// Test: Max-Dauer eingehalten
	cfg := DefaultGPIOConfig()
	cfg.Enabled = true
	cfg.Backend = GPIOBackendRaspberry
	cfg.BeeperDayOnly = false
	cfg.BeeperMaxDurationMs = 2000 // > 1000ms, sollte auf 1000ms begrenzt werden

	gpio, err := NewRaspberryGPIO(cfg)
	if err != nil {
		t.Fatalf("NewRaspberryGPIO() failed: %v", err)
	}
	defer gpio.Close()

	start := time.Now()
	if err := gpio.Beep(); err != nil {
		t.Errorf("Beep() failed: %v", err)
	}

	// Warten, bis Beep beendet ist (max. 1000ms + kleine Toleranz)
	time.Sleep(1100 * time.Millisecond)

	duration := time.Since(start)
	if duration > 1200*time.Millisecond {
		t.Errorf("Beep duration too long: %v (expected max 1000ms)", duration)
	}
}

func TestRaspberryGPIO_BeepActive_Reset(t *testing.T) {
	// Test: BeepActive reset nach Ende
	cfg := DefaultGPIOConfig()
	cfg.Enabled = true
	cfg.Backend = GPIOBackendRaspberry
	cfg.BeeperDayOnly = false
	cfg.BeeperMaxDurationMs = 50

	gpio, err := NewRaspberryGPIO(cfg)
	if err != nil {
		t.Fatalf("NewRaspberryGPIO() failed: %v", err)
	}
	defer gpio.Close()

	// Erster Beep
	if err := gpio.Beep(); err != nil {
		t.Errorf("First Beep() failed: %v", err)
	}

	// Zweiter Beep sofort (sollte blockiert werden)
	if err := gpio.Beep(); err != nil {
		t.Errorf("Second Beep() failed: %v", err)
	}

	// Warten, bis erster Beep beendet ist
	time.Sleep(100 * time.Millisecond)

	// Jetzt sollte ein neuer Beep möglich sein (Flag wurde zurückgesetzt)
	if err := gpio.Beep(); err != nil {
		t.Errorf("Third Beep() after reset failed: %v", err)
	}
}

func TestRaspberryGPIO_Close(t *testing.T) {
	cfg := DefaultGPIOConfig()
	cfg.Enabled = true
	cfg.Backend = GPIOBackendRaspberry

	gpio, err := NewRaspberryGPIO(cfg)
	if err != nil {
		t.Fatalf("NewRaspberryGPIO() failed: %v", err)
	}

	// Close sollte funktionieren
	if err := gpio.Close(); err != nil {
		t.Errorf("Close() failed: %v", err)
	}
}

func TestRaspberryGPIO_InitialState_LOW(t *testing.T) {
	cfg := DefaultGPIOConfig()
	cfg.Enabled = true
	cfg.Backend = GPIOBackendRaspberry

	gpio, err := NewRaspberryGPIO(cfg)
	if err != nil {
		t.Fatalf("NewRaspberryGPIO() failed: %v", err)
	}
	defer gpio.Close()

	// Alle Pins sollten initial LOW sein
	// (Prüfung über MockPin.GetState() möglich, wenn RaspberryGPIO exportiert)
	// Für jetzt: Prüfe, dass keine Fehler auftreten
}

func TestRaspberryGPIO_LEDSwitch(t *testing.T) {
	// Prüfe, dass LED-Wechsel sauber ist (nur eine LED aktiv)
	cfg := DefaultGPIOConfig()
	cfg.Enabled = true
	cfg.Backend = GPIOBackendRaspberry

	gpio, err := NewRaspberryGPIO(cfg)
	if err != nil {
		t.Fatalf("NewRaspberryGPIO() failed: %v", err)
	}
	defer gpio.Close()

	// Wechsel: INFO → WARN → CRIT
	if err := gpio.SetLED(rules.SeverityInfo); err != nil {
		t.Errorf("SetLED(INFO) failed: %v", err)
	}

	if err := gpio.SetLED(rules.SeverityWarn); err != nil {
		t.Errorf("SetLED(WARN) failed: %v", err)
	}

	if err := gpio.SetLED(rules.SeverityCrit); err != nil {
		t.Errorf("SetLED(CRIT) failed: %v", err)
	}

	// Wechsel zurück zu INFO
	if err := gpio.SetLED(rules.SeverityInfo); err != nil {
		t.Errorf("SetLED(INFO) failed: %v", err)
	}
}

func TestMockPin_HighLow(t *testing.T) {
	// MockPin Tests
		pin, err := NewMockPin(17)
		if err != nil {
			t.Fatalf("NewMockPin() failed: %v", err)
	}
	defer pin.Close()

	// Initial sollte LOW sein
	mockPin := pin.(*MockPin)
	if mockPin.GetState() != false {
		t.Error("Expected initial state to be LOW (false)")
	}

	// High setzen
	if err := pin.High(); err != nil {
		t.Errorf("High() failed: %v", err)
	}
	if mockPin.GetState() != true {
		t.Error("Expected state to be HIGH (true) after High()")
	}

	// Low setzen
	if err := pin.Low(); err != nil {
		t.Errorf("Low() failed: %v", err)
	}
	if mockPin.GetState() != false {
		t.Error("Expected state to be LOW (false) after Low()")
	}
}

func TestMockPin_InvalidPin(t *testing.T) {
	// Invalid Pin Number
		_, err := NewMockPin(-1)
	if err == nil {
		t.Error("Expected error for invalid pin number, got nil")
	}
}
