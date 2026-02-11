# Datenmodelle (State & Enums)

## Ziel
Klar definiert. Persistierbar. Type-safe.

---

## 1️⃣ Enums

### `Severity`

**Typ:**
```go
type Severity int

const (
    SeverityInfo Severity = iota
    SeverityWarn
    SeverityCrit
)
```

**Werte:**
- `SeverityInfo` (0): Informativ, kein Push
- `SeverityWarn` (1): Aufmerksamkeit nötig, Push optional
- `SeverityCrit` (2): Handlungsbedarf, Push immer

**String-Methode:**
```go
func (s Severity) String() string {
    switch s {
    case SeverityInfo:
        return "INFO"
    case SeverityWarn:
        return "WARN"
    case SeverityCrit:
        return "CRIT"
    default:
        return "UNKNOWN"
    }
}
```

---

## 2️⃣ Event-Metadaten

### `Event`

**Typ:**
```go
type Event struct {
    Source    string    // journald, systemd, etc.
    Priority  int       // syslog priority
    Timestamp time.Time
    // Kein Log-Text
    // Keine IPs
    // Keine Hostnames
}
```

**Verwendung:**
- Input für Pattern-Matcher
- Keine Persistierung
- Nur Metadaten

---

## 3️⃣ Pattern-Hit

### `PatternHit`

**Typ:**
```go
type PatternHit struct {
    EventID   string
    PatternID string
    Severity  Severity
    Timestamp time.Time
}
```

**Verwendung:**
- Output von Pattern-Matcher
- Input für State-Store
- Keine Log-Inhalte

---

## 4️⃣ Count State

### `CountState`

**Typ:**
```go
type CountState struct {
    EventID   string
    Severity  Severity
    Count     int
    FirstSeen time.Time
    LastSeen  time.Time
}
```

**Verwendung:**
- Persistiert in SQLite
- Zähler für Event-ID
- Zeitfenster-Tracking

**SQLite-Schema:**
```sql
CREATE TABLE events (
    event_id TEXT PRIMARY KEY,
    severity INTEGER NOT NULL,
    count INTEGER NOT NULL DEFAULT 0,
    first_seen INTEGER NOT NULL,
    last_seen INTEGER NOT NULL
);
```

---

## 5️⃣ Cooldown

### `Cooldown`

**Typ:**
```go
type Cooldown struct {
    EventID string
    Until   time.Time
}
```

**Verwendung:**
- Push-Unterdrückung
- Pro Event-ID
- Standard: 30 Minuten

**SQLite-Schema:**
```sql
CREATE TABLE cooldowns (
    event_id TEXT PRIMARY KEY,
    cooldown_until INTEGER NOT NULL,
    FOREIGN KEY (event_id) REFERENCES events(event_id)
);
```

---

## 6️⃣ Acknowledge

### `Acknowledge`

**Typ:**
```go
type Acknowledge struct {
    EventID string
    Until   time.Time
}
```

**Verwendung:**
- Manuelle Quittierung
- Push-Unterdrückung
- Gültig bis Zeitablauf

**SQLite-Schema:**
```sql
CREATE TABLE acknowledges (
    event_id TEXT PRIMARY KEY,
    ack_until INTEGER NOT NULL,
    FOREIGN KEY (event_id) REFERENCES events(event_id)
);
```

---

## 7️⃣ Push Message

### `PushMessage`

**Typ:**
```go
type PushMessage struct {
    EventID   string
    Severity  Severity
    Timestamp time.Time
    // Kein Log-Text
    // Keine IPs
    // Keine Hostnames
}
```

**Verwendung:**
- Input für Push-Adapter
- Metadaten-Only
- Keine sensiblen Daten

---

## 8️⃣ Config Models

### `Config`

**Typ:**
```go
type Config struct {
    System  SystemConfig
    Proxmox ProxmoxConfig
    Alerts  AlertsConfig
    Paths   PathsConfig
}

type SystemConfig struct {
    Timezone string
}

type ProxmoxConfig struct {
    Host     string // Platzhalter, keine IP
    APIPort  int
}

type AlertsConfig struct {
    Channel string // "local-only" | "ntfy"
}

type PathsConfig struct {
    StateDB  string
    Config   string
    Secrets  string
}
```

**Validierung:**
- Keine IPs
- Keine realen Hostnames
- Keine Pfade im Repo

---

## 9️⃣ Pattern Definition

### `Pattern`

**Typ:**
```go
type Pattern struct {
    PatternID string
    Source    string
    MatchType MatchType
    Severity  Severity
    CountRule CountRule
}

type MatchType int

const (
    MatchTypeKeyword MatchType = iota
    MatchTypeRegex
    MatchTypeEvent
)

type CountRule struct {
    Threshold int
    Window    time.Duration
}
```

**Verwendung:**
- Metadaten im Repo
- Keine Regex-Texte
- Lokale Zuordnung erlaubt

---

## 🔟 State Store Models

### `State`

**Typ:**
```go
type State struct {
    Events      map[string]CountState
    Cooldowns   map[string]Cooldown
    Acknowledges map[string]Acknowledge
}
```

**Verwendung:**
- In-Memory-Cache
- Persistiert in SQLite
- Thread-safe (SQLite intern)

---

## 1️⃣1️⃣ Event-ID Schema

### Format

**Schema:**
```
<DOMAIN>.<CATEGORY>.<EVENT>
```

**Beispiele:**
- `host.network.link_down`
- `host.kernel.nic_reset`
- `scheduler.cron.missed`
- `container.lifecycle.crash`

**Regeln:**
- Keine IPs
- Keine Hostnames
- Keine Pfade
- Nur abstrakte Bezeichner

---

## 1️⃣2️⃣ SQLite Schema (vollständig)

```sql
-- Events
CREATE TABLE events (
    event_id TEXT PRIMARY KEY,
    severity INTEGER NOT NULL,
    count INTEGER NOT NULL DEFAULT 0,
    first_seen INTEGER NOT NULL,
    last_seen INTEGER NOT NULL
);

-- Cooldowns
CREATE TABLE cooldowns (
    event_id TEXT PRIMARY KEY,
    cooldown_until INTEGER NOT NULL,
    FOREIGN KEY (event_id) REFERENCES events(event_id)
);

-- Acknowledges
CREATE TABLE acknowledges (
    event_id TEXT PRIMARY KEY,
    ack_until INTEGER NOT NULL,
    FOREIGN KEY (event_id) REFERENCES events(event_id)
);

-- Indizes
CREATE INDEX idx_events_severity ON events(severity);
CREATE INDEX idx_events_last_seen ON events(last_seen);
CREATE INDEX idx_cooldowns_until ON cooldowns(cooldown_until);
CREATE INDEX idx_acknowledges_until ON acknowledges(ack_until);
```

---

## Status

- Datenmodelle definiert
- Type-safe
- Persistierbar
- Datenschutz-konform
