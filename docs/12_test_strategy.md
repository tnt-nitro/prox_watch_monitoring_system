# Teststrategie (ohne reale Logs)

## Ziel
Deterministisch. Reproduzierbar. Datenschutz-konform.

---

## 1️⃣ Test-Prinzipien

### Keine realen Daten

**Verboten:**
- Reale Log-Zeilen
- Reale IP-Adressen
- Reale Hostnames
- Reale Pfade
- Reale Container-IDs

**Erlaubt:**
- Synthetische Daten
- Platzhalter
- Mock-Objekte
- Fakes

---

## 2️⃣ Test-Typen

### Unit-Tests

**Ziel:**
- Isolierte Komponenten
- Mock-Dependencies
- Schnelle Ausführung

**Bereiche:**
- Pattern-Matching
- Severity-Evaluator
- State-Store
- Config-Validierung
- Push-Adapter

---

### Integration-Tests

**Ziel:**
- Komponenten-Integration
- End-to-End-Flows
- Mock-Interfaces

**Bereiche:**
- Core-Runner
- Event-Flow
- State-Persistenz

---

### Smoke-Tests

**Ziel:**
- Basis-Funktionalität
- Start/Stop
- Konfiguration

**Bereiche:**
- Service-Start
- Config-Laden
- State-DB-Zugriff

---

## 3️⃣ Mock-Objekte

### Journal-Reader Mock

**Interface:**
```go
type MockJournalReader struct {
    Events []Event
    Index  int
}

func (m *MockJournalReader) Next() (Event, error) {
    if m.Index >= len(m.Events) {
        return Event{}, io.EOF
    }
    event := m.Events[m.Index]
    m.Index++
    return event, nil
}
```

**Verwendung:**
- Synthetische Events
- Keine realen Log-Zeilen
- Kontrollierbare Sequenzen

---

### Pattern-Matcher Mock

**Interface:**
```go
type MockPatternMatcher struct {
    Hits map[string]PatternHit
}

func (m *MockPatternMatcher) Match(e Event) (PatternHit, bool) {
    hit, ok := m.Hits[e.Source]
    return hit, ok
}
```

**Verwendung:**
- Vordefinierte Treffer
- Testbare Szenarien
- Keine Regex-Ausführung

---

### State-Store Mock

**Interface:**
```go
type MockStateStore struct {
    Counters map[string]int
    Cooldowns map[string]time.Time
    Acknowledges map[string]time.Time
}

func (m *MockStateStore) Increment(eventID string, ts time.Time) (CountState, error) {
    m.Counters[eventID]++
    return CountState{
        EventID: eventID,
        Count: m.Counters[eventID],
    }, nil
}
```

**Verwendung:**
- In-Memory-State
- Keine SQLite-Abhängigkeit
- Schnelle Tests

---

### Push-Adapter Mock

**Interface:**
```go
type MockPushAdapter struct {
    Messages []PushMessage
}

func (m *MockPushAdapter) Send(msg PushMessage) error {
    m.Messages = append(m.Messages, msg)
    return nil
}
```

**Verwendung:**
- Keine Netzwerk-Zugriffe
- Nachrichten sammeln
- Verifizierbar

---

## 4️⃣ Test-Daten

### Synthetische Events

**Beispiele:**
```go
var testEvents = []Event{
    {
        Source:    "systemd",
        Priority:  3, // ERR
        Timestamp: time.Now(),
    },
    {
        Source:    "kernel",
        Priority:  4, // WARNING
        Timestamp: time.Now().Add(1 * time.Minute),
    },
}
```

**Regeln:**
- Keine realen Log-Inhalte
- Nur Metadaten
- Platzhalter-Quellen

---

### Synthetische Pattern-Hits

**Beispiele:**
```go
var testPatternHits = map[string]PatternHit{
    "test_event_1": {
        EventID:   "host.network.link_down",
        PatternID: "pattern_1",
        Severity:  SeverityCrit,
    },
}
```

**Regeln:**
- Abstrakte Event-IDs
- Keine realen Referenzen
- Testbare Szenarien

---

## 5️⃣ Test-Szenarien

### Szenario 1: Einfacher Treffer

**Setup:**
- Mock-Reader: 1 Event
- Mock-Matcher: 1 Treffer
- Mock-Store: leer

**Erwartung:**
- Event verarbeitet
- Zähler = 1
- Severity = INFO (1×)

---

### Szenario 2: Schwellwert erreicht

**Setup:**
- Mock-Reader: 3 Events (gleiche Event-ID)
- Mock-Matcher: 3 Treffer
- Mock-Store: Zähler = 2

**Erwartung:**
- Zähler = 3
- Severity = WARN (3× in 10 min)
- Push optional

