package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"prox-watch/internal/config"
)

// InitWizard runs the configuration initialization wizard.
func InitWizard(configPath string) error {
	cfg := config.DefaultConfig()

	// Override config path
	cfg.Paths.Config = configPath

	fmt.Println("prox-watch Configuration Wizard")
	fmt.Println("================================")
	fmt.Println()

	// Timezone
	timezone, err := askQuestion("Timezone", cfg.System.Timezone)
	if err != nil {
		return err
	}
	cfg.System.Timezone = timezone

	// Proxmox Host
	host, err := askQuestion("Proxmox Host (placeholder, no IPs)", cfg.Proxmox.Host)
	if err != nil {
		return err
	}
	cfg.Proxmox.Host = host

	// API Port
	apiPort, err := askQuestionInt("API Port", fmt.Sprintf("%d", cfg.Proxmox.APIPort))
	if err != nil {
		return err
	}
	cfg.Proxmox.APIPort = apiPort

	// Alert Channel
	channel, err := askChoice("Alert Channel", []string{"local-only", "ntfy"}, cfg.Alerts.Channel)
	if err != nil {
		return err
	}
	cfg.Alerts.Channel = channel

	// If ntfy, ask for server
	if channel == "ntfy" {
		server, err := askQuestion("ntfy Server URL (optional)", "")
		if err != nil {
			return err
		}
		cfg.Alerts.Ntfy.Server = server
	}

	// Config Path confirmation
	fmt.Println()
	fmt.Printf("Config will be saved to: %s\n", configPath)
	confirm, err := askYesNo("Confirm this path?")
	if err != nil {
		return err
	}
	if !confirm {
		return fmt.Errorf("configuration cancelled")
	}

	// Validate before saving
	// Find repo path (look for .git directory)
	repoPath, err := findRepoPath()
	if err == nil {
		if err := config.ValidatePaths(cfg, repoPath); err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}
	}

	// Validate content
	if err := config.Validate(cfg); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Validate permissions
	if err := config.ValidatePermissions(cfg); err != nil {
		return fmt.Errorf("permission check failed: %w", err)
	}

	// Create directory if needed
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Save configuration
	if err := cfg.Save(configPath); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	// Create empty secrets file
	secretsPath := cfg.Paths.Secrets
	secretsDir := filepath.Dir(secretsPath)
	if err := os.MkdirAll(secretsDir, 0700); err != nil {
		return fmt.Errorf("failed to create secrets directory: %w", err)
	}

	secretsContent := `# Secrets file
# This file is not versioned and contains sensitive information

proxmox:
  api_token: ""

ntfy:
  token: ""
`
	if err := os.WriteFile(secretsPath, []byte(secretsContent), 0600); err != nil {
		return fmt.Errorf("failed to create secrets file: %w", err)
	}

	fmt.Println()
	fmt.Println("✓ Configuration saved successfully")
	fmt.Printf("  Config: %s\n", configPath)
	fmt.Printf("  Secrets: %s\n", secretsPath)
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Review the configuration file")
	fmt.Println("  2. Edit secrets.yaml if needed")
	fmt.Println("  3. Start the service: systemctl start prox-watch")

	return nil
}

// askQuestion asks a question and returns the answer.
func askQuestion(prompt, defaultValue string) (string, error) {
	fmt.Printf("%s", prompt)
	if defaultValue != "" {
		fmt.Printf(" [%s]", defaultValue)
	}
	fmt.Print(": ")

	reader := bufio.NewReader(os.Stdin)
	answer, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	answer = strings.TrimSpace(answer)
	if answer == "" {
		return defaultValue, nil
	}

	return answer, nil
}

// askQuestionInt asks a question and returns an integer.
func askQuestionInt(prompt, defaultValue string) (int, error) {
	for {
		answer, err := askQuestion(prompt, defaultValue)
		if err != nil {
			return 0, err
		}

		var value int
		if _, err := fmt.Sscanf(answer, "%d", &value); err != nil {
			fmt.Println("Invalid number, please try again")
			continue
		}

		return value, nil
	}
}

// askChoice asks a question with multiple choices.
func askChoice(prompt string, choices []string, defaultValue string) (string, error) {
	fmt.Printf("%s", prompt)
	fmt.Print(" (")
	for i, choice := range choices {
		if i > 0 {
			fmt.Print(" | ")
		}
		if choice == defaultValue {
			fmt.Print(strings.ToUpper(choice))
		} else {
			fmt.Print(choice)
		}
	}
	fmt.Print("): ")

	reader := bufio.NewReader(os.Stdin)
	answer, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	answer = strings.TrimSpace(strings.ToLower(answer))
	if answer == "" {
		return defaultValue, nil
	}

	// Check if answer matches any choice
	for _, choice := range choices {
		if strings.ToLower(choice) == answer {
			return choice, nil
		}
	}

	fmt.Printf("Invalid choice, using default: %s\n", defaultValue)
	return defaultValue, nil
}

// askYesNo asks a yes/no question.
func askYesNo(prompt string) (bool, error) {
	answer, err := askQuestion(prompt, "yes")
	if err != nil {
		return false, err
	}

	answer = strings.ToLower(strings.TrimSpace(answer))
	return answer == "yes" || answer == "y", nil
}

// findRepoPath finds the repository root by looking for .git directory.
func findRepoPath() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	dir := wd
	for {
		gitPath := filepath.Join(dir, ".git")
		if _, err := os.Stat(gitPath); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("repository root not found")
}
