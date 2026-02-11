package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ValidationError represents a configuration validation error.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error in %s: %s", e.Field, e.Message)
}

var (
	ipv4Pattern = regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`)
	ipv6Pattern = regexp.MustCompile(`\b(?:[0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}\b`)
)

// Validate validates the configuration and returns an error if invalid.
func Validate(c *Config) error {
	// Phase 1: Content validation (IPs, values)
	if err := validateContent(c); err != nil {
		return err
	}

	// Phase 2: Path validation (must be done after repo path is determined)
	// This is done separately in ValidatePaths

	return nil
}

// ValidatePaths validates that all paths are outside the repository.
func ValidatePaths(c *Config, repoPath string) error {
	repoAbs, err := filepath.Abs(repoPath)
	if err != nil {
		return fmt.Errorf("failed to resolve repo path: %w", err)
	}

	// Check config path
	if isInRepo(c.Paths.Config, repoAbs) {
		return &ValidationError{
			Field:   "paths.config",
			Message: "config path must be outside repository",
		}
	}

	// Check state DB path
	if isInRepo(c.Paths.StateDB, repoAbs) {
		return &ValidationError{
			Field:   "paths.state_db",
			Message: "state database path must be outside repository",
		}
	}

	// Check secrets path
	if isInRepo(c.Paths.Secrets, repoAbs) {
		return &ValidationError{
			Field:   "paths.secrets",
			Message: "secrets path must be outside repository",
		}
	}

	return nil
}

// ValidatePermissions validates file system permissions.
func ValidatePermissions(c *Config) error {
	// Check state DB directory
	stateDir := filepath.Dir(c.Paths.StateDB)
	if err := checkWritePermission(stateDir); err != nil {
		return fmt.Errorf("no write permission for state database path: %w", err)
	}

	// Check config directory
	configDir := filepath.Dir(c.Paths.Config)
	if err := checkWritePermission(configDir); err != nil {
		return fmt.Errorf("no write permission for config directory: %w", err)
	}

	return nil
}

// validateContent validates configuration content (IPs, values, etc.).
func validateContent(c *Config) error {
	// Check for IP addresses
	if err := checkIPs(c); err != nil {
		return err
	}

	// Validate timezone
	if err := validateTimezone(c.System.Timezone); err != nil {
		return &ValidationError{
			Field:   "system.timezone",
			Message: err.Error(),
		}
	}

	// Validate API port
	if c.Proxmox.APIPort < 1 || c.Proxmox.APIPort > 65535 {
		return &ValidationError{
			Field:   "proxmox.api_port",
			Message: "invalid API port (must be 1-65535)",
		}
	}

	// Validate alert channel
	if c.Alerts.Channel != "local-only" && c.Alerts.Channel != "ntfy" {
		return &ValidationError{
			Field:   "alerts.channel",
			Message: "invalid alert channel (must be 'local-only' or 'ntfy')",
		}
	}

	// Validate thresholds
	if c.Rules.Thresholds.Warn <= 0 {
		return &ValidationError{
			Field:   "rules.thresholds.warn",
			Message: "thresholds must be greater than 0",
		}
	}
	if c.Rules.Thresholds.Crit <= 0 {
		return &ValidationError{
			Field:   "rules.thresholds.crit",
			Message: "thresholds must be greater than 0",
		}
	}

	// Validate time windows
	if _, err := c.Rules.GetWindowDuration("warn"); err != nil {
		return &ValidationError{
			Field:   "rules.windows.warn",
			Message: fmt.Sprintf("invalid time window: %v", err),
		}
	}
	if _, err := c.Rules.GetWindowDuration("crit"); err != nil {
		return &ValidationError{
			Field:   "rules.windows.crit",
			Message: fmt.Sprintf("invalid time window: %v", err),
		}
	}

	// Validate cooldown
	if _, err := c.Rules.GetCooldownDuration(); err != nil {
		return &ValidationError{
			Field:   "rules.cooldown",
			Message: fmt.Sprintf("invalid cooldown duration: %v", err),
		}
	}

	// Warn about hostname with dots (non-blocking)
	if strings.Contains(c.Proxmox.Host, ".") {
		// This is a warning, not an error
		// Could be logged but doesn't block
	}

	return nil
}

// checkIPs checks for IP addresses in string fields.
func checkIPs(c *Config) error {
	fields := map[string]string{
		"proxmox.host":        c.Proxmox.Host,
		"alerts.ntfy.server":  c.Alerts.Ntfy.Server,
		"alerts.ntfy.topics.info": c.Alerts.Ntfy.Topics.Info,
		"alerts.ntfy.topics.warn": c.Alerts.Ntfy.Topics.Warn,
		"alerts.ntfy.topics.crit": c.Alerts.Ntfy.Topics.Crit,
	}

	for field, value := range fields {
		if containsIP(value) {
			return &ValidationError{
				Field:   field,
				Message: "IP addresses not allowed in configuration",
			}
		}
	}

	return nil
}

// containsIP checks if a string contains an IP address.
func containsIP(s string) bool {
	return ipv4Pattern.MatchString(s) || ipv6Pattern.MatchString(s)
}

// isInRepo checks if a path is within the repository directory.
func isInRepo(path, repoPath string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	return strings.HasPrefix(absPath, repoPath)
}

// validateTimezone validates an IANA timezone string.
func validateTimezone(tz string) error {
	// Basic validation - could be enhanced with actual timezone database check
	if tz == "" {
		return fmt.Errorf("timezone cannot be empty")
	}
	// For MVP, just check it's not obviously invalid
	// Full validation would require time.LoadLocation()
	return nil
}

// checkWritePermission checks if a directory is writable.
func checkWritePermission(dir string) error {
	// Check if directory exists
	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			// Try to create parent directory
			parent := filepath.Dir(dir)
			if err := os.MkdirAll(parent, 0700); err != nil {
				return fmt.Errorf("cannot create directory: %w", err)
			}
			return nil
		}
		return err
	}

	// Check if it's a directory
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory")
	}

	// Check write permission by trying to create a test file
	testFile := filepath.Join(dir, ".write-test")
	if err := os.WriteFile(testFile, []byte("test"), 0600); err != nil {
		return fmt.Errorf("directory not writable: %w", err)
	}
	os.Remove(testFile) // Clean up

	return nil
}