---

### Szenario 3: Kritischer Fehler

**Setup:**
- Mock-Reader: 1 Event (harter Fehler)
- Mock-Matcher: 1 Treffer (CRIT)
- Mock-Store: leer

**Erwartung:**
- Zähler = 1
- Severity = CRIT (sofort)
- Push immer

---

### Szenario 4: Cooldown aktiv

**Setup:**
- Mock-Reader: 1 Event
- Mock-Matcher: 1 Treffer
- Mock-Store: Cooldown aktiv

**Erwartung:**
- Zähler erhöht
- Push unterdrückt
- Cooldown respektiert

---

### Szenario 5: Zeitfenster abgelaufen

**Setup:**
- Mock-Reader: 1 Event
- Mock-Matcher: 1 Treffer
- Mock-Store: Zähler = 3, Window abgelaufen

**Erwartung:**
- Zähler zurückgesetzt
- Neuer Zähler = 1
- Severity = INFO

---

## 6️⃣ Test-Frameworks

### Go-Testing

**Standard:**
- `testing` Package
- `go test`
- Table-Driven Tests

**Beispiel:**
```go
func TestSeverityEvaluator(t *testing.T) {
    tests := []struct {
        name     string
        count    int
        window   time.Duration
        expected Severity
    }{
        {"info_1x", 1, 0, SeverityInfo},
        {"warn_3x", 3, 10 * time.Minute, SeverityWarn},
        {"crit_10x", 10, 15 * time.Minute, SeverityCrit},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test-Logik
        })
    }
}
```

---

### Testify (optional)

**Vorteile:**
- Assertions
- Mocks
- Suites

**Verwendung:**
- Komplexere Tests
- Mock-Generierung

---

## 7️⃣ Test-Coverage

### Ziel-Coverage

**Minimum:**
- Core-Logik: 80%
- Validierung: 90%
- State-Store: 85%

**Ausnahmen:**
- Main-Funktion: niedrig
- CLI-Wizard: niedrig (interaktiv)

---

### Coverage-Report

**Kommando:**
```bash
go test -cover -coverprofile=coverage.out
go tool cover -html=coverage.out
```

**Ausgabe:**
- HTML-Report
- Zeilen-Coverage
- Branch-Coverage

---

## 8️⃣ Test-Datenbank

### SQLite in-Memory

**Verwendung:**
```go
db, err := sql.Open("sqlite3", ":memory:")
```

**Vorteile:**
- Keine Dateien
- Schnell
- Isoliert

**Nachteile:**
- Keine Persistenz-Tests
- Separater Test nötig

---

### SQLite Temp-Datei

**Verwendung:**
```go
tmpfile, _ := os.CreateTemp("", "test-*.db")
defer os.Remove(tmpfile.Name())
db, err := sql.Open("sqlite3", tmpfile.Name())
```

**Vorteile:**
- Persistenz testbar
- Realistische Bedingungen

**Nachteile:**
- Datei-System nötig
- Cleanup erforderlich

---

## 9️⃣ Test-Isolation

### Cleanup

**Regeln:**
- Jeder Test isoliert
- Keine Seiteneffekte
- Cleanup nach Test

**Beispiel:**
```go
func TestStateStore(t *testing.T) {
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)
    
    // Test-Logik
}
```

---

### Parallel-Tests

**Regeln:**
- Tests parallelisierbar
- Keine Shared-State
- Thread-safe Mocks

**Beispiel:**
```go
func TestParallel(t *testing.T) {
    t.Parallel()
    // Test-Logik
}
```

---

## 🔟 Test-Kategorien

### Fast Tests

**Kriterien:**
- < 1 Sekunde
- Keine I/O
- In-Memory

**Tags:**
```go
//go:build fast
```

---

### Integration Tests

**Kriterien:**
- > 1 Sekunde
- I/O möglich
- Externe Dependencies

**Tags:**
```go
//go:build integration
```

---

## 1️⃣1️⃣ Test-Validierung

### Datenschutz-Checks

**Prüfungen:**
- Keine IPs in Tests
- Keine realen Hostnames
- Keine realen Pfade
- Keine Log-Inhalte

**Tool:**
- Linter-Regel
- Pre-Commit-Hook

---

## 1️⃣2️⃣ CI/CD-Integration

### Test-Ausführung

**Pipeline:**
1. Unit-Tests
2. Integration-Tests
3. Coverage-Report
4. Linter-Checks

**Fehlerbehandlung:**
- Test-Fehler → Pipeline-Fehler
- Coverage < Ziel → Warnung

---

## Status

- Teststrategie definiert
- Mock-Objekte spezifiziert
- Szenarien beschrieben
- Datenschutz gewährleistet
