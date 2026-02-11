# Schnittstellen (Interfaces)

## Ziel
Klare Verträge. Testbar. Ersetzbar.

---

## 1️⃣ Journal Reader Interface

### `journal.Reader`

**Verantwortung:**
- Log-Einträge aus systemd journal lesen
- Streaming (keine Speicherung)
- Read-only

**Interface:**
```go
type Reader interface {
    // Read startet das Streaming
    Read(ctx context.Context) (<-chan Entry, error)
    
    // Close beendet das Streaming
    Close() error
}

type Entry struct {
    // Nur Metadaten, kein Log-Text
    Priority    int
    Source      string
    Timestamp   time.Time
    // Text wird NICHT gespeichert
}
```

**Implementierung:**
- `internal/journal/reader.go` (systemd journal)

**Wichtig:**
- Keine Log-Inhalte in Entry
- Nur Metadaten für Pattern-Matching

---

## 2️⃣ Pattern Matcher Interface

### `pattern.Matcher`

**Verantwortung:**
- Pattern-Matching gegen Log-Einträge
- Event-ID-Generierung
- Abstrakt (keine Regex im Repo)

**Interface:**
```go
type Matcher interface {
    // Match prüft Eintrag gegen Patterns
    Match(ctx context.Context, entry journal.Entry) (*MatchResult, error)
    
    // LoadPatterns lädt Pattern-Definitionen (Metadaten)
    LoadPatterns(path string) error
}

type MatchResult struct {
    EventID   string
    Severity  severity.Level
    PatternID string
}
```

**Implementierung:**
- `internal/pattern/matcher.go`

**Wichtig:**
- Keine Log-Inhalte in MatchResult
- Nur Event-ID und Severity

---

## 3️⃣ State Store Interface

### `state.Store`

**Verantwortung:**
- Event-Zähler verwalten
- Cooldown-Verwaltung
- Persistenz (SQLite)

**Interface:**
```go
type Store interface {
    // IncrementCounter erhöht Zähler für Event-ID
    IncrementCounter(ctx context.Context, eventID string) error
    
    // GetCounter liefert aktuellen Zähler
    GetCounter(ctx context.Context, eventID string) (int, error)
    
    // GetEvent liefert vollständiges Event
    GetEvent(ctx context.Context, eventID string) (*Event, error)
    
    // SetCooldown setzt Cooldown für Event-ID
    SetCooldown(ctx context.Context, eventID string, duration time.Duration) error
    
    // IsInCooldown prüft, ob Event-ID in Cooldown
    IsInCooldown(ctx context.Context, eventID string) (bool, error)
    
    // Acknowledge setzt Acknowledge für Event-ID
    Acknowledge(ctx context.Context, eventID string, until time.Time) error
    
    // IsAcknowledged prüft, ob Event-ID quittiert
    IsAcknowledged(ctx context.Context, eventID string) (bool, error)
    
    // Close schließt Verbindung
    Close() error
}

type Event struct {
    EventID   string
    Severity  severity.Level
    Count     int
    FirstSeen time.Time
    LastSeen  time.Time
}
```

**Implementierung:**
- `internal/state/store.go` (SQLite)

**Wichtig:**
- Thread-safe (SQLite intern)
- Persistenz über Neustarts

---

## 4️⃣ Push Adapter Interface

### `push.Adapter`

**Verantwortung:**
- Push-Nachrichten versenden
- Topic-Mapping
- Metadaten-Only

**Interface:**
```go
type Adapter interface {
    // Send versendet Nachricht an Topic
    Send(ctx context.Context, topic string, message Message) error
    
    // GetTopic liefert Topic für Severity
    GetTopic(severity severity.Level) string
}

type Message struct {
    EventID   string
    Severity  severity.Level
    Timestamp time.Time
    // Kein Log-Text
    // Keine IPs
    // Keine Hostnames
}
```

**Implementierung:**
- `internal/push/ntfy.go` (ntfy)

**Wichtig:**
- Einweg (keine Rückkanäle)
- Nur Metadaten
- Keine Log-Inhalte

---

## 5️⃣ Severity Evaluator Interface

### `rules.Evaluator`

**Verantwortung:**
- Severity-Bewertung
- Zählregeln anwenden
- Zeitfenster prüfen

**Interface:**
```go
type Evaluator interface {
    // Evaluate bewertet Event basierend auf Zähler und Zeitfenster
    Evaluate(ctx context.Context, eventID string, count int, window time.Duration) (Level, error)
    
    // IsHardError prüft, ob Event harter Fehler (sofort CRIT)
    IsHardError(eventID string) bool
    
    // GetWindow liefert Zeitfenster für Severity
    GetWindow(severity Level) time.Duration
    
    // GetThreshold liefert Schwellwert für Severity
    GetThreshold(severity Level) int
}

type Level int

const (
    LevelInfo Level = iota
    LevelWarn
    LevelCrit
)
```

**Implementierung:**
- `internal/rules/severity.go`

**Wichtig:**
- Deterministisch
- Keine IO-Operationen

---

## 6️⃣ Config Loader Interface

### `config.Loader`

**Verantwortung:**
- Konfiguration laden
- Validierung
- Datenschutz-Guards

**Interface:**
```go
type Loader interface {
    // Load lädt Konfiguration aus Pfad
    Load(path string) (*Config, error)
    
    // Validate prüft Konfiguration
    Validate(c *Config) error
    
    // CheckRepoPath prüft, ob Pfad im Repo liegt
    CheckRepoPath(path string) error
    
    // CheckIPs prüft, ob IPs enthalten
    CheckIPs(c *Config) error
}

type Config struct {
    System  SystemConfig
    Proxmox ProxmoxConfig
    Alerts  AlertsConfig
    Paths   PathsConfig
}
```

**Implementierung:**
- `internal/config/loader.go`

**Wichtig:**
- Harte Validierung
- Datenschutz-Guards

---

## 7️⃣ Core Runner Interface

### `core.Runner`

**Verantwortung:**
- Orchestrierung
- Event-Loop
- Komponenten-Koordination

**Interface:**
```go
type Runner interface {
    // Run startet Event-Loop
    Run(ctx context.Context) error
    
    // Stop beendet Event-Loop graceful
    Stop() error
    
    // ProcessEntry verarbeitet Log-Eintrag
    ProcessEntry(ctx context.Context, entry journal.Entry) error
}
```

**Implementierung:**
- `internal/core/runner.go`

**Wichtig:**
- Single-Thread
- Deterministischer Ablauf

---

## 8️⃣ Abhängigkeits-Graph

```
core.Runner
  ├─ journal.Reader
  ├─ pattern.Matcher
  ├─ state.Store
  ├─ rules.Evaluator
  └─ push.Adapter
```

**Alle Interfaces:**
- Testbar (Mocks möglich)
- Ersetzbar (andere Implementierungen)
- Klare Verträge

---

## 9️⃣ Mock-Fähigkeit

### Für Tests

- `journal.MockReader`
- `pattern.MockMatcher`
- `state.MockStore`
- `push.MockAdapter`

**Vorteil:**
- Unit-Tests isoliert
- Integration-Tests kontrolliert
- Keine echten Dependencies in Tests

---

## Status

- Interfaces definiert
- Verträge klar
- Testbar
- Ersetzbar
