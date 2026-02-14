# GPIO Test Tool - Übersicht & Architektur

## Status

**Phase 1 - Implementierung** ✅

## Ziel

Ein separates Test-Tool für GPIO-Hardware (NeoPixel, Beeper, Button), das unabhängig vom Hauptprogramm funktioniert und es Benutzern ermöglicht, ihre Hardware zu testen.

## Architektur

### Projektstruktur

```
cmd/
  ledtest/
    main.go              # Hauptprogramm für LED-Test-Tool

internal/
  gpio/
    gpio.go             # GPIO-Pin-Konstanten
    pin.go              # Pin-Interface
    pin_raspberry.go    # Hardware-Implementierung (Build-Tag: raspberry)
    pin_mock.go         # Mock-Implementierung (Entwicklung)
    beeper.go           # Beeper-Steuerung (KY-012)
    button.go           # Button-Lesung
    init.go             # periph.io Initialisierung
    init_mock.go        # Mock-Initialisierung

  neopixel/
    strip.go            # NeoPixel-Strip (WS2812) - Placeholder

  ui/
    menu.go             # Terminal-UI-Menü (raspi-config Style)
    gpio_matrix.go      # GPIO-Pin-Matrix (dauerhaft sichtbar)
    gpio_overview.go    # GPIO-Übersicht (tabellarisch)

installer/
  install_ws2812.sh     # WS2812-Bibliothek Installation
  create_all_ledtest_files.sh  # Dateien-Erstellung (temporär)
```

### Hardware-Layer vs. UI-Layer

**Trennung:**
- **Hardware Layer** (`internal/gpio/`, `internal/neopixel/`) - Hardware-Abstraktion
- **UI Layer** (`internal/ui/`) - Terminal-Interface
- **Keine Monitoring-Abhängigkeit** - Vollständig unabhängig

## GPIO-Pin-Zuordnung (BCM)

| Funktion | GPIO Pin (BCM) | Physical Pin | Farbe | Hinweis |
|----------|----------------|--------------|-------|---------|
| NeoPixel DIN | 18 | Pin 12 | Lila | PWM0, DMA |
| Beeper + | 24 | Pin 18 | Grün | Aktiver Buzzer |
| Button NO | 23 | Pin 16 | Schwarz | Pull-Up, gedrückt = LOW |

### Pin-Farbcodierung (Matrix Voice Style)

| Farbe | GPIO-Pins | Verwendung |
|-------|----------|-------------|
| Orange | 1, 17 | 3.3V |
| Rot | 2, 4 | 5V |
| Schwarz | 6, 9, 14, 20, etc. | GND |
| Pink | 3, 5 | I2C |
| Grün | 7, 11, 12, 13, etc. | GPIO |
| Blau | 8, 10, etc. | UART/SPI |
| Gelb | 27, 28 | ID EEPROM |

## Hardware-Verkabelung

### NeoPixel WS2812 Stick (8× WS2812)

```
NeoPixel Stick          Raspberry Pi 3B+
─────────────────       ─────────────────
5V  ────────────────→  Pin 2 (5V)
GND ────────────────→  Pin 6 (GND)
DIN ──[330Ω]────────→  Pin 12 (GPIO18)
DOUT ────────────────  (nicht anschließen)
```

**Wichtig:**
- ⚠️ **GND muss verbunden sein**
- ⚠️ **Nur 5V verwenden** (nicht 3.3V)
- ⚠️ **330Ω Widerstand** zwischen GPIO18 und DIN (empfohlen)
- ⚠️ **1000µF Elko** zwischen 5V und GND (optional, für Dauerbetrieb)

### Beeper KY-012 (Aktiver Buzzer)

```
KY-012 Beeper          Raspberry Pi 3B+
─────────────────       ─────────────────
+ / S ──────────────→  Pin 18 (GPIO24)
GND ────────────────→  Pin 14 (GND)
```

**Hinweis:** KY-012 ist ein **aktiver Buzzer** - benötigt nur HIGH/LOW, kein PWM.

### Button (Taster)

```
Button                 Raspberry Pi 3B+
─────────────────       ─────────────────
COM ────────────────→  Pin 20 (GND)
NO  ────────────────→  Pin 16 (GPIO23)
```

**Hinweis:**
- Interner Pull-Up aktiv
- Gedrückt = LOW (Pull-Up wird nach GND gezogen)

## Terminal-UI

### Layout (Split-Screen)

