# GPIO-Abstraktion (Watcher Phase 1)

## Status

**Phase 1 - Spezifikation** 🟡

## Ziel

- ✅ LED Status anzeigen
- ✅ Beeper bei CRIT
- ✅ Keine Hardwareabhängigkeit im Core
- ✅ Raspberry optional

**Nicht enthalten:**
- ❌ Hardware-spezifische Implementierung (Phase 1: NoOp)
- ❌ Dauerpiepen
- ❌ Blocking-Operationen

## Interface

```go
type GPIO interface {
    SetLED(sev rules.Severity) error
    Beep() error
    Close() error
}
```

**Methoden:**
- `SetLED(sev)` - Setzt LED-Farbe basierend auf Severity
- `Beep()` - Aktiviert Beeper (nur bei CRIT, Tag)
- `Close()` - Schließt GPIO-Ressourcen

## Verhalten

| Severity | LED | Beeper |
|----------|-----|--------|
| INFO | Grün | ❌ |
| WARN | Gelb | ❌ |
| CRIT | Rot | ✔ (Tag only) |

### LED-Mapping

- **INFO** → LED Grün
- **WARN** → LED Gelb
- **CRIT** → LED Rot

### Beeper-Regeln

**Beeper nur bei:**
- ✅ **CRIT-Severity**
- ✅ **Tageszeit** (Tag = 06:00 - 22:00, konfigurierbar)
- ✅ **Statuswechsel** (INFO→CRIT oder WARN→CRIT)

**Beeper nicht bei:**
- ❌ INFO oder WARN
- ❌ Nacht (22:00 - 06:00)
- ❌ Gleiche Severity (CRIT→CRIT, kein Statuswechsel)

## Sicherheitsregeln

### Verboten

- ❌ **Dauerpiepen:**
  - Beeper wird nur kurz aktiviert (z.B. 1 Sekunde)
  - Kein kontinuierliches Piepen

- ❌ **Blocking:**
  - GPIO-Operationen blockieren nicht den Event-Loop
  - Timeout für GPIO-Operationen (optional)

### Erlaubt

- ✔ **GPIO optional** (wenn `gpio.enabled: false` → NoOp)
- ✔ **Beeper nur bei CRIT + Tag + Statuswechsel**
- ✔ **LED-Update bei jeder Severity-Änderung**

## NoOp-Implementierung (MVP)

```go
type NoOpGPIO struct{}

func (n *NoOpGPIO) SetLED(sev rules.Severity) error {
    // No-op: Immer erfolgreich
    return nil
}

func (n *NoOpGPIO) Beep() error {
    // No-op: Immer erfolgreich
    return nil
}

func (n *NoOpGPIO) Close() error {
    // No-op: Immer erfolgreich
    return nil
}
```

**Eigenschaften:**
- **Immer erfolgreich** (keine Fehler)
- **Keine Hardwareanforderung** (funktioniert überall)
- **MVP-Standard** (Phase 1: NoOp als Default)

## GPIO-Integration im Runner

```go
// Pseudocode
if config.GPIO.Enabled {
    gpio := NewRaspberryGPIO(config.GPIO.LEDPin, config.GPIO.BeeperPin)
} else {
    gpio := NewNoOpGPIO()
}

// Im Event-Loop
gpio.SetLED(newSeverity)

if newSeverity == rules.SeverityCrit && isDayTime() && isStatusChange() {
    gpio.Beep()
}
```

## Hardware-Implementierung (später)

**Phase 1: Nur NoOp**

Hardware-Implementierung (z.B. Raspberry Pi GPIO) ist **nicht** Teil von Phase 1.

**Später (optional):**
- Raspberry Pi GPIO (periph.io/periph)
- LED-Steuerung (GPIO-Pin)
- Beeper-Steuerung (GPIO-Pin)
- Tageszeit-Erkennung

## Konfiguration

```yaml
gpio:
  enabled: false              # Phase 1: Default false (NoOp)
  led_pin: 17                 # BCM-Pin (nur wenn enabled)
  beeper_pin: 27              # BCM-Pin (nur wenn enabled)
  beeper_day_only: true       # Beeper nur tagsüber
```

**Validierung:**
- Wenn `enabled: false` → NoOp-GPIO verwenden
- Wenn `enabled: true` → Hardware-GPIO (später, nicht Phase 1)

## Nächste Schritte

1. Runner-Implementierung
2. Phase 1 Implementierungsreihenfolge
