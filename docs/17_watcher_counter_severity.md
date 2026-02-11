# Counter & Severity-Flow (Watcher Phase 1)

## Status

**Phase 1 - Spezifikation** 🟡

## Ziel

- ✅ Fehlversuche zählen
- ✅ Severity bestimmen
- ✅ Push auslösen
- ✅ Reset bei Erfolg

**Nicht enthalten:**
- ❌ SQLite-Persistenz (Phase 1: In-Memory)
- ❌ Cooldown-Mechanismus (optional später)
- ❌ Acknowledge-Funktion (optional später)

## Counter-Modell

### Struktur

```go
type Counter struct {
    FailCount int
}
```

**Eigenschaften:**
- **Kein Persistenzspeicher** (Phase 1: In-Memory)
- **Reset bei Success** (Health Check erfolgreich)
- **Minimal** (nur FailCount, keine Zeitstempel in Phase 1)

### Interface

```go
type Counter interface {
    Increment() error
    GetCount() int
    Reset() error
}
```

**Methoden:**
- `Increment()` - Erhöht FailCount um 1
- `GetCount()` - Gibt aktuellen FailCount zurück
- `Reset()` - Setzt FailCount auf 0

## Ablauf pro Intervall

```
Health Check → Result

IF Success:
    FailCount = 0
    Severity = INFO
ELSE:
    FailCount++
    IF FailCount >= crit_threshold:
        Severity = CRIT
    ELSE IF FailCount >= warn_threshold:
        Severity = WARN
    ELSE:
        Severity = INFO
```

### Detaillierter Ablauf

1. **Health Check ausführen:**
   ```go
   result := healthChecker.Check(ctx)
   ```

2. **Erfolg (Success = true):**
   ```go
   counter.Reset()
   newSeverity = rules.SeverityInfo
   ```

3. **Fehler (Success = false):**
   ```go
   counter.Increment()
   failCount := counter.GetCount()
   
   if failCount >= critThreshold {
       newSeverity = rules.SeverityCrit
   } else if failCount >= warnThreshold {
       newSeverity = rules.SeverityWarn
   } else {
       newSeverity = rules.SeverityInfo
   }
   ```

4. **Severity-Evaluierung:**
   - Nutzt `internal/rules` (shared package)
   - Thresholds aus Config: `thresholds.warn`, `thresholds.crit`

## Event-ID

### Fest definiert

```
external.availability.down
```

**Eigenschaften:**
- **Kein dynamisches Event** (immer gleich)
- **Keine Host-Referenz** (anonym)
- **Eindeutig** (nur ein Event-Typ im Watcher)

### Verwendung

Die Event-ID wird für Push-Benachrichtigungen verwendet:

```go
pushMessage := push.PushMessage{
    EventID:  "external.availability.down",
    Severity: newSeverity,
    At:       time.Now(),
}
```

## Push-Regeln

### Severity → Push-Mapping

| Severity | Push |
|----------|------|
| INFO | ❌ Kein Push |
| WARN | ✔️ Push an `prox-watch-warn` |
| CRIT | ✔️ Push an `prox-watch-crit` |

### Push-Trigger (Statuswechsel)

**Push nur bei Statuswechsel:**

- `INFO → WARN` → Push auslösen
- `WARN → CRIT` → Push auslösen
- `CRIT → CRIT` → **Kein Push** (kein Statuswechsel)
- `WARN → WARN` → **Kein Push** (kein Statuswechsel)
- `INFO → INFO` → **Kein Push** (kein Statuswechsel)

**Reset-Szenario:**
- `WARN → INFO` → **Kein Push** (Verbesserung, kein Alarm)
- `CRIT → INFO` → **Kein Push** (Verbesserung, kein Alarm)

### Cooldown (Phase 1)

**Phase 1: Kein Cooldown**

- Cooldown-Mechanismus ist **nicht** Teil von Phase 1
- Push wird bei jedem Statuswechsel ausgelöst
- Optional später: Cooldown für wiederholte CRIT-Pushes

## Sicherheitsregeln

### Verboten

- ❌ **Kein Push bei jedem Intervall:**
  - Push nur bei Statuswechsel (siehe oben)

- ❌ **Kein Push-Spam:**
  - Keine wiederholten Pushes bei gleicher Severity
  - Keine Pushes bei Verbesserung (CRIT→INFO, WARN→INFO)

### Erlaubt

- ✔ **Push bei Statuswechsel** (INFO→WARN, WARN→CRIT)
- ✔ **Push-Metadaten** (EventID, Severity, Timestamp)
- ✔ **Keine Host/IP** in Push-Nachrichten

