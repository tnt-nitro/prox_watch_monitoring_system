package state

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"prox-watch/internal/rules"
)

// Store is the interface for state management.
type Store interface {
	// Increment increments the counter for an event ID and returns the new count state.
	Increment(eventID string, ts time.Time) (CountState, error)

	// Get retrieves the count state for an event ID.
	Get(eventID string) (CountState, error)

	// SetCooldown sets a cooldown period for an event ID.
	SetCooldown(eventID string, until time.Time) error

	// IsCooldown checks if an event ID is currently in cooldown.
	IsCooldown(eventID string, now time.Time) bool

	// Acknowledge sets an acknowledgment for an event ID.
	Ack(eventID string, until time.Time) error

	// IsAcked checks if an event ID is currently acknowledged.
	IsAcked(eventID string, now time.Time) bool

	// Close closes the database connection.
	Close() error
}

// SQLiteStore implements the Store interface using SQLite.
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore creates a new SQLite store and initializes the database.
func NewSQLiteStore(path string) (*SQLiteStore, error) {
	// Create directory if needed
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create state directory: %w", err)
	}

	// Open database
	db, err := sql.Open("sqlite3", path+"?_journal_mode=WAL&_foreign_keys=1")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	store := &SQLiteStore{db: db}

	// Initialize schema
	if err := store.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return store, nil
}

// initSchema creates the database schema if it doesn't exist.
func (s *SQLiteStore) initSchema() error {
	schema := `
	-- Events table
	CREATE TABLE IF NOT EXISTS events (
		event_id TEXT PRIMARY KEY,
		severity INTEGER NOT NULL,
		count INTEGER NOT NULL DEFAULT 0,
		first_seen INTEGER NOT NULL,
		last_seen INTEGER NOT NULL
	);

	-- Cooldowns table
	CREATE TABLE IF NOT EXISTS cooldowns (
		event_id TEXT PRIMARY KEY,
		cooldown_until INTEGER NOT NULL,
		FOREIGN KEY (event_id) REFERENCES events(event_id) ON DELETE CASCADE
	);

	-- Acknowledges table
	CREATE TABLE IF NOT EXISTS acknowledges (
		event_id TEXT PRIMARY KEY,
		ack_until INTEGER NOT NULL,
		FOREIGN KEY (event_id) REFERENCES events(event_id) ON DELETE CASCADE
	);

	-- Indexes
	CREATE INDEX IF NOT EXISTS idx_events_severity ON events(severity);
	CREATE INDEX IF NOT EXISTS idx_events_last_seen ON events(last_seen);
	CREATE INDEX IF NOT EXISTS idx_cooldowns_until ON cooldowns(cooldown_until);
	CREATE INDEX IF NOT EXISTS idx_acknowledges_until ON acknowledges(ack_until);
	`

	if _, err := s.db.Exec(schema); err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	return nil
}

