# Health-Check-Engine Design (Phase 1)

## Status

**Phase 1 - Spezifikation** 🟡

## Ziel

Deterministisch prüfen:
- ✅ Ping erreichbar?
- ✅ HTTPS erreichbar?
- ✅ Timeout?
- ✅ Kombinationslogik korrekt?

**Nicht enthalten:**
- ❌ Logging von Host/IP
- ❌ Retry-Logik
- ❌ Aggressives Polling
- ❌ Automatische Anpassung

## Interface

```go
type HealthChecker interface {
    Check(ctx context.Context) (Result, error)
}
```

### Result-Typ

```go
type Result struct {
    Success   bool
    Mode      string              // "ping", "https", "ping+https"
    CheckedAt time.Time
    Latency   int                 // Millisekunden (0 wenn nicht erreichbar)
    Error     string              // Leer wenn Success=true
}
```

**Wichtig:**
- ❌ **Kein Host** im Result
- ❌ **Keine IP** im Result
- ❌ **Keine Log-Inhalte**

## Ablauf pro Intervall

```
Loop (alle interval_seconds)
 ├─ ctx mit Timeout erstellen
 ├─ Ping prüfen (optional, je nach mode)
 ├─ HTTPS prüfen (optional, je nach mode)
 ├─ Ergebnis kombinieren (gemäß mode)
 └─ Result zurückgeben
```

### Detaillierter Ablauf

1. **Context mit Timeout:**
   ```go
   ctx, cancel := context.WithTimeout(context.Background(), timeout_seconds)
   defer cancel()
   ```

2. **Ping-Prüfung (wenn `mode` = `ping` oder `ping+https`):**
   - ICMP-Ping senden
   - Timeout beachten
   - Ergebnis: `pingOK bool`

3. **HTTPS-Prüfung (wenn `mode` = `https` oder `ping+https`):**
   - HTTPS-Request an `host:port`
   - Timeout beachten
   - Ergebnis: `httpsOK bool`

4. **Kombination (gemäß `mode`):**
   - `ping`: `Success = pingOK`
   - `https`: `Success = httpsOK`
   - `ping+https`: `Success = pingOK && httpsOK`

5. **Result erstellen:**
   - `Success`: Kombiniertes Ergebnis
   - `Mode`: Aus Config übernommen
   - `CheckedAt`: `time.Now()`
   - `Latency`: Ping-Latency (falls Ping verwendet) oder 0
   - `Error`: Leer wenn `Success=true`, sonst Fehlermeldung (ohne Host/IP)

## Kombinationslogik

| Mode | Bedingung für SUCCESS |
|------|----------------------|
| `ping` | Ping OK |
| `https` | HTTPS OK |
| `ping+https` | **Beide** OK |

**Fehler = alles andere.**

### Beispiele

**Beispiel 1: `mode: ping`**
- Ping OK → `Success = true`
- Ping fehlgeschlagen → `Success = false`

**Beispiel 2: `mode: https`**
- HTTPS OK → `Success = true`
- HTTPS fehlgeschlagen → `Success = false`

**Beispiel 3: `mode: ping+https`**
- Ping OK **UND** HTTPS OK → `Success = true`
- Ping OK, HTTPS fehlgeschlagen → `Success = false`
- Ping fehlgeschlagen, HTTPS OK → `Success = false`
- Beide fehlgeschlagen → `Success = false`

## Timeout-Regeln

### Timeout-Konfiguration

- **Timeout = `config.target.timeout_seconds`**
- **Default:** 5 Sekunden
- **Min:** 1 Sekunde
- **Max:** 30 Sekunden

### Timeout-Verhalten

1. **Einzelprüfung darf Intervall nicht blockieren:**
   - Wenn Timeout erreicht → `Success = false`, `Error = "timeout"`
   - Kein Warten über Timeout hinaus

2. **Keine Parallelität (Single Thread):**
   - Ping und HTTPS werden **sequentiell** ausgeführt
   - Reihenfolge: Ping zuerst (falls aktiv), dann HTTPS (falls aktiv)
   - Gesamtzeit: `ping_time + https_time <= timeout_seconds`

3. **Context-Cancellation:**
   - Context wird bei Timeout automatisch gecancelt
   - Alle laufenden Checks werden abgebrochen

### Beispiel-Zeitablauf

**Konfiguration:**
- `interval_seconds: 30`
- `timeout_seconds: 5`
- `mode: ping+https`

**Ablauf:**
```
T+0s:  Check startet
T+1s:  Ping OK (1s)
T+2s:  HTTPS OK (1s)
T+2s:  Result: Success=true, Latency=1000ms
T+30s: Nächster Check startet
```

**Timeout-Szenario:**
```
T+0s:  Check startet
T+1s:  Ping OK (1s)
T+5s:  HTTPS Timeout
T+5s:  Result: Success=false, Error="timeout"
T+30s: Nächster Check startet
```

## Sicherheitsgrenzen

### Verboten

- ❌ **Kein Retry innerhalb eines Intervalls:**
  - Jeder Check wird **einmal** ausgeführt
  - Bei Fehler: `Success = false`, warten auf nächstes Intervall

- ❌ **Kein aggressives Polling:**
  - Minimales Intervall: 10 Sekunden
  - Maximales Intervall: 300 Sekunden

- ❌ **Kein Logging sensibler Daten:**
  - Keine Host/IP in Logs
  - Keine Fehlermeldungen mit Host/IP
  - Nur abstrakte Fehlercodes

- ❌ **Keine automatische Anpassung:**
  - Timeout wird nicht dynamisch angepasst
  - Intervall wird nicht dynamisch angepasst
  - Mode wird nicht dynamisch geändert

### Erlaubt

- ✔ **Einmaliger Check pro Intervall**
- ✔ **Timeout-basierte Abbruch**
- ✔ **Abstrakte Fehlermeldungen** (ohne Host/IP)
- ✔ **Deterministisches Verhalten**

## Fehlerbehandlung

### Fehlertypen

1. **Timeout:**
   - `Success = false`
   - `Error = "timeout"`

2. **Netzwerkfehler:**
   - `Success = false`
   - `Error = "network_error"` (ohne Details)

3. **HTTPS-Fehler (z.B. Zertifikat):**
   - `Success = false`
   - `Error = "https_error"` (ohne Details)

4. **Ping-Fehler:**
   - `Success = false`
   - `Error = "ping_error"` (ohne Details)

### Fehler-Propagierung

- Fehler werden **nicht** geloggt (keine Host/IP)
- Fehler werden nur im `Result.Error` gespeichert
- Runner entscheidet über weitere Behandlung (Counter, Push, etc.)

## Integration mit Runner

Der `HealthChecker` wird vom `Runner` aufgerufen:

```go
// Pseudocode
for {
    result := healthChecker.Check(ctx)
    if !result.Success {
        counter.Increment(result.CheckedAt)
        // ... Severity-Evaluierung, Push, etc.
    } else {
        counter.Reset()
    }
    time.Sleep(interval_seconds)
}
```

## Nächste Schritte

1. Counter & Severity-Flow
2. Event-ID-Generierung
3. Push-Integration im Watcher
4. GPIO-Integration
