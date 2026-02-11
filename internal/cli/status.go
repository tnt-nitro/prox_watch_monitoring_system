package cli

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"prox-watch/internal/config"
	"prox-watch/internal/rules"
	"prox-watch/internal/state"
)

// RunStatus displays the current status of events.
func RunStatus(cfgPath string) error {
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

	// Query all events
	// Note: This requires a method to list all events
	// For MVP, we'll show a summary or specific events
	// This is a simplified version - full implementation would query all events

	fmt.Println("prox-watch Status")
	fmt.Println("================")
	fmt.Println()

	// Show summary
	fmt.Println("State Database:", cfg.Paths.StateDB)
	fmt.Println("Config:", cfg.Paths.Config)
	fmt.Println()

	// Note: Full implementation would list all events
	// For MVP, we show a placeholder
	fmt.Println("Events: (use specific event ID to query)")
	fmt.Println("  Use 'prox-watch status <event_id>' for details")

	return nil
}

// RunStatusEvent displays status for a specific event.
func RunStatusEvent(cfgPath string, eventID string) error {
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

	// Get event
	event, err := store.GetEvent(eventID)
	if err != nil {
		return fmt.Errorf("event not found: %s", eventID)
	}

	// Get count state
	countState, err := store.Get(eventID)
	if err != nil {
		return fmt.Errorf("failed to get count state: %w", err)
	}

	// Check cooldown
	isCooldown := store.IsCooldown(eventID, time.Now())
	isAcked := store.IsAcked(eventID, time.Now())

	// Display event information
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "Event ID:\t%s\n", event.EventID)
	fmt.Fprintf(w, "Severity:\t%s\n", event.Severity.String())
	fmt.Fprintf(w, "Count:\t%d\n", event.Count)
	fmt.Fprintf(w, "First Seen:\t%s\n", event.FirstSeen.Format(time.RFC3339))
	fmt.Fprintf(w, "Last Seen:\t%s\n", event.LastSeen.Format(time.RFC3339))
	fmt.Fprintf(w, "In Cooldown:\t%v\n", isCooldown)
	fmt.Fprintf(w, "Acknowledged:\t%v\n", isAcked)
	w.Flush()

	return nil
}
