package core

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"prox-watch/internal/config"
)

// Lifecycle manages the application lifecycle.
type Lifecycle struct {
	runner *Runner
	config *config.Config
}

// NewLifecycle creates a new lifecycle manager.
func NewLifecycle(cfg *config.Config) (*Lifecycle, error) {
	runner, err := NewRunner(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create runner: %w", err)
	}

	return &Lifecycle{
		runner: runner,
		config: cfg,
	}, nil
}

// Start starts the application and handles signals.
func (l *Lifecycle) Start() error {
	// Create context for signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start runner in goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := l.runner.Run(ctx); err != nil {
			errChan <- err
		}
	}()

	// Wait for signal or error
	select {
	case sig := <-sigChan:
		fmt.Printf("Received signal: %v\n", sig)
		cancel()
		if err := l.runner.Stop(); err != nil {
			return fmt.Errorf("failed to stop runner: %w", err)
		}
		l.runner.Wait()
		return nil

	case err := <-errChan:
		return fmt.Errorf("runner error: %w", err)
	}
}

// Stop stops the application gracefully.
func (l *Lifecycle) Stop() error {
	return l.runner.Stop()
}
