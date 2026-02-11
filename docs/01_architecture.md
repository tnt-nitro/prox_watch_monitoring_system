# Architektur

## Komponenten

### 1. Journal Reader

**Verantwortung:**
- systemd journald Streaming
- Log-Einträge lesen (nur Metadaten)
- Non-blocking

**Implementierung:**
- `internal/journal/reader.go` - Interface
- `internal/journal/fake.go` - Fake für Tests

### 2. Pattern Matcher

**Verantwortung:**
- Pattern-Definitionen laden (Metadaten)
- Event-ID-Generierung
- Pattern-Matching gegen Journal-Entries

**Implementierung:**
- `internal/pattern/matcher.go` - Pattern Matching
- `internal/pattern/registry.go` - Pattern-Registry

### 3. State Store

**Verantwortung:**
- SQLite-basierte Persistenz
- Event-Zähler
- Cooldown-Verwaltung
- Acknowledge-Verwaltung

**Implementierung:**
- `internal/state/store.go` - SQLite-Store

### 4. Severity Evaluator

**Verantwortung:**
- Zählregeln anwenden (1× / 3× / 10×)
- Zeitfenster-Prüfung
- Harte Fehler-Erkennung

**Implementierung:**
- `internal/rules/evaluator.go` - Severity-Bewertung

### 5. Push Adapter

**Verantwortung:**
- ntfy-Integration (optional)
- Topic-Mapping
- Metadaten-Versand

**Implementierung:**
- `internal/push/ntfy.go` - ntfy-Adapter
- `internal/push/adapter.go` - Interface

### 6. Core Runner

**Verantwortung:**
- Orchestrierung
- Event-Loop
- Komponenten-Koordination

**Implementierung:**
- `internal/core/runner.go` - Event-Loop
- `internal/core/lifecycle.go` - Lifecycle-Management

## Datenfluss

### Event-Processing-Pipeline

```
Journal Entry
    ↓
Pattern Matcher → EventID
    ↓
State Store → Increment Counter
    ↓
Severity Evaluator → Severity (INFO/WARN/CRIT)
    ↓
Cooldown/Ack Check → Skip if active
    ↓
Push Adapter → Send Notification (if CRIT/WARN)
    ↓
State Store → Set Cooldown
```

### Single-Thread Event-Loop

- Synchroner Ablauf
- Keine Race-Conditions
- Deterministische Reihenfolge

## Fehlerszenarien

### Journal-Reader-Fehler

- **Verhalten:** Continue on error
- **Logging:** Fehler wird geloggt, Processing geht weiter

### Pattern-Matching-Fehler

- **Verhalten:** No match → Skip
- **Logging:** Kein Fehler, Event wird ignoriert

### State-Store-Fehler

- **Verhalten:** Fehler wird geloggt, Processing stoppt
- **Recovery:** Service-Restart erforderlich

### Push-Fehler

- **Verhalten:** Non-blocking, Fehler wird geloggt
- **Recovery:** Nächster Push-Versuch beim nächsten Event

### Config-Validierungs-Fehler

- **Verhalten:** Service startet nicht
- **Recovery:** Config korrigieren, Service neu starten
