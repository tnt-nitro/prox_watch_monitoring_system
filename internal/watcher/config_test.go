package watcher

import (
	"testing"
)

func TestDefaultWatcherConfig(t *testing.T) {
	cfg := DefaultWatcherConfig()

	// Watcher-Section
	if cfg.Watcher.IntervalSeconds != 30 {
		t.Errorf("Expected IntervalSeconds=30, got %d", cfg.Watcher.IntervalSeconds)
	}
	if cfg.Watcher.CooldownSeconds != 600 {
		t.Errorf("Expected CooldownSeconds=600, got %d", cfg.Watcher.CooldownSeconds)
	}

	// Target-Section
	if cfg.Target.Mode != "ping+https" {
		t.Errorf("Expected Mode=ping+https, got %q", cfg.Target.Mode)
	}
	if cfg.Target.Port != 8006 {
		t.Errorf("Expected Port=8006, got %d", cfg.Target.Port)
	}

	// Thresholds-Section
	if cfg.Thresholds.Warn != 3 {
		t.Errorf("Expected Warn=3, got %d", cfg.Thresholds.Warn)
	}
	if cfg.Thresholds.Crit != 10 {
		t.Errorf("Expected Crit=10, got %d", cfg.Thresholds.Crit)
	}
}

func TestWatcherConfig_Validate_Cooldown_Default(t *testing.T) {
	cfg := DefaultWatcherConfig()
	if err := cfg.Validate(); err != nil {
		t.Errorf("Expected no error for default config, got %v", err)
	}
}

func TestWatcherConfig_Validate_Cooldown_Zero(t *testing.T) {
	cfg := DefaultWatcherConfig()
	cfg.Watcher.CooldownSeconds = 0
	if err := cfg.Validate(); err != nil {
		t.Errorf("Expected no error for cooldown=0 (disabled), got %v", err)
	}
}

func TestWatcherConfig_Validate_Cooldown_Negative(t *testing.T) {
	cfg := DefaultWatcherConfig()
	cfg.Watcher.CooldownSeconds = -1
	if err := cfg.Validate(); err == nil {
		t.Error("Expected error for negative cooldown, got nil")
	}
}

func TestWatcherConfig_Validate_Cooldown_Valid(t *testing.T) {
	cfg := DefaultWatcherConfig()
	cfg.Watcher.CooldownSeconds = 900 // 15 Minuten
	if err := cfg.Validate(); err != nil {
		t.Errorf("Expected no error for cooldown=900, got %v", err)
	}
}

func TestWatcherConfig_Validate_Cooldown_TooHigh(t *testing.T) {
	cfg := DefaultWatcherConfig()
	cfg.Watcher.CooldownSeconds = 999999
	if err := cfg.Validate(); err == nil {
		t.Error("Expected error for cooldown > 86400, got nil")
	}
}

func TestWatcherConfig_Validate_Cooldown_MaxAllowed(t *testing.T) {
	cfg := DefaultWatcherConfig()
	cfg.Watcher.CooldownSeconds = 86400 // 24 Stunden (Max)
	if err := cfg.Validate(); err != nil {
		t.Errorf("Expected no error for cooldown=86400 (max), got %v", err)
	}
}

func TestWatcherConfig_Validate_Cooldown_JustOverMax(t *testing.T) {
	cfg := DefaultWatcherConfig()
	cfg.Watcher.CooldownSeconds = 86401 // 1 Sekunde über Max
	if err := cfg.Validate(); err == nil {
		t.Error("Expected error for cooldown=86401 (> max), got nil")
	}
}

func TestWatcherConfig_Validate_Interval(t *testing.T) {
	cfg := DefaultWatcherConfig()

	// Valid
	cfg.Watcher.IntervalSeconds = 30
	if err := cfg.Validate(); err != nil {
		t.Errorf("Expected no error for interval=30, got %v", err)
	}

	// Too low
	cfg.Watcher.IntervalSeconds = 9
	if err := cfg.Validate(); err == nil {
		t.Error("Expected error for interval < 10, got nil")
	}

	// Too high
	cfg.Watcher.IntervalSeconds = 301
	if err := cfg.Validate(); err == nil {
		t.Error("Expected error for interval > 300, got nil")
	}
}

func TestWatcherConfig_Validate_Thresholds(t *testing.T) {
	cfg := DefaultWatcherConfig()

	// Valid
	cfg.Thresholds.Warn = 3
	cfg.Thresholds.Crit = 10
	if err := cfg.Validate(); err != nil {
		t.Errorf("Expected no error for valid thresholds, got %v", err)
	}

	// Warn >= Crit
	cfg.Thresholds.Warn = 10
	cfg.Thresholds.Crit = 10
	if err := cfg.Validate(); err == nil {
		t.Error("Expected error for warn >= crit, got nil")
	}

	// Warn < 1
	cfg.Thresholds.Warn = 0
	cfg.Thresholds.Crit = 10
	if err := cfg.Validate(); err == nil {
		t.Error("Expected error for warn < 1, got nil")
	}
}
