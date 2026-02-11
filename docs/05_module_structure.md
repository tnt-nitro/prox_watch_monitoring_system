# Go-Modulstruktur

## Ziel
Klare Trennung. Wartbar. Testbar.

---

## 1️⃣ Paket-Hierarchie

```
prox-watch/
├─ cmd/
│  └─ prox-watch/
│     └─ main.go              # Entry-Point
├─ internal/
│  ├─ config/                 # Konfiguration
│  ├─ journal/                # systemd journal Reader
│  ├─ pattern/                # Pattern-Engine
│  ├─ state/                  # State-Management (SQLite)
│  ├─ severity/               # Severity-Logik
│  ├─ push/                   # Push-Adapter (ntfy)
│  └─ core/                   # Orchestrierung
├─ pkg/
│  └─ cli/                    # CLI-Kommandos
└─ config/
   └─ config.yaml.example     # Beispiel-Konfiguration
```

---

## 2️⃣ Paket-Verantwortlichkeiten

### `cmd/prox-watch/main.go`

**Verantwortung:**
- Entry-Point
- Signal-Handling (SIGTERM, SIGINT)
- systemd-Integration
- Daemon-Modus
- CLI-Parsing

**Abhängigkeiten:**
- `internal/core`
- `pkg/cli`

---

### `internal/config`

**Verantwortung:**
- Konfiguration laden
- Validierung (keine IPs, keine Pfade im Repo)
- Default-Werte
- Pfad-Auflösung

**Strukturen:**
- `Config` (System, Proxmox, Alerts)
- `Secrets` (Tokens, Keys)

**Funktionen:**
- `Load(path string) (*Config, error)`
- `Validate(c *Config) error`
- `GetDefaultPath() string`

---

### `internal/journal`

**Verantwortung:**
- systemd journald Streaming
- Log-Zeilen lesen
- Filterung (Quellen, Prioritäten)
- Non-blocking

**Strukturen:**
- `Reader` (journal-Reader)
- `Entry` (Log-Eintrag)

**Funktionen:**
- `NewReader() (*Reader, error)`
- `Read() (<-chan Entry, error)`
- `Close() error`

**Wichtig:**
- Keine Log-Speicherung
- Nur Streaming

---

### `internal/pattern`

**Verantwortung:**
- Pattern-Definitionen laden (Metadaten)
- Pattern-Matching
- Event-ID-Generierung
- Lokale Regex-Zuordnung

**Strukturen:**
- `Pattern` (pattern_id, match_type, severity)
- `Matcher` (Pattern-Engine)
- `MatchResult` (event_id, severity)

**Funktionen:**
- `LoadPatterns(path string) ([]Pattern, error)`
- `Match(entry journal.Entry, patterns []Pattern) (*MatchResult, error)`
- `GenerateEventID(pattern Pattern, entry journal.Entry) string`

**Wichtig:**
- Keine Regex im Repo
- Nur Metadaten versioniert

---

### `internal/state`

**Verantwortung:**
- SQLite-Verwaltung
- Event-Zähler
- Zeitfenster-Tracking
- Cooldown-Verwaltung
- Persistenz

**Strukturen:**
- `DB` (SQLite-Verbindung)
- `Event` (event_id, severity, count, timestamps)
- `Cooldown` (event_id, until)

**Funktionen:**
- `Open(path string) (*DB, error)`
- `IncrementCounter(eventID string) error`
- `GetCounter(eventID string) (int, error)`
- `SetCooldown(eventID string, duration time.Duration) error`
- `IsInCooldown(eventID string) (bool, error)`
- `Close() error`

**Tabellen:**
- `events` (event_id, severity, count, first_seen, last_seen)
- `cooldowns` (event_id, cooldown_until)

---

### `internal/severity`

**Verantwortung:**
- Severity-Bewertung
- Zählregeln (1× / 3× / 10×)
- Zeitfenster-Prüfung
- Harte Fehler-Erkennung

**Strukturen:**
- `Evaluator` (Severity-Logik)
- `Rule` (Zählregel, Zeitfenster)

**Funktionen:**
- `Evaluate(eventID string, count int, window time.Duration) (Severity, error)`
- `IsHardError(eventID string) bool`
- `GetWindow(severity Severity) time.Duration`

**Konstanten:**
- `SeverityInfo`, `SeverityWarn`, `SeverityCrit`
- Zählregeln (1, 3, 10)
- Zeitfenster (10 min, 15 min)

---

### `internal/push`

**Verantwortung:**
- ntfy-Integration
- Topic-Mapping
- Metadaten-Versand
- Retry-Logik (optional)

**Strukturen:**
- `Client` (ntfy-Client)
- `Message` (event_id, severity, timestamp)

**Funktionen:**
- `NewClient(server string) (*Client, error)`
- `Send(topic string, message Message) error`
- `GetTopic(severity Severity) string`

**Wichtig:**
- Keine Log-Inhalte
- Nur Metadaten

---

### `internal/core`

**Verantwortung:**
- Orchestrierung
- Event-Loop
- Komponenten-Koordination
- Fehlerbehandlung

**Strukturen:**
- `Watcher` (Haupt-Komponente)
- `Event` (interne Event-Struktur)

**Funktionen:**
- `NewWatcher(config *config.Config) (*Watcher, error)`
- `Run() error`
- `Stop() error`
- `ProcessLogEntry(entry journal.Entry) error`

**Ablauf:**
1. Journal-Reader → Log-Eintrag
2. Pattern-Engine → Event-ID
3. State → Zähler erhöhen
4. Severity → Bewertung
5. Push → Versand (bei Bedarf)

---

### `pkg/cli`

**Verantwortung:**
- CLI-Kommandos
- Wizard (`init`)
- Status (`status`)
- Acknowledge (`ack`)

**Strukturen:**
- `Command` (CLI-Kommando)
- `InitWizard` (Konfigurations-Wizard)

**Funktionen:**
- `RunInit() error`
- `RunStatus() error`
- `RunAck(eventID string) error`
- `AskQuestion(prompt string, default string) (string, error)`
- `ValidateInput(input string) error`

---

## 3️⃣ Abhängigkeiten

### Externe Pakete (Go)

- `github.com/coreos/go-systemd/v22/journal` (journald)
- `github.com/mattn/go-sqlite3` (SQLite)
- `github.com/spf13/cobra` (CLI, optional)
- `net/http` (ntfy)

### Standard-Library

- `context`
- `os/signal`
- `time`
- `encoding/yaml` (Config)
- `fmt`
- `log`

---

## 4️⃣ Datenfluss

```
main.go
  ↓
core.Watcher
  ↓
journal.Reader → pattern.Matcher → state.DB → severity.Evaluator → push.Client
```

**Single-Thread, Event-Loop:**
- Synchroner Ablauf
- Keine Race-Conditions
- Klare Reihenfolge

---

## 5️⃣ Testbarkeit

### Unit-Tests

- Jedes Paket testbar
- Mock-Interfaces möglich
- Isolierte Tests

### Integration-Tests

- `internal/core` als Integration
- Mock journal-Reader
- Mock Push-Client

---

## 6️⃣ Erweiterbarkeit

### Spätere Erweiterungen

- Weitere Pattern-Typen → `internal/pattern`
- Weitere Push-Adapter → `internal/push`
- Weitere Log-Quellen → `internal/journal`

### Offen für:

- Externer Wächter (separates Binary)
- GUI (separates Binary)
- Observability-Add-ons (separate Binaries)

---

## Status

- Modulstruktur definiert
- Verantwortlichkeiten klar
- Go-spezifisch
- MVP-fokussiert