```
┌─────────────────────┬─────────────────────────────────────────────┐
│ LED & BEEPER        │ GPIO PIN MATRIX    NeoPixel  Beeper  Button │
│ TEST TOOL           │                                             │
│                     │ 3.3V  1 │  2 5V ───[NeoPixel 5V]           │
│ > LED Test          │ GPIO3  5 │  6 GND ──[NeoPixel GND]         │
│   Beeper Test       │ GPIO17 11 │ 12 GPIO18──[NeoPixel DIN]       │
│   GPIO Übersicht    │ GPIO27 13 │ 14 GND ───────────[Beeper GND] │
│   Exit              │ GPIO22 15 │ 16 GPIO23────────────[Button NO]│
│                     │ 3.3V 17 │ 18 GPIO24──────────[Beeper +]    │
│                     │ GPIO10 19 │ 20 GND ─────────────[Button GND]│
│                     │ ... (alle 40 Pins)                          │
└─────────────────────┴─────────────────────────────────────────────┘
```

### Features

1. **Dauerhaft sichtbare GPIO-Matrix:**
   - Rechts im Split-Screen
   - Alle 40 Pins mit Farbcodierung
   - Verwendete Pins mit Kabeln markiert

2. **Kabel-Darstellung:**
   - Kabel in Pin-Farbe (Vordergrund)
   - GND-Kabel in dunklem Grau (sichtbar auf schwarzem Hintergrund)
   - Format: `PinBeschreibung-------[Verwendung]`

3. **Menü-Navigation:**
   - ↑ ↓ = Auswahl
   - Enter = öffnen
   - ESC = zurück
   - q = beenden

## Build & Installation

### Voraussetzungen

1. **WS2812-Bibliothek installieren:**
   ```bash
   curl -fsSL https://raw.githubusercontent.com/tnt-nitro/prox_watch_monitoring_system/main/installer/install_ws2812.sh | sudo bash
   ```

2. **Go-Bindings installieren:**
   ```bash
   cd ~/prox_watch_monitoring_system
   go get github.com/rpi-ws281x/rpi-ws281x-go
   go mod tidy
   ```

3. **CGO aktivieren:**
   ```bash
   export CGO_ENABLED=1
   ```

### Build

```bash
cd ~/prox_watch_monitoring_system
go build -tags raspberry ./cmd/ledtest
```

### Ausführung

```bash
sudo ./ledtest
```

**Hinweis:** `sudo` ist erforderlich für GPIO-Zugriff.

## Menü-Funktionen (Phase 1)

### LED Test
- **Status:** Placeholder
- **Geplant:**
  - LED 0 → Rot
  - LED 1 → Grün
  - LED 2 → Blau
  - Alle aus
  - Lauflicht

### Beeper Test
- **Status:** Placeholder
- **Geplant:**
  - Einzelton (500ms)
  - 3× kurz
  - SOS (· · · – – – · · ·)

### GPIO Übersicht
- **Status:** Implementiert
- **Funktion:**
  - Tabellarische GPIO-Übersicht
  - Farbcodierte Pins
  - Verwendungs-Info

## Technische Details

### Build-Tags

- **`raspberry`** - Hardware-Implementierung (periph.io)
- **Default** - Mock-Implementierung (Entwicklung ohne Hardware)

### Abhängigkeiten

- `github.com/rpi-ws281x/rpi-ws281x-go` - WS2812-Bibliothek
- `golang.org/x/term` - Terminal-UI
- `periph.io/x/conn/v3/gpio` - GPIO-Zugriff (nur mit `raspberry` Tag)
- `periph.io/x/host/v3` - periph.io Initialisierung

### GPIO-Initialisierung

1. **periph.io Host initialisieren** (thread-safe, nur einmal)
2. **Pins über BCM-ID laden** (z.B. "GPIO18")
3. **Konfigurieren** (Output/Input mit Pull-Up/Pull-Down)
4. **Initial auf LOW setzen** (Sicherheit)

## Nächste Schritte (Phase 2)

1. **WS2812 NeoPixel-Integration:**
   - DMA-Konfiguration
   - LED-Test-Funktionen implementieren
   - Lauflicht-Modus

2. **Beeper-Test implementieren:**
   - Einzelton
   - 3× kurz
   - SOS-Morse-Code

3. **Button-Integration:**
   - Button-Status anzeigen
   - Interaktive Tests

4. **Installationsskript erweitern:**
   - Automatische WS2812-Installation
   - Automatischer Build
   - Alles in einem Schritt

## Verwendung

### Als separater Befehl

```bash
# Über watcher-Befehl
prox-watch-watcher ledtest

# Oder direkt
sudo prox-watch-ledtest
```

### Integration in Installationsskript

Das Tool wird automatisch gebaut, wenn `ENABLE_GPIO=1` gesetzt ist:

```bash
ENABLE_GPIO=1 curl -fsSL https://raw.githubusercontent.com/tnt-nitro/prox_watch_monitoring_system/main/installer/install_watcher.sh | sudo bash
```

## Dokumentation

- [Verkabelung](docs/30_gpio_wiring.md) - Hardware-Verkabelung
- [GPIO-Matrix](docs/31_gpio_matrix.md) - Pin-Layout
- [Setup-Anleitung](docs/25_gpio_test_tool_setup.md) - Schritt-für-Schritt Setup
