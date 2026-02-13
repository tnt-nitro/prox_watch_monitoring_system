package core

import (
	"context"
	"fmt"
	"time"

	"prox-watch/internal/config"
	"prox-watch/internal/journal"
	"prox-watch/internal/pattern"
	"prox-watch/internal/push"
	"prox-watch/internal/rules"
	"prox-watch/internal/state"
)

// Runner orchestrates the event processing loop.
type Runner struct {
	config    *config.Config
	journal   journal.Reader
	matcher   pattern.Matcher
	store     state.Store
	evaluator *rules.Evaluator
	push      push.Adapter

	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{}
}

// NewRunner creates a new core runner.
func NewRunner(cfg *config.Config) (*Runner, error) {
	// Create journal reader
	journalReader := journal.NewSystemdReader()

	// Create pattern matcher
	matcher := pattern.NewPatternMatcher()

	// Create state store
	store, err := state.NewSQLiteStore(cfg.Paths.StateDB)
	if err != nil {
		return nil, fmt.Errorf("failed to create state store: %w", err)
	}

	// Create severity evaluator
	evaluator := rules.NewEvaluator(cfg)

	// Create push adapter
	var pushAdapter push.Adapter
	if cfg.Alerts.Channel == "ntfy" {
		// Load token from secrets if available
		// For MVP, we use empty token
		token := "" // TODO: Load from secrets
		pushAdapter = push.NewNtfyAdapter(cfg, token)
	} else {
		// Local-only mode
		pushAdapter = push.NewLocalOnlyAdapter()
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Runner{
		config:    cfg,
		journal:   journalReader,
		matcher:   matcher,
		store:     store,
		evaluator: evaluator,
		push:      pushAdapter,
		ctx:       ctx,
		cancel:    cancel,
		done:      make(chan struct{}),
	}, nil
}

// Run starts the event processing loop.
func (r *Runner) Run(ctx context.Context) error {
	// Load patterns if configured
	// For MVP, patterns are loaded separately or use defaults

	// Start journal reader
	entries, err := r.journal.Read(r.ctx)
	if err != nil {
		return fmt.Errorf("failed to start journal reader: %w", err)
	}

	// Process entries in loop
	for {
		select {
		case entry, ok := <-entries:
			if !ok {
				// Channel closed
				return fmt.Errorf("journal reader closed")
			}

			// Process entry
			if err := r.ProcessEntry(ctx, entry); err != nil {
				// Log error but continue processing
				// In production, this would use proper logging
				continue
			}

		case <-ctx.Done():
			// Context cancelled
			return ctx.Err()

		case <-r.ctx.Done():
			// Internal cancellation
			return nil
		}
	}
}

// ProcessEntry processes a single journal entry.
func (r *Runner) ProcessEntry(ctx context.Context, entry journal.Entry) error {
	// Step 1: Pattern matching
	matchResult, err := r.matcher.Match(ctx, entry)
	if err != nil {
		return fmt.Errorf("pattern matching failed: %w", err)
	}

	// No match, skip processing
	if matchResult == nil {
		return nil
	}

	// Step 2: Increment counter in state store
	countState, err := r.store.Increment(matchResult.EventID, entry.Timestamp)
	if err != nil {
		return fmt.Errorf("failed to increment counter: %w", err)
	}

	// Step 3: Evaluate severity
	isHardError := r.evaluator.IsHardError(matchResult.EventID)
	severity := r.evaluator.Evaluate(countState, time.Now(), isHardError)

	// Step 4: Update severity in state store if changed
	if severity != matchResult.Severity {
		// Update severity in store (convert to int to avoid import cycles)
		if err := r.store.SetSeverity(matchResult.EventID, int(severity)); err != nil {
			// Log error but don't fail
		}
	}

	// Step 5: Check cooldown and acknowledge
	if r.store.IsCooldown(matchResult.EventID, time.Now()) {
		// In cooldown, skip push
		return nil
	}

	if r.store.IsAcked(matchResult.EventID, time.Now()) {
		// Acknowledged, skip push
		return nil
	}

	// Step 6: Push notification (if severity requires it)
	if severity == rules.SeverityCrit {
		// CRIT always pushes
		topic := r.push.GetTopic(severity)
		message := push.Message{
			EventID:   matchResult.EventID,
			Severity:  severity,
			Timestamp: time.Now(),
		}
		if err := r.push.Send(ctx, topic, message); err != nil {
			// Log error but don't fail
			// In production, this would use proper logging
		}
	} else if severity == rules.SeverityWarn {
		// WARN pushes optionally (based on config)
		// For MVP, we push WARN as well
		topic := r.push.GetTopic(severity)
		message := push.Message{
			EventID:   matchResult.EventID,
			Severity:  severity,
			Timestamp: time.Now(),
		}
		if err := r.push.Send(ctx, topic, message); err != nil {
			// Log error but don't fail
		}
	}
	// INFO never pushes

	// Step 7: Set cooldown after push
	cooldownDuration, err := r.config.Rules.GetCooldownDuration()
	if err != nil {
		// Use default if parsing fails
		cooldownDuration = 30 * time.Minute
	}
	cooldownUntil := time.Now().Add(cooldownDuration)
	if err := r.store.SetCooldown(matchResult.EventID, cooldownUntil); err != nil {
		// Log error but don't fail
	}

	return nil
}

// Stop gracefully stops the runner.
func (r *Runner) Stop() error {
	// Cancel context
	r.cancel()

	// Close journal reader
	if err := r.journal.Close(); err != nil {
		return fmt.Errorf("failed to close journal reader: %w", err)
	}

	// Close state store
	if err := r.store.Close(); err != nil {
		return fmt.Errorf("failed to close state store: %w", err)
	}

	// Signal done
	close(r.done)

	return nil
}

// Wait waits for the runner to stop.
func (r *Runner) Wait() {
	<-r.done
}
