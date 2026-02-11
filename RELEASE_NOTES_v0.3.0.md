# Release Notes v0.3.0

**v0.3.0 - Hardware GPIO (Phase 1.5)**

## Status

✅ **Release-fähig** - Hardware GPIO (Phase 1.5) implementiert und getestet

## Features (neu in v0.3.0)

### ✅ Raspberry GPIO (periph.io Integration)

- **Hardware-Backend:** Echte GPIO-Ansteuerung über periph.io für Raspberry Pi
- **Build-Tags:** Trennung zwischen Mock (Default) und Hardware (`-tags raspberry`)
- **Pure Go:** Kein CGO erforderlich, periph.io ist Pure Go
- **BCM-Pin-Mapping:** Unterstützung für BCM-Nummerierung (GPIO17, GPIO27, GPIO22, GPIO23)

### ✅ LED-Statusanzeige

- **Severity-Mapping:**
  - `INFO` → Grüne LED
  - `WARN` → Gelbe LED
  - `CRIT` → Rote LED
- **Sauberer Wechsel:** Nur eine LED ist gleichzeitig aktiv
- **Initial State:** Alle LEDs initial auf LOW

### ✅ Beeper mit Sicherheitslogik

- **Eskalations-Trigger:** Beeper nur bei Eskalation zu CRIT (INFO→CRIT oder WARN→CRIT)
- **Zeitfenster:** Konfigurierbares Tag-Zeitfenster (z.B. 08:00-22:00)
- **Nachtbereich:** Unterstützung für über-Mitternacht-Fenster (z.B. 22:00-06:00)
- **Maximaldauer:** Begrenzt auf `beeper_max_ms` (max. 1000ms)
- **Atomic Concurrency-Schutz:** Verhindert parallele Beep-Aktivierungen
- **Asynchron:** Beeper läuft in Goroutine, blockiert Event-Loop nicht

### ✅ Konfiguration

- **Backend-Auswahl:** `backend: noop` oder `backend: raspberry`
- **Pin-Konfiguration:** Separate Pins für Green/Yellow/Red LED und Beeper
- **Validierung:** Pin-Duplikate und Konflikte werden blockiert
- **Zeitfenster:** `beeper_window_start` und `beeper_window_end` im Format "HH:MM"

### ✅ Tests

- **MockPin-Implementierung:** Vollständige Tests ohne Hardware
- **Hardware-Pfad:** Tests mit Build-Tag `raspberry` (nur auf Raspberry Pi)
- **Integrationstests:** Runner + RaspberryGPIO End-to-End
- **Concurrency-Tests:** Doppelte Beep-Aufrufe, Flag-Reset, Max-Dauer

## Sicherheitsgarantien

- 🔒 **Kein Auto-Reboot:** Automatische System-Neustarts sind nicht implementiert
- 🔒 **Kein Power-Cycle:** Automatische Hardware-Aktionen sind nicht implementiert
- 🔒 **Keine Persistenz:** GPIO-State wird nicht gespeichert
- 🔒 **Kein Fernzugriff:** Keine Remote-Steuerung ohne lokale Freigabe
- 🔒 **Kein Host/IP im Push:** Push-Nachrichten enthalten keine sensiblen Daten
- 🔒 **Kein Beep-Spam:** Atomic Flag verhindert parallele Beep-Aktivierungen
- 🔒 **Max-Dauer:** Beeper maximal 1000ms aktiv
- 🔒 **Zeitfenster:** Beeper nur im konfigurierten Zeitfenster (wenn aktiviert)

## Nicht enthalten (v0.3.0)

- ❌ **Power-Cycle:** Automatische Hardware-Aktionen sind nicht implementiert
- ❌ **Persistenz:** GPIO-State wird nicht über Neustarts hinweg gespeichert
- ❌ **Cooldown:** Kein Cooldown-Mechanismus für Beeper (jeder CRIT-Eskalation löst Beep aus)
- ❌ **Acknowledge:** Keine manuelle Quittierung für Beeper
- ❌ **Mehrere Beeper:** Nur ein Beeper wird unterstützt

## Build-Hinweis

### Standard-Build (Mock-Implementierung)

```bash
go build -o prox-watch-watcher ./cmd/watcher
```

### Hardware-Build (Raspberry Pi)

```bash
# Dependencies installieren
go get periph.io/x/conn/v3/gpio
go get periph.io/x/conn/v3/gpio/gpioreg
go get periph.io/x/host/v3

# Build mit Hardware-Support
go build -tags raspberry -o prox-watch-watcher ./cmd/watcher
```

### Tests

```bash
# Standard-Tests (Mock-Implementierung)
go test ./internal/watcher

# Hardware-Tests (nur auf Raspberry Pi)
go test -tags raspberry ./internal/watcher
```

## Konfiguration

### Beispiel-Konfiguration (watcher.yaml)

```yaml
gpio:
  enabled: true
  backend: "raspberry"  # oder "noop" für Mock
  led_pin_green: 17
  led_pin_yellow: 27
  led_pin_red: 22
  beeper_pin: 23
  beeper_day_only: true
  beeper_window_start: "08:00"
  beeper_window_end: "22:00"
  beeper_max_duration_ms: 1000
```

## Breaking Changes

Keine.

## Bekannte Einschränkungen

1. **Hardware-Build:** Erfordert periph.io Dependencies und Raspberry Pi Hardware
2. **Zeitfenster:** Verwendet lokale Zeit, keine TZ-Handling
3. **Beeper-Dauer:** Maximal 1000ms, keine konfigurierbare Frequenz
4. **Tests:** Zeitfenster-Tests sind ohne Mock-Zeit-Interface nicht vollständig deterministisch

## Nächste Schritte

- **Phase 2 – Persistenz + Cooldown (Watcher):** Einführung eines persistenten States und Cooldown-Mechanismus
- **Phase 3 – Gesicherter Power-Cycle (Watcher):** Design und Implementierung einer kontrollierten Power-Cycle-Funktionalität
- **Stabilisierung:** Erweiterte Tests, Performance-Optimierungen

## Danksagung

Phase 1.5 wurde erfolgreich abgeschlossen. Die Hardware-GPIO-Integration ist stabil, sicher und deterministisch.
