package watcher

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// WatcherConfig enthält die vollständige Konfiguration für den Watcher.
// Phase 2: Erweitert um cooldown_seconds.
// Phase 3: Erweitert um PowerCycleConfig.
type WatcherConfig struct {
	Watcher    WatcherSection   `yaml:"watcher"`
	Target     TargetSection    `yaml:"target"`
	Thresholds ThresholdsSection `yaml:"thresholds"`
	Push       PushSection      `yaml:"push"`
	GPIO       GPIOConfig       `yaml:"gpio"`
	PowerCycle PowerCycleConfig `yaml:"powercycle"` // Phase 3: Power-Cycle-Konfiguration
	Security   SecuritySection  `yaml:"security"`
}

// WatcherSection enthält Watcher-spezifische Einstellungen.
type WatcherSection struct {
	IntervalSeconds int `yaml:"interval_seconds"`
	CooldownSeconds int `yaml:"cooldown_seconds"` // Phase 2: Cooldown in Sekunden
}

// TargetSection enthält die Ziel-Host-Konfiguration.
type TargetSection struct {
	Mode          string `yaml:"mode"`
	Host          string `yaml:"host"`
	Port          int    `yaml:"port"`
	TimeoutSeconds int   `yaml:"timeout_seconds"`
}

// ThresholdsSection enthält die Schwellwerte für Severity.
type ThresholdsSection struct {
	Warn int `yaml:"warn"`
	Crit int `yaml:"crit"`
}

// PushSection enthält Push-Benachrichtigungs-Konfiguration.
type PushSection struct {
	Enabled bool            `yaml:"enabled"`
	Adapter string          `yaml:"adapter"`
	Topics  PushTopicsSection `yaml:"topics"`
}

// PushTopicsSection enthält die Topics für Push-Benachrichtigungen.
type PushTopicsSection struct {
	Warn string `yaml:"warn"`
	Crit string `yaml:"crit"`
}

// SecuritySection enthält Sicherheitseinstellungen.
type SecuritySection struct {
	BlockIPLiterals         bool `yaml:"block_ip_literals"`
	RequireManualPowerCycle bool `yaml:"require_manual_powercycle"`
}

// DefaultWatcherConfig gibt eine Standard-Konfiguration zurück.
func DefaultWatcherConfig() WatcherConfig {
	return WatcherConfig{
		Watcher: WatcherSection{
			IntervalSeconds: 30,
			CooldownSeconds: 600, // Default: 10 Minuten
		},
		Target: TargetSection{
			Mode:          "ping+https",
			Host:          "PLACEHOLDER",
			Port:          8006,
			TimeoutSeconds: 5,
		},
		Thresholds: ThresholdsSection{
			Warn: 3,
			Crit: 10,
		},
		Push: PushSection{
			Enabled: true,
			Adapter: "ntfy",
			Topics: PushTopicsSection{
				Warn: "prox-watch-warn",
				Crit: "prox-watch-crit",
			},
		},
		GPIO:       DefaultGPIOConfig(),
		PowerCycle: DefaultPowerCycleConfig(), // Phase 3: Default Power-Cycle-Konfiguration
		Security: SecuritySection{
			BlockIPLiterals:         true,
			RequireManualPowerCycle: true,
		},
	}
}

// Validate prüft die Watcher-Konfiguration auf Konsistenz.
// Phase 2: Erweitert um Cooldown-Validierung.
func (c WatcherConfig) Validate() error {
	// Watcher-Section
	if c.Watcher.IntervalSeconds < 10 || c.Watcher.IntervalSeconds > 300 {
		return errors.New("watcher.interval_seconds must be between 10 and 300")
	}

	// Phase 2: Cooldown-Validierung
	if c.Watcher.CooldownSeconds < 0 {
		return errors.New("watcher.cooldown_seconds must be >= 0")
	}
	// Max: 86400 Sekunden (24 Stunden)
	if c.Watcher.CooldownSeconds > 86400 {
		return fmt.Errorf("watcher.cooldown_seconds must be <= 86400 (24 hours), got %d", c.Watcher.CooldownSeconds)
	}

	// Target-Section
	if c.Target.Mode != "ping" && c.Target.Mode != "https" && c.Target.Mode != "ping+https" {
		return fmt.Errorf("target.mode must be one of: ping, https, ping+https, got %q. Note: IP addresses belong in target.host, not target.mode", c.Target.Mode)
	}
	if c.Target.Port < 1 || c.Target.Port > 65535 {
		return errors.New("target.port must be between 1 and 65535")
	}
	if c.Target.TimeoutSeconds < 1 || c.Target.TimeoutSeconds > 30 {
		return errors.New("target.timeout_seconds must be between 1 and 30")
	}

	// Thresholds-Section
	if c.Thresholds.Warn < 1 {
		return errors.New("thresholds.warn must be >= 1")
	}
	if c.Thresholds.Crit < 1 {
		return errors.New("thresholds.crit must be >= 1")
	}
	if c.Thresholds.Warn >= c.Thresholds.Crit {
		return errors.New("thresholds.warn must be < thresholds.crit")
	}

	// GPIO-Validierung
	if err := c.GPIO.Validate(); err != nil {
		return fmt.Errorf("gpio validation failed: %w", err)
	}

	// Phase 3: Power-Cycle-Validierung
	if err := c.PowerCycle.Validate(); err != nil {
		return fmt.Errorf("powercycle validation failed: %w", err)
	}

	return nil
}

// LoadWatcherConfig lädt die Watcher-Konfiguration aus einer YAML-Datei.
// Falls die Datei nicht existiert, wird eine Default-Konfiguration zurückgegeben.
func LoadWatcherConfig(path string) (*WatcherConfig, error) {
	// Start mit Defaults
	cfg := DefaultWatcherConfig()

	// Versuche Datei zu laden
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Datei existiert nicht → Defaults zurückgeben
			return &cfg, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// YAML parsen
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Debug: Zeige geparste Werte bei Validierungsfehler (nur für target.mode)
	// Dies hilft bei YAML-Parsing-Problemen

	// Validierung
	if err := cfg.Validate(); err != nil {
		// Bei target.mode Fehler: Zeige auch host-Wert für Debugging
		if cfg.Target.Mode == "" || (cfg.Target.Mode != "ping" && cfg.Target.Mode != "https" && cfg.Target.Mode != "ping+https") {
			return nil, fmt.Errorf("config validation failed: %w (parsed: mode=%q, host=%q)", err, cfg.Target.Mode, cfg.Target.Host)
		}
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}
