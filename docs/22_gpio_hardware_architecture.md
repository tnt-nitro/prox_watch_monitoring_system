# Hardware GPIO-Architektur (Phase 1.5)

## Status

**Phase 1.5 - Spezifikation** 🟡

## Ziel

Reale LED + Beeper am Raspberry Pi, ohne Architekturbruch, ohne Risiko für Core.

## Hardware-Abstraktionsschicht

### Architektur

```
internal/watcher/gpio.go
    ├─ NoOpGPIO (bestehend) ✅
    └─ RaspberryGPIO (neu) 🟡
```

**Runner kennt nur:**
```go
type GPIO interface {
    SetLED(sev rules.Severity) error
    Beep() error
    Close() error
}
```

**Keine Änderung am Runner notwendig** - Interface bleibt gleich.

## Raspberry Pin Layout

### Pin-Zuordnung (BCM)

| Funktion | GPIO Pin (BCM) | Hinweis |
|----------|----------------|---------|
| LED Grün | 17 | Widerstand 220Ω |
| LED Gelb | 27 | Widerstand 220Ω |
| LED Rot | 22 | Widerstand 220Ω |
| Beeper | 23 | Optional Transistor |

### Warum BCM?

- **Stabil über Modelle** - BCM-Nummerierung ist konsistent
- **Keine Board-Nummern-Abhängigkeit** - Funktioniert auf allen Raspberry Pi Modellen
- **Standard** - BCM ist der de-facto Standard für GPIO-Bibliotheken

## Sicherheitsmechanismen

### Beeper-Regeln

**Beeper nur bei:**
- ✅ **Eskalation zu CRIT** (INFO→CRIT oder WARN→CRIT)
- ✅ **Zwischen 08:00–22:00** (Tag-Zeitfenster)
- ✅ **Maximal 1 Sekunde** - Kurzer Beep, kein Dauer-Piepen
- ✅ **Kein Loop** - Kein wiederholtes Piepen

**Beeper nicht bei:**
- ❌ INFO oder WARN (nur CRIT)
- ❌ Nacht (22:00 - 08:00)
- ❌ Gleiche Severity (CRIT→CRIT, kein Statuswechsel)
- ❌ Verbesserung (CRIT→INFO, kein Alarm)

### LED-Regeln

| Severity | Verhalten |
|----------|-----------|
| INFO | Grün ON, Gelb OFF, Rot OFF |
| WARN | Gelb ON, Grün OFF, Rot OFF |
| CRIT | Rot ON, Grün OFF, Gelb OFF |
| Wechsel | Alte LED OFF, neue LED ON |

**Wichtig:**
- Nur eine LED gleichzeitig aktiv
- Sauberer Wechsel (kein Flackern)
- Kein Blocking beim LED-Wechsel

## GPIO-Treiber-Architektur

### Optionen

**A) periph.io (empfohlen)**
- ✅ Pure Go
- ✅ Kein CGO
- ✅ Stabil
- ✅ Gut dokumentiert

**B) sysfs (veraltet)**
- ❌ Veraltet, nicht empfohlen

**C) rpi.gpio (CGO → vermeiden)**
- ❌ CGO-Abhängigkeit
- ❌ Nicht empfohlen

### Empfehlung: periph.io

**Struktur:**
```go
internal/watcher/gpio_raspberry.go

type RaspberryGPIO struct {
    greenPin  Pin
    yellowPin Pin
    redPin    Pin
    beeperPin Pin
    isDayTime func() bool
}
```

**Initialisierung:**
```go
func NewRaspberryGPIO(cfg GPIOConfig) (GPIO, error)
```

**Dependencies:**
- `periph.io/x/conn/v3/gpio`
- `periph.io/x/conn/v3/gpio/gpioreg`
- `periph.io/x/host/v3`

## Pin-Abstraktion

### Interface

```go
type Pin interface {
    High() error
    Low() error
    Close() error
}
```

**Vorteile:**
- Testbar ohne Hardware
- Austauschbar (periph.io, sysfs, etc.)
- Mock-fähig

## Teststrategie (ohne Hardware)

### Interface-Mocks

**MockPin:**
```go
type MockPin struct {
    state bool
    mu    sync.Mutex
}

func (m *MockPin) High() error { ... }
func (m *MockPin) Low() error { ... }
```

**Tests prüfen:**
- ✅ Richtige Pins gesetzt (High/Low)
- ✅ Keine Race Conditions
- ✅ Kein Dauerbeep (max. 1 Sekunde)
- ✅ Tag-Zeitfenster korrekt
- ✅ LED-Wechsel sauber

**Keine echte Hardware nötig** - Alle Tests mit Mocks.

## Implementierungsreihenfolge Phase 1.5

1. **GPIOConfig erweitern** - Pin-Konfiguration
2. **Pin-Abstraktion einführen** - Interface für Pins
3. **RaspberryGPIO implementieren** - periph.io-Integration
4. **Beeper-Zeitfenster implementieren** - Tag/Nacht-Logik
5. **Tests mit MockPins** - Vollständige Testabdeckung
6. **Dokumentation ergänzen** - Hardware-Setup, Pin-Layout

## Abbruchkriterien

- ❌ **CGO-Zwang** - Keine CGO-Abhängigkeiten
- ❌ **Root-Zwang ohne Notwendigkeit** - GPIO sollte ohne Root funktionieren (periph.io unterstützt das)
- ❌ **Blocking Sleep > 1s** - Beep max. 1 Sekunde, kein Blocking
- ❌ **Mehrere Goroutines für Beep** - Single-Thread, kein Parallelismus
- ❌ **Logging sensibler Daten** - Keine Host/IP in Logs

## Nächste Schritte

1. GPIOConfig erweitern
2. Pin-Abstraktion einführen
3. RaspberryGPIO implementieren