// Increment increments the counter for an event ID.
func (s *SQLiteStore) Increment(eventID string, ts time.Time) (CountState, error) {
	ctx := context.Background()

	// Use transaction for atomicity
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return CountState{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Check if event exists
	var count int
	var firstSeen, lastSeen int64
	err = tx.QueryRowContext(ctx,
		"SELECT count, first_seen, last_seen FROM events WHERE event_id = ?",
		eventID).Scan(&count, &firstSeen, &lastSeen)

	if err == sql.ErrNoRows {
		// Create new event
		firstSeen = ts.Unix()
		lastSeen = ts.Unix()
		count = 1

		_, err = tx.ExecContext(ctx,
			"INSERT INTO events (event_id, severity, count, first_seen, last_seen) VALUES (?, ?, ?, ?, ?)",
			eventID, int(rules.SeverityInfo), count, firstSeen, lastSeen)
		if err != nil {
			return CountState{}, fmt.Errorf("failed to insert event: %w", err)
		}
	} else if err != nil {
		return CountState{}, fmt.Errorf("failed to query event: %w", err)
	} else {
		// Update existing event
		count++
		lastSeen = ts.Unix()

		_, err = tx.ExecContext(ctx,
			"UPDATE events SET count = ?, last_seen = ? WHERE event_id = ?",
			count, lastSeen, eventID)
		if err != nil {
			return CountState{}, fmt.Errorf("failed to update event: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return CountState{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return CountState{
		EventID:   eventID,
		Count:     count,
		FirstSeen: time.Unix(firstSeen, 0),
		LastSeen:  time.Unix(lastSeen, 0),
	}, nil
}

// Get retrieves the count state for an event ID.
func (s *SQLiteStore) Get(eventID string) (CountState, error) {
	ctx := context.Background()

	var count int
	var firstSeen, lastSeen int64
	err := s.db.QueryRowContext(ctx,
		"SELECT count, first_seen, last_seen FROM events WHERE event_id = ?",
		eventID).Scan(&count, &firstSeen, &lastSeen)

	if err == sql.ErrNoRows {
		return CountState{
			EventID: eventID,
			Count:   0,
		}, nil
	}
	if err != nil {
		return CountState{}, fmt.Errorf("failed to query event: %w", err)
	}

	return CountState{
		EventID:   eventID,
		Count:     count,
		FirstSeen: time.Unix(firstSeen, 0),
		LastSeen:  time.Unix(lastSeen, 0),
	}, nil
}

// GetEvent retrieves the full event including severity.
func (s *SQLiteStore) GetEvent(eventID string) (*Event, error) {
	ctx := context.Background()

	var severity int
	var count int
	var firstSeen, lastSeen int64
	err := s.db.QueryRowContext(ctx,
		"SELECT severity, count, first_seen, last_seen FROM events WHERE event_id = ?",
		eventID).Scan(&severity, &count, &firstSeen, &lastSeen)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("event not found: %s", eventID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query event: %w", err)
	}

	return &Event{
		EventID:   eventID,
		Severity:  rules.Severity(severity),
		Count:     count,
		FirstSeen: time.Unix(firstSeen, 0),
		LastSeen:  time.Unix(lastSeen, 0),
	}, nil
}

// SetSeverity updates the severity for an event.
func (s *SQLiteStore) SetSeverity(eventID string, severity rules.Severity) error {
	ctx := context.Background()

	_, err := s.db.ExecContext(ctx,
		"UPDATE events SET severity = ? WHERE event_id = ?",
		int(severity), eventID)
	if err != nil {
		return fmt.Errorf("failed to update severity: %w", err)
	}

	return nil
}

// SetCooldown sets a cooldown period for an event ID.
func (s *SQLiteStore) SetCooldown(eventID string, until time.Time) error {
	ctx := context.Background()

	untilUnix := until.Unix()

	// Use INSERT OR REPLACE for upsert
	_, err := s.db.ExecContext(ctx,
		"INSERT OR REPLACE INTO cooldowns (event_id, cooldown_until) VALUES (?, ?)",
		eventID, untilUnix)
	if err != nil {
		return fmt.Errorf("failed to set cooldown: %w", err)
	}

	return nil
}

// IsCooldown checks if an event ID is currently in cooldown.
func (s *SQLiteStore) IsCooldown(eventID string, now time.Time) bool {
	ctx := context.Background()

	var untilUnix int64
	err := s.db.QueryRowContext(ctx,
		"SELECT cooldown_until FROM cooldowns WHERE event_id = ?",
		eventID).Scan(&untilUnix)

	if err == sql.ErrNoRows {
		return false
	}
	if err != nil {
		// On error, assume not in cooldown
		return false
	}

	until := time.Unix(untilUnix, 0)
	return now.Before(until)
}

// Ack sets an acknowledgment for an event ID.
func (s *SQLiteStore) Ack(eventID string, until time.Time) error {
	ctx := context.Background()

	untilUnix := until.Unix()

	// Use INSERT OR REPLACE for upsert
	_, err := s.db.ExecContext(ctx,
		"INSERT OR REPLACE INTO acknowledges (event_id, ack_until) VALUES (?, ?)",
		eventID, untilUnix)
	if err != nil {
		return fmt.Errorf("failed to set acknowledge: %w", err)
	}

	return nil
}

// IsAcked checks if an event ID is currently acknowledged.
func (s *SQLiteStore) IsAcked(eventID string, now time.Time) bool {
	ctx := context.Background()

	var untilUnix int64
	err := s.db.QueryRowContext(ctx,
		"SELECT ack_until FROM acknowledges WHERE event_id = ?",
		eventID).Scan(&untilUnix)

	if err == sql.ErrNoRows {
		return false
	}
	if err != nil {
		// On error, assume not acknowledged
		return false
	}

	until := time.Unix(untilUnix, 0)
	return now.Before(until)
}

// GetAllEvents retrieves all events from the database.
func (s *SQLiteStore) GetAllEvents() ([]*Event, error) {
	ctx := context.Background()

	rows, err := s.db.QueryContext(ctx,
		"SELECT event_id, severity, count, first_seen, last_seen FROM events ORDER BY last_seen DESC")
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	var events []*Event
	for rows.Next() {
		var eventID string
		var severity int
		var count int
		var firstSeen, lastSeen int64

		if err := rows.Scan(&eventID, &severity, &count, &firstSeen, &lastSeen); err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}

		events = append(events, &Event{
			EventID:   eventID,
			Severity:  rules.Severity(severity),
			Count:     count,
			FirstSeen: time.Unix(firstSeen, 0),
			LastSeen:  time.Unix(lastSeen, 0),
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating events: %w", err)
	}

	return events, nil
}

// Close closes the database connection.
func (s *SQLiteStore) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}