## Zustandsmodell

### State-Struktur

```go
type State struct {
    CurrentSeverity rules.Severity
    FailCount       int
}
```

**Eigenschaften:**
- **In-Memory** (Phase 1: keine Persistenz)
- **Aktueller Zustand** (CurrentSeverity, FailCount)
- **Vergleichslogik** (NewSeverity > CurrentSeverity)

### Push-Logik

```go
// Pseudocode
if newSeverity > state.CurrentSeverity {
    // Statuswechsel nach oben (INFO→WARN, WARN→CRIT)
    pushAdapter.Send(pushMessage)
    state.CurrentSeverity = newSeverity
} else if newSeverity < state.CurrentSeverity {
    // Verbesserung (CRIT→INFO, WARN→INFO)
    state.CurrentSeverity = newSeverity
    // Kein Push
} else {
    // Gleiche Severity (WARN→WARN, CRIT→CRIT)
    // Kein Push
}
```

### Severity-Vergleich

```go
// Severity-Ordnung (aus internal/rules)
SeverityInfo < SeverityWarn < SeverityCrit
```

**Beispiele:**
- `SeverityInfo < SeverityWarn` → `true` (Push auslösen)
- `SeverityWarn < SeverityCrit` → `true` (Push auslösen)
- `SeverityCrit < SeverityCrit` → `false` (kein Push)
- `SeverityWarn < SeverityInfo` → `false` (kein Push, Verbesserung)

## Integration mit Runner

Der Counter & Severity-Flow wird vom Runner orchestriert:

```go
// Pseudocode
for {
    result := healthChecker.Check(ctx)
    
    if result.Success {
        counter.Reset()
        newSeverity = rules.SeverityInfo
    } else {
        counter.Increment()
        failCount := counter.GetCount()
        
        if failCount >= critThreshold {
            newSeverity = rules.SeverityCrit
        } else if failCount >= warnThreshold {
            newSeverity = rules.SeverityWarn
        } else {
            newSeverity = rules.SeverityInfo
        }
    }
    
    // Push nur bei Statuswechsel nach oben
    if newSeverity > state.CurrentSeverity {
        pushAdapter.Send(pushMessage)
    }
    
    // GPIO aktualisieren
    gpioHandler.SetLED(newSeverity)
    
    // State aktualisieren
    state.CurrentSeverity = newSeverity
    
    time.Sleep(interval_seconds)
}
```

## Beispiel-Ablauf

### Szenario 1: Erfolgreiche Checks

```
T+0s:  Health Check → Success
       Counter.Reset() → FailCount = 0
       Severity = INFO
       Push: ❌ (INFO, kein Push)
       GPIO: LED Grün

T+30s: Health Check → Success
       Counter.Reset() → FailCount = 0
       Severity = INFO
       Push: ❌ (kein Statuswechsel)
       GPIO: LED Grün
```

### Szenario 2: Fehlversuche (WARN → CRIT)

```
T+0s:  Health Check → Failure
       Counter.Increment() → FailCount = 1
       Severity = INFO
       Push: ❌ (INFO, kein Push)
       GPIO: LED Grün

T+30s: Health Check → Failure
       Counter.Increment() → FailCount = 2
       Severity = INFO
       Push: ❌ (INFO, kein Push)
       GPIO: LED Grün

T+60s: Health Check → Failure
       Counter.Increment() → FailCount = 3
       Severity = WARN (threshold: 3)
       Push: ✔️ (INFO→WARN, Push an prox-watch-warn)
       GPIO: LED Gelb

T+90s: Health Check → Failure
       Counter.Increment() → FailCount = 4
       Severity = WARN
       Push: ❌ (WARN→WARN, kein Statuswechsel)
       GPIO: LED Gelb

T+120s: Health Check → Failure
       Counter.Increment() → FailCount = 10
       Severity = CRIT (threshold: 10)
       Push: ✔️ (WARN→CRIT, Push an prox-watch-crit)
       GPIO: LED Rot, Beeper an (Tag)
```

### Szenario 3: Recovery (CRIT → INFO)

```
T+0s:  Health Check → Failure
       FailCount = 10, Severity = CRIT
       GPIO: LED Rot

T+30s: Health Check → Success
       Counter.Reset() → FailCount = 0
       Severity = INFO
       Push: ❌ (CRIT→INFO, Verbesserung, kein Push)
       GPIO: LED Grün
```

## Nächste Schritte

1. GPIO-Abstraktion
2. Finales Runner-Modell
3. Phase 1 Implementierungsreihenfolge
