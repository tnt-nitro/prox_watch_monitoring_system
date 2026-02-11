package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the complete configuration.
type Config struct {
	System  SystemConfig  `yaml:"system"`
	Proxmox ProxmoxConfig `yaml:"proxmox"`
	Alerts  AlertsConfig  `yaml:"alerts"`
	Paths   PathsConfig   `yaml:"paths"`
	Rules   RulesConfig   `yaml:"rules"`
}

// SystemConfig contains system-wide settings.
type SystemConfig struct {
	Timezone string `yaml:"timezone"`
	RunMode  string `yaml:"run_mode"`
}

// ProxmoxConfig contains Proxmox connection settings.
type ProxmoxConfig struct {
	Host    string `yaml:"host"`
	APIPort int    `yaml:"api_port"`
}

// AlertsConfig contains alerting settings.
type AlertsConfig struct {
	Channel string     `yaml:"channel"`
	Ntfy    NtfyConfig `yaml:"ntfy,omitempty"`
}

// NtfyConfig contains ntfy-specific settings.
type NtfyConfig struct {
	Server string      `yaml:"server"`
	Topics TopicConfig `yaml:"topics"`
}

// TopicConfig contains topic names for different severity levels.
type TopicConfig struct {
	Info string `yaml:"info"`
	Warn string `yaml:"warn"`
	Crit string `yaml:"crit"`
}

// PathsConfig contains file system paths.
type PathsConfig struct {
	StateDB string `yaml:"state_db"`
	Config  string `yaml:"config"`
	Secrets string `yaml:"secrets"`
}

// RulesConfig contains counting and timing rules.
type RulesConfig struct {
	Thresholds ThresholdConfig `yaml:"thresholds"`
	Windows    WindowConfig    `yaml:"windows"`
	Cooldown   string          `yaml:"cooldown"`
}

// ThresholdConfig contains threshold values.
type ThresholdConfig struct {
	Warn int `yaml:"warn"`
	Crit int `yaml:"crit"`
}

// WindowConfig contains time window durations.
type WindowConfig struct {
	Warn string `yaml:"warn"`
	Crit string `yaml:"crit"`
}

// DefaultConfig returns a configuration with default values.
func DefaultConfig() *Config {
	return &Config{
		System: SystemConfig{
			Timezone: "UTC",
			RunMode:  "daemon",
		},
		Proxmox: ProxmoxConfig{
			Host:    "PROXMOX_HOSTNAME",
			APIPort: 8006,
		},
		Alerts: AlertsConfig{
			Channel: "local-only",
			Ntfy: NtfyConfig{
				Server: "",
				Topics: TopicConfig{
					Info: "prox-watch-info",
					Warn: "prox-watch-warn",
					Crit: "prox-watch-crit",
				},
			},
		},
		Paths: PathsConfig{
			StateDB: "/var/lib/prox-watch/state.db",
			Config:  "/var/lib/prox-watch/config.yaml",
			Secrets: "/var/lib/prox-watch/secrets.yaml",
		},
		Rules: RulesConfig{
			Thresholds: ThresholdConfig{
				Warn: 3,
				Crit: 10,
			},
			Windows: WindowConfig{
				Warn: "10m",
				Crit: "15m",
			},
			Cooldown: "30m",
		},
	}
}

// Load loads configuration from a YAML file.
// If the file doesn't exist, returns default configuration.
func Load(path string) (*Config, error) {
	// Start with defaults
	cfg := DefaultConfig()

	// Override config path
	cfg.Paths.Config = path

	// Try to load file
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, return defaults
			return cfg, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Ensure config path is set
	if cfg.Paths.Config == "" {
		cfg.Paths.Config = path
	}

	return cfg, nil
}

// Save saves configuration to a YAML file.
func (c *Config) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetWindowDuration parses a time window string and returns a duration.
func (r *RulesConfig) GetWindowDuration(severity string) (time.Duration, error) {
	var windowStr string
	switch severity {
	case "warn":
		windowStr = r.Windows.Warn
	case "crit":
		windowStr = r.Windows.Crit
	default:
		return 0, fmt.Errorf("unknown severity: %s", severity)
	}

	duration, err := time.ParseDuration(windowStr)
	if err != nil {
		return 0, fmt.Errorf("invalid window duration %q: %w", windowStr, err)
	}

	return duration, nil
}

// GetCooldownDuration parses the cooldown string and returns a duration.
func (r *RulesConfig) GetCooldownDuration() (time.Duration, error) {
	duration, err := time.ParseDuration(r.Cooldown)
	if err != nil {
		return 0, fmt.Errorf("invalid cooldown duration %q: %w", r.Cooldown, err)
	}

	return duration, nil
}
