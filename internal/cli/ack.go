package cli

import (
	"fmt"
	"time"

	"prox-watch/internal/config"
	"prox-watch/internal/state"
)

// RunAck acknowledges an event for a specified duration.
func RunAck(cfgPath string, eventID string, duration time.Duration) error {
	// Load config
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Open state store
	store, err := state.NewSQLiteStore(cfg.Paths.StateDB)
	if err != nil {
		return fmt.Errorf("failed to open state store: %w", err)
	}
	defer store.Close()

	// Check if event exists
	_, err = store.GetEvent(eventID)
	if err != nil {
		return fmt.Errorf("event not found: %s", eventID)
	}

	// Set acknowledge
	until := time.Now().Add(duration)
	if err := store.Ack(eventID, until); err != nil {
		return fmt.Errorf("failed to acknowledge event: %w", err)
	}

	fmt.Printf("Event %s acknowledged until %s\n", eventID, until.Format(time.RFC3339))
	return nil
}

// RunAckDefault acknowledges an event for the default duration (24 hours).
func RunAckDefault(cfgPath string, eventID string) error {
	return RunAck(cfgPath, eventID, 24*time.Hour)
}
