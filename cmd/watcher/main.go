package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"prox-watch/internal/config"
	"prox-watch/internal/push"
	"prox-watch/internal/watcher"
)

const (
	version = "v0.5.0"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	switch command {
	case "run":
		runCommand()
	case "status":
		statusCommand()
	case "ledtest":
		ledtestCommand()
	case "version":
		fmt.Printf("prox-watch-watcher %s\n", version)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: prox-watch-watcher <command>")
	fmt.Println("Commands:")
	fmt.Println("  run     - Start watcher daemon")
	fmt.Println("  status  - Show current status")
	fmt.Println("  ledtest - Start GPIO test tool (LED & Beeper)")
	fmt.Println("  version - Show version")
}

func runCommand() {
	// Parse flags
	var configPath string
	flag.StringVar(&configPath, "config-path", "/etc/prox-watch-watcher/watcher.yaml", "Path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := watcher.LoadWatcherConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatalf("ERROR: Configuration validation failed: %v", err)
	}

	// Warn if host is still PLACEHOLDER
	if cfg.Target.Host == "PLACEHOLDER" {
		log.Fatalf("ERROR: target.host is still 'PLACEHOLDER'. Please update %s with your Proxmox hostname/IP", configPath)
	}

	// Initialize components
	health := watcher.NewHealthChecker(watcher.HealthCheckerConfig{
		Host:    cfg.Target.Host,
		Port:    cfg.Target.Port,
		Mode:    cfg.Target.Mode,
		Timeout: time.Duration(cfg.Target.TimeoutSeconds) * time.Second,
	})

	counter := watcher.NewCounter()

	// Initialize Push Service
	var pushAdapter push.Adapter
	if cfg.Push.Enabled && cfg.Push.Adapter == "ntfy" {
		// Create a minimal config.Config for ntfy adapter
		// We need to adapt WatcherConfig to config.Config format
		coreConfig := &config.Config{
			Alerts: config.AlertsConfig{
				Channel: "ntfy",
				Ntfy: config.NtfyConfig{
					Server: "", // Use default ntfy.sh
					Topics: config.TopicConfig{
						Warn: cfg.Push.Topics.Warn,
						Crit: cfg.Push.Topics.Crit,
						Info: "prox-watch-info", // Default
					},
				},
			},
		}
		pushAdapter = push.NewNtfyAdapter(coreConfig, "")
	} else {
		pushAdapter = push.NewLocalOnlyAdapter()
	}

	pushService := watcher.NewPushService(watcher.PushServiceConfig{
		Adapter: pushAdapter,
		Enabled: cfg.Push.Enabled,
	})

	// Initialize GPIO
	gpio, err := watcher.NewGPIOFromConfig(cfg.GPIO)
	if err != nil {
		log.Fatalf("Failed to initialize GPIO: %v", err)
	}

	// Initialize State Store
	stateDBPath := filepath.Join("/var/lib/prox-watch-watcher", "watcher_state.db")
	store, err := watcher.NewSQLiteStateStore(stateDBPath)
	if err != nil {
		log.Fatalf("Failed to initialize state store: %v", err)
	}
	defer store.Close()

	// Initialize PowerCycler (optional, Phase 3)
	var powerCycler watcher.PowerCycler
	if cfg.PowerCycle.Enabled {
		// Create GPIO pin for power cycle if needed
		// For now, we'll create a base power cycler (without GPIO)
		// GPIO power cycler would require actual pin initialization
		baseCycler, err := watcher.NewPowerCycler(cfg.PowerCycle, store)
		if err != nil {
			log.Printf("WARNING: Failed to initialize power cycler: %v (continuing without power cycle)", err)
		} else {
			powerCycler = baseCycler
		}
	}

	// Create Runner
	runner, err := watcher.NewRunner(watcher.RunnerConfig{
		Health:        health,
		Counter:       counter,
		Push:          pushService,
		GPIO:          gpio,
		PowerCycler:   powerCycler,
		Store:         store,
		Interval:      time.Duration(cfg.Watcher.IntervalSeconds) * time.Second,
		WarnThreshold: cfg.Thresholds.Warn,
		CritThreshold: cfg.Thresholds.Crit,
		CooldownSecs:  cfg.Watcher.CooldownSeconds,
		PowerCycleCfg: cfg.PowerCycle,
	})
	if err != nil {
		log.Fatalf("Failed to create runner: %v", err)
	}

	// Setup signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start runner in goroutine
	errChan := make(chan error, 1)
	go func() {
		log.Printf("✓ Watcher started successfully")
		log.Printf("  Version: %s", version)
		log.Printf("  Interval: %ds", cfg.Watcher.IntervalSeconds)
		log.Printf("  Target: %s:%d (mode: %s)", cfg.Target.Host, cfg.Target.Port, cfg.Target.Mode)
		log.Printf("  Thresholds: WARN=%d, CRIT=%d", cfg.Thresholds.Warn, cfg.Thresholds.Crit)
		log.Printf("  Push notifications: %v", cfg.Push.Enabled)
		if err := runner.Run(ctx); err != nil {
			errChan <- err
		}
	}()

	// Wait for signal or error
	select {
	case sig := <-sigChan:
		log.Printf("Received signal: %v, shutting down gracefully...", sig)
		cancel()
		if err := runner.Stop(); err != nil {
			log.Printf("Error stopping runner: %v", err)
		}
		if err := runner.Wait(); err != nil {
			log.Printf("Error waiting for runner: %v", err)
		}
		log.Println("Watcher stopped")

	case err := <-errChan:
		log.Fatalf("Runner error: %v", err)
	}
}

func statusCommand() {
	// TODO: Implement status command
	// Should read state from SQLite and display current status
	fmt.Println("Status command not yet implemented")
	fmt.Println("Use 'journalctl -u prox-watch-watcher.service -f' to view logs")
}

func ledtestCommand() {
	// Das LED-Test-Tool wird als separates Binary gebaut
	// und sollte im Installationsskript installiert werden
	// Für jetzt: Hinweis, dass das Tool separat gebaut werden muss
	
	fmt.Println("LED & Beeper Test Tool")
	fmt.Println("")
	fmt.Println("Hinweis: Dieses Tool muss separat gebaut werden:")
	fmt.Println("  go build -tags raspberry -o prox-watch-ledtest ./cmd/ledtest")
	fmt.Println("")
	fmt.Println("Dann ausführen mit:")
	fmt.Println("  sudo ./prox-watch-ledtest")
	fmt.Println("")
	fmt.Println("Oder direkt testen mit:")
	fmt.Println("  go run -tags raspberry ./cmd/ledtest")
}
