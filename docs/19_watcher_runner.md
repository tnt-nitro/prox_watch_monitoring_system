# Finales Runner-Modell (Watcher Phase 1)

## Status

**Phase 1 - Spezifikation** 🟡

## Ziel

Single-Thread Event-Loop für Watcher.

**Eigenschaften:**
- ✅ Single Thread
- ✅ Kein Mutex
- ✅ Kein Persistenzspeicher
- ✅ Kein Retry-Sturm
- ✅ Graceful Shutdown via Context

## Ablauf

```
Start
 └─ Ticker (interval_seconds)
     └─ Health Check
         └─ Counter
             └─ Severity
                 ├─ Push (bei Eskalation)
                 └─ GPIO Update
```

### Detaillierter Ablauf

1. **Start:**
   ```go
   runner := NewRunner(config, healthChecker, counter, pushAdapter, gpio)
   ctx := context.WithCancel(context.Background())
   go runner.Run(ctx)
   ```

2. **Event-Loop (alle `interval_seconds`):**
   ```go
   ticker := time.NewTicker(interval_seconds)
   defer ticker.Stop()
   
   for {
       select {
       case <-ctx.Done():
           return nil
       case <-ticker.C:
           // Health Check
           result := healthChecker.Check(ctx)
           
           // Counter & Severity
           if result.Success {
               counter.Reset()
               newSeverity = rules.SeverityInfo
           } else {
               counter.Increment()
               failCount := counter.GetCount()
               newSeverity = evaluateSeverity(failCount)
           }
           
           // Push (nur bei Statuswechsel nach oben)
           if newSeverity > state.CurrentSeverity {
               pushAdapter.Send(pushMessage)
           }
           
           // GPIO Update
           gpio.SetLED(newSeverity)
           if newSeverity == rules.SeverityCrit && isDayTime() && isStatusChange() {
               gpio.Beep()
           }
           
           // State aktualisieren
           state.CurrentSeverity = newSeverity
       }
   }
   ```

3. **Graceful Shutdown:**
   ```go
   cancel() // Context-Cancellation
   runner.Wait() // Warten auf Loop-Ende
   ```

## Komponenten

### Runner-Struktur

```go
type Runner struct {
    config        *WatcherConfig
    healthChecker HealthChecker
    counter       Counter
    pushAdapter   push.Adapter
    gpio          GPIO
    state         *State
    ticker        *time.Ticker
    ctx           context.Context
    cancel        context.CancelFunc
}
```

### Dependencies

- **HealthChecker** - Health-Check-Engine
- **Counter** - Fehlzähler
- **PushAdapter** - Push-Benachrichtigungen (shared: `internal/push`)
- **GPIO** - LED/Beeper-Steuerung
- **State** - Aktueller Zustand (In-Memory)

## Interface

```go
type Runner interface {
    Run(ctx context.Context) error
    Stop() error
    Wait() error
}
```

**Methoden:**
- `Run(ctx)` - Startet Event-Loop (blockierend)
- `Stop()` - Stoppt Event-Loop (Context-Cancellation)
- `Wait()` - Wartet auf Loop-Ende

## Fehlerbehandlung

### Health-Check-Fehler

- **Timeout:** → `result.Success = false`, Counter erhöhen
- **Netzwerkfehler:** → `result.Success = false`, Counter erhöhen
- **Kein Stopp:** Runner läuft weiter, nächster Check in `interval_seconds`

### Push-Fehler

- **Push-Fehler:** → Log (optional), kein Stopp
- **Kein Retry:** Push wird nicht wiederholt
- **Runner läuft weiter:** Nächster Check wie geplant

### GPIO-Fehler

- **GPIO-Fehler:** → Log (optional), kein Stopp
- **NoOp-Fallback:** Wenn Hardware-GPIO fehlschlägt → NoOp
- **Runner läuft weiter:** Nächster Check wie geplant

## Graceful Shutdown

### Signal-Handling

```go
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

go func() {
    <-sigChan
    runner.Stop()
}()

runner.Run(ctx)
```

### Shutdown-Ablauf

1. **Signal empfangen** (SIGTERM, SIGINT)
2. **Context-Cancellation** (`cancel()`)
3. **Ticker stoppen** (läuft aus)
4. **Aktueller Check beenden** (wenn möglich)
5. **Ressourcen schließen** (GPIO.Close(), etc.)
6. **Runner beendet**

## Integration mit systemd

### Service-Unit (später)

```ini
[Unit]
Description=prox-watch Watcher
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/prox-watch-watcher run
Restart=on-failure
RestartSec=10

[Install]
WantedBy=multi-user.target
```

**Eigenschaften:**
- **Type=simple** - Foreground-Prozess
- **Restart=on-failure** - Automatischer Neustart bei Fehler
- **Graceful Shutdown** - SIGTERM → Context-Cancellation

## Beispiel-Ablauf

### Start

```
T+0s:  Runner.Start()
       Ticker erstellt (30s)
       Health Check startet
```

### Normaler Betrieb

```
T+0s:  Health Check → Success
       Counter.Reset()
       Severity = INFO
       GPIO: LED Grün
       State: CurrentSeverity = INFO

T+30s: Health Check → Success
       Counter.Reset()
       Severity = INFO
       GPIO: LED Grün
       State: CurrentSeverity = INFO (kein Statuswechsel)
```

### Fehler-Szenario

```
T+0s:  Health Check → Failure
       Counter.Increment() → FailCount = 1
       Severity = INFO
       GPIO: LED Grün
       State: CurrentSeverity = INFO

T+30s: Health Check → Failure
       Counter.Increment() → FailCount = 2
       Severity = INFO
       GPIO: LED Grün
       State: CurrentSeverity = INFO

T+60s: Health Check → Failure
       Counter.Increment() → FailCount = 3
       Severity = WARN (threshold: 3)
       Push: ✔️ (INFO→WARN)
       GPIO: LED Gelb
       State: CurrentSeverity = WARN

T+90s: Health Check → Failure
       Counter.Increment() → FailCount = 10
       Severity = CRIT (threshold: 10)
       Push: ✔️ (WARN→CRIT)
       GPIO: LED Rot, Beeper (Tag)
       State: CurrentSeverity = CRIT
```

### Shutdown

```
T+0s:  SIGTERM empfangen
       cancel() → Context-Cancellation
       Ticker stoppt
       Aktueller Check beendet
       GPIO.Close()
       Runner beendet
```

## Nächste Schritte

1. Phase 1 Implementierungsreihenfolge
2. Implementierung starten
