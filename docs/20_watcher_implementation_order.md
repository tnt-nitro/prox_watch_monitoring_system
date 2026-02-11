# Phase 1 Implementierungsreihenfolge (Watcher)

## Status

**Phase 1 - Spezifikation** 🟡

## Ziel

Strukturierte, sequentielle Implementierung des Watchers.

## Implementierungsreihenfolge

### 1️⃣ Counter implementieren

**Datei:** `internal/watcher/counter.go`

**Aufgabe:**
- In-Memory Counter (keine Persistenz)
- `Increment()`, `GetCount()`, `Reset()`
- Thread-safe (Single-Thread, aber defensiv)

**Tests:**
- Increment/Reset
- GetCount
- Thread-Safety (optional)

**Abbruchkriterien:**
- Nicht-deterministisches Verhalten
- Race-Conditions

---

### 2️⃣ Severity-Integration

**Datei:** `internal/watcher/severity.go` (neu)

**Aufgabe:**
- Severity-Evaluierung basierend auf FailCount
- Nutzt `internal/rules` (shared package)
- Thresholds aus Config

**Tests:**
- Threshold-Grenzen (2→3, 9→10)
- Severity-Ordnung (INFO < WARN < CRIT)

**Abbruchkriterien:**
- Abweichung von Spezifikation
- IO-Abhängigkeit

---

### 3️⃣ HealthChecker (Ping + HTTPS minimal)

**Datei:** `internal/watcher/health.go`

**Aufgabe:**
- Ping-Check (ICMP)
- HTTPS-Check (Port 8006)
- Kombinationslogik (ping/https/ping+https)
- Timeout-Handling

**Dependencies:**
- `golang.org/x/net/icmp` (Ping)
- `net/http` (HTTPS)

**Tests:**
- Ping-Check (Success/Failure)
- HTTPS-Check (Success/Failure)
- Kombinationslogik
- Timeout-Verhalten

**Abbruchkriterien:**
- Host/IP in Logs
- Blocking-Operationen
- Keine Timeout-Behandlung

---

### 4️⃣ Push-Integration

**Datei:** `internal/watcher/runner.go` (teilweise)

**Aufgabe:**
- Push-Adapter aus `internal/push` nutzen
- Push nur bei Statuswechsel nach oben
- Event-ID: `external.availability.down`
- Topic-Mapping (WARN/CRIT)

**Tests:**
- Push bei Statuswechsel (INFO→WARN, WARN→CRIT)
- Kein Push bei gleicher Severity
- Kein Push bei Verbesserung

**Abbruchkriterien:**
- Push-Spam
- Host/IP in Push-Nachrichten

---

### 5️⃣ GPIO NoOp

**Datei:** `internal/watcher/gpio.go`

**Aufgabe:**
- NoOp-GPIO implementieren
- Interface erfüllen
- Immer erfolgreich

**Tests:**
- NoOp-Verhalten (SetLED, Beep, Close)
- Keine Fehler

**Abbruchkriterien:**
- Hardware-Abhängigkeit
- Blocking-Operationen

---

### 6️⃣ Runner implementieren

**Datei:** `internal/watcher/runner.go`

**Aufgabe:**
- Event-Loop (Ticker)
- Health Check → Counter → Severity → Push → GPIO
- Graceful Shutdown (Context-Cancellation)
- State-Management

**Tests:**
- Event-Loop (einzelne Iteration)
- Graceful Shutdown
- State-Updates

**Abbruchkriterien:**
- Race-Conditions
- Blocking-Operationen
- Kein Graceful Shutdown

---

### 7️⃣ Tests

**Dateien:** `internal/watcher/*_test.go`

**Aufgabe:**
- Unit-Tests für alle Komponenten
- Integration-Tests für Runner
- Fake-Implementierungen (HealthChecker, GPIO)

**Coverage-Ziel:**
- Core-Logik ≥ 90%
- Guards ≥ 100%

**Abbruchkriterien:**
- Reale Daten in Tests
- Externe Abhängigkeiten

---

### 8️⃣ Dokumentation ergänzen

**Dateien:** `cmd/watcher/README.md`, `docs/`

**Aufgabe:**
- README aktualisieren
- Installationsanleitung (später)
- Troubleshooting (später)

**Abbruchkriterien:**
- Sensible Daten in Docs
- Unvollständige Dokumentation

---

## Abhängigkeiten

```
Counter
  └─ Severity (nutzt internal/rules)
      └─ HealthChecker
          └─ Push (nutzt internal/push)
              └─ GPIO (NoOp)
                  └─ Runner (orchestriert alles)
```

## Test-Strategie

### Unit-Tests

- **Counter:** Increment/Reset/GetCount
- **Severity:** Threshold-Grenzen
- **HealthChecker:** Ping/HTTPS/Kombination
- **GPIO:** NoOp-Verhalten

### Integration-Tests

- **Runner:** End-to-End Event-Flow
- **Push:** Statuswechsel → Push
- **GPIO:** Severity → LED

### Fake-Implementierungen

- **FakeHealthChecker:** Synthetische Results
- **FakeGPIO:** In-Memory State

## Nächste Schritte

1. Task 1 starten: Counter implementieren
2. Sequentiell durch alle Tasks
3. Tests nach jedem Task
