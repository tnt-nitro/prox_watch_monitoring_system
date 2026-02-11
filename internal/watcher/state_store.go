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
// Phase 3: Erweitert um PowerAttempts, LastPowerAttempt.
type PersistedState struct {
	FailCount        int
	CurrentSeverity  rules.Severity
	LastEscalation   time.Time
	PowerAttempts    int       // Phase 3: Anzahl Power-Cycle-Versuche
	LastPowerAttempt time.Time // Phase 3: Timestamp des letzten Power-Cycle-Versuchs
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
// Phase 3: Erweitert um power_attempts und last_power_attempt.
func (s *sqliteStateStore) initSchema() error {
	// Erstelle Tabelle (falls nicht vorhanden)
	schema := `
CREATE TABLE IF NOT EXISTS watcher_state (
	id INTEGER PRIMARY KEY CHECK (id = 1),
	fail_count INTEGER NOT NULL,
	current_severity INTEGER NOT NULL,
	last_escalation INTEGER NOT NULL,
	power_attempts INTEGER NOT NULL DEFAULT 0,
	last_power_attempt INTEGER NOT NULL DEFAULT 0
);`

	if _, err := s.db.Exec(schema); err != nil {
		return fmt.Errorf("watcher state: failed to create schema: %w", err)
	}

	// Phase 3: Migriere bestehende Tabellen (füge Spalten hinzu, falls nicht vorhanden)
	// SQLite unterstützt kein IF NOT EXISTS für ALTER TABLE, daher ignorieren wir Fehler
	migration1 := `ALTER TABLE watcher_state ADD COLUMN power_attempts INTEGER NOT NULL DEFAULT 0;`
	migration2 := `ALTER TABLE watcher_state ADD COLUMN last_power_attempt INTEGER NOT NULL DEFAULT 0;`
	// Ignoriere Fehler (Spalten existieren bereits oder Tabelle existiert nicht)
	_ = s.db.Exec(migration1)
	_ = s.db.Exec(migration2)

	return nil
}

// Load lädt den persistenten Zustand.
// Wenn keine Zeile existiert, wird ein Default-State zurückgegeben.
// Phase 3: Erweitert um PowerAttempts, LastPowerAttempt.
func (s *sqliteStateStore) Load() (PersistedState, error) {
	ctx := context.Background()

	row := s.db.QueryRowContext(ctx,
		"SELECT fail_count, current_severity, last_escalation, power_attempts, last_power_attempt FROM watcher_state WHERE id = 1",
	)

	var failCount int
	var severityInt int
	var lastEscUnix int64
	var powerAttempts int
	var lastPowerAttemptUnix int64

	err := row.Scan(&failCount, &severityInt, &lastEscUnix, &powerAttempts, &lastPowerAttemptUnix)
	if err == sql.ErrNoRows {
		// Default-State: keine Eskalation, INFO, FailCount 0, LastEscalation = Unix(0)
		return PersistedState{
			FailCount:        0,
			CurrentSeverity:  rules.SeverityInfo,
			LastEscalation:   time.Unix(0, 0),
			PowerAttempts:    0,
			LastPowerAttempt: time.Unix(0, 0),
		}, nil
	}
	if err != nil {
		return PersistedState{}, fmt.Errorf("watcher state: failed to load state: %w", err)
	}

	return PersistedState{
		FailCount:        failCount,
		CurrentSeverity:  rules.Severity(severityInt),
		LastEscalation:   time.Unix(lastEscUnix, 0),
		PowerAttempts:    powerAttempts,
		LastPowerAttempt: time.Unix(lastPowerAttemptUnix, 0),
	}, nil
}

// Save speichert den gegebenen Zustand atomar (UPSERT auf id = 1).
// Phase 3: Erweitert um PowerAttempts, LastPowerAttempt.
func (s *sqliteStateStore) Save(state PersistedState) error {
	ctx := context.Background()

	// Unix-Timestamps
	lastEscUnix := state.LastEscalation.Unix()
	lastPowerAttemptUnix := state.LastPowerAttempt.Unix()

	// INSERT OR REPLACE garantiert genau eine Zeile (id = 1).
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO watcher_state (id, fail_count, current_severity, last_escalation, power_attempts, last_power_attempt)
		 VALUES (1, ?, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET
		   fail_count = excluded.fail_count,
		   current_severity = excluded.current_severity,
		   last_escalation = excluded.last_escalation,
		   power_attempts = excluded.power_attempts,
		   last_power_attempt = excluded.last_power_attempt`,
		state.FailCount, int(state.CurrentSeverity), lastEscUnix, state.PowerAttempts, lastPowerAttemptUnix,
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

