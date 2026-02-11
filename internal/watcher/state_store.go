package watcher

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

// PersistedState enthält den minimalen, persistenten Zustand des Watchers.
// Phase 2: FailCount, CurrentSeverity, LastEscalation.
type PersistedState struct {
	FailCount       int
	CurrentSeverity rules.Severity
	LastEscalation  time.Time
}

// StateStore ist das Interface für den minimalen Persistenzlayer des Watchers.
type StateStore interface {
	Load() (PersistedState, error)
	Save(PersistedState) error
	Close() error
}

// sqliteStateStore implementiert StateStore mit einer einzelnen SQLite-Tabelle
// und genau einer Zeile (id = 1).
type sqliteStateStore struct {
	db *sql.DB
}

// NewSQLiteStateStore erstellt einen neuen StateStore für den Watcher.
// path ist der vollständige Pfad zur Datei (z.B. /var/lib/prox-watch-watcher/watcher_state.db).
func NewSQLiteStateStore(path string) (StateStore, error) {
	// Verzeichnis anlegen (0700, nur Owner)
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("watcher state: failed to create directory: %w", err)
	}

	// SQLite-DB öffnen (WAL optional, aber hilfreich)
	db, err := sql.Open("sqlite3", path+"?_journal_mode=WAL&_foreign_keys=1")
	if err != nil {
		return nil, fmt.Errorf("watcher state: failed to open database: %w", err)
	}

	store := &sqliteStateStore{db: db}

	if err := store.initSchema(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("watcher state: failed to init schema: %w", err)
	}

	return store, nil
}

// initSchema legt die minimale Tabelle für den Watcher-State an.
func (s *sqliteStateStore) initSchema() error {
	schema := `
CREATE TABLE IF NOT EXISTS watcher_state (
	id INTEGER PRIMARY KEY CHECK (id = 1),
	fail_count INTEGER NOT NULL,
	current_severity INTEGER NOT NULL,
	last_escalation INTEGER NOT NULL
);`

	if _, err := s.db.Exec(schema); err != nil {
		return fmt.Errorf("watcher state: failed to create schema: %w", err)
	}

	return nil
}

// Load lädt den persistenten Zustand.
// Wenn keine Zeile existiert, wird ein Default-State zurückgegeben.
func (s *sqliteStateStore) Load() (PersistedState, error) {
	ctx := context.Background()

	row := s.db.QueryRowContext(ctx,
		"SELECT fail_count, current_severity, last_escalation FROM watcher_state WHERE id = 1",
	)

	var failCount int
	var severityInt int
	var lastEscUnix int64

	err := row.Scan(&failCount, &severityInt, &lastEscUnix)
	if err == sql.ErrNoRows {
		// Default-State: keine Eskalation, INFO, FailCount 0, LastEscalation = Unix(0)
		return PersistedState{
			FailCount:       0,
			CurrentSeverity: rules.SeverityInfo,
			LastEscalation:  time.Unix(0, 0),
		}, nil
	}
	if err != nil {
		return PersistedState{}, fmt.Errorf("watcher state: failed to load state: %w", err)
	}

	return PersistedState{
		FailCount:       failCount,
		CurrentSeverity: rules.Severity(severityInt),
		LastEscalation:  time.Unix(lastEscUnix, 0),
	}, nil
}

// Save speichert den gegebenen Zustand atomar (UPSERT auf id = 1).
func (s *sqliteStateStore) Save(state PersistedState) error {
	ctx := context.Background()

	// Unix-Timestamp für LastEscalation
	lastEscUnix := state.LastEscalation.Unix()

	// INSERT OR REPLACE garantiert genau eine Zeile (id = 1).
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO watcher_state (id, fail_count, current_severity, last_escalation)
		 VALUES (1, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET
		   fail_count = excluded.fail_count,
		   current_severity = excluded.current_severity,
		   last_escalation = excluded.last_escalation`,
		state.FailCount, int(state.CurrentSeverity), lastEscUnix,
	)
	if err != nil {
		return fmt.Errorf("watcher state: failed to save state: %w", err)
	}

	return nil
}

// Close schließt die Datenbankverbindung.
func (s *sqliteStateStore) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

