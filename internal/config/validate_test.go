package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidate_IPAddress(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Proxmox.Host = "192.168.1.1" // IP address should be rejected

	err := Validate(cfg)
	if err == nil {
		t.Error("Validate() should reject IP addresses in hostname")
	}

	validationErr, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("Expected ValidationError, got %T", err)
	}

	if validationErr.Field != "proxmox.host" {
		t.Errorf("Expected field 'proxmox.host', got '%s'", validationErr.Field)
	}
}

func TestValidate_ValidHostname(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Proxmox.Host = "PROXMOX_HOSTNAME" // Placeholder is OK

	if err := Validate(cfg); err != nil {
		t.Errorf("Validate() should accept placeholder hostname: %v", err)
	}
}

func TestValidate_InvalidChannel(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Alerts.Channel = "invalid-channel"

	err := Validate(cfg)
	if err == nil {
		t.Error("Validate() should reject invalid channel")
	}
}

func TestValidate_ValidChannel(t *testing.T) {
	tests := []string{"local-only", "ntfy"}

	for _, channel := range tests {
		cfg := DefaultConfig()
		cfg.Alerts.Channel = channel

		if err := Validate(cfg); err != nil {
			t.Errorf("Validate() should accept channel '%s': %v", channel, err)
		}
	}
}

func TestValidatePaths_OutsideRepo(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "repo")
	os.MkdirAll(repoPath, 0755)

	cfg := DefaultConfig()
	cfg.Paths.Config = filepath.Join(tmpDir, "config.yaml")
	cfg.Paths.StateDB = filepath.Join(tmpDir, "state.db")
	cfg.Paths.Secrets = filepath.Join(tmpDir, "secrets.yaml")

	if err := ValidatePaths(cfg, repoPath); err != nil {
		t.Errorf("ValidatePaths() should accept paths outside repo: %v", err)
	}
}

func TestValidatePaths_InsideRepo(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := tmpDir

	cfg := DefaultConfig()
	cfg.Paths.Config = filepath.Join(repoPath, "config.yaml")

	err := ValidatePaths(cfg, repoPath)
	if err == nil {
		t.Error("ValidatePaths() should reject paths inside repo")
	}

	validationErr, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("Expected ValidationError, got %T", err)
	}

	if validationErr.Field != "paths.config" {
		t.Errorf("Expected field 'paths.config', got '%s'", validationErr.Field)
	}
}

func TestValidatePermissions(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := DefaultConfig()
	cfg.Paths.Config = filepath.Join(tmpDir, "config.yaml")

	// Create config file
	if err := os.WriteFile(cfg.Paths.Config, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Should pass (file exists and is readable)
	if err := ValidatePermissions(cfg); err != nil {
		t.Errorf("ValidatePermissions() should pass for readable file: %v", err)
	}
}

func TestValidatePermissions_MissingFile(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := DefaultConfig()
	cfg.Paths.Config = filepath.Join(tmpDir, "nonexistent.yaml")

	// Should fail (file doesn't exist)
	err := ValidatePermissions(cfg)
	if err == nil {
		t.Error("ValidatePermissions() should fail for missing file")
	}
}
