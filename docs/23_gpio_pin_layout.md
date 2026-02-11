# GPIO Pin Layout (Raspberry Pi)

## Status

**Phase 1.5 - Spezifikation** 🟡

## Pin-Zuordnung (BCM)

### Übersicht

| Funktion | GPIO Pin (BCM) | Physical Pin | Hinweis |
|----------|----------------|--------------|----------|
| LED Grün | 17 | Pin 11 | Widerstand 220Ω |
| LED Gelb | 27 | Pin 13 | Widerstand 220Ω |
| LED Rot | 22 | Pin 15 | Widerstand 220Ω |
| Beeper | 23 | Pin 16 | Optional Transistor |

### Pin-Mapping (Physical)

```
Raspberry Pi GPIO Header (40-Pin)

    3.3V  [1]  [2]  5V
   GPIO2  [3]  [4]  5V
   GPIO3  [5]  [6]  GND
   GPIO4  [7]  [8]  GPIO14
     GND  [9]  [10] GPIO15
  GPIO17  [11] [12] GPIO18  ← LED Grün
  GPIO27  [13] [14] GND     ← LED Gelb
  GPIO22  [15] [16] GPIO23  ← LED Rot, Beeper
    3.3V  [17] [18] GPIO24
  GPIO10  [19] [20] GND
   GPIO9  [21] [22] GPIO25
  GPIO11  [23] [24] GPIO8
     GND  [25] [26] GPIO7
   ...
```

## Hardware-Setup

### LED-Schaltung

**Komponenten:**
- 3× LED (Grün, Gelb, Rot)
- 3× Widerstand 220Ω
- Jumper-Kabel

**Schaltung:**
```
GPIO Pin → Widerstand (220Ω) → LED → GND
```

**Wichtig:**
- Widerstand ist **erforderlich** (schützt GPIO-Pin)
- LED-Anode (+) → Widerstand → GPIO Pin
- LED-Kathode (-) → GND

### Beeper-Schaltung

**Komponenten:**
- 1× Beeper (5V, aktiv)
- 1× Transistor (optional, für höhere Ströme)
- 1× Widerstand (optional, für Transistor)

**Schaltung (einfach):**
```
GPIO Pin → Beeper → GND
```

**Schaltung (mit Transistor):**
```
GPIO Pin → Widerstand → Transistor (Base)
5V → Beeper → Transistor (Collector)
GND → Transistor (Emitter)
```

**Hinweis:**
- Aktiver Beeper benötigt 5V
- GPIO-Pin liefert 3.3V
- Transistor empfohlen für zuverlässigen Betrieb

## Konfiguration

### YAML-Konfiguration

```yaml
gpio:
  enabled: true
  led_pin_green: 17
  led_pin_yellow: 27
  led_pin_red: 22
  beeper_pin: 23
  beeper_day_only: true
  beeper_duration_seconds: 1
```

### Default-Werte

- `led_pin_green: 17` (BCM)
- `led_pin_yellow: 27` (BCM)
- `led_pin_red: 22` (BCM)
- `beeper_pin: 23` (BCM)
- `beeper_day_only: true`
- `beeper_duration_seconds: 1`

## Sicherheitshinweise

### Hardware-Schutz

- **Widerstände erforderlich** - Schützt GPIO-Pins vor Überlastung
- **Maximaler Strom** - GPIO-Pin max. 16mA (mit Widerstand sicher)
- **Polarität beachten** - LED-Anode/Kathode korrekt anschließen

### Software-Schutz

- **Kein Dauerbeep** - Max. 1 Sekunde
- **Tag-Zeitfenster** - Beeper nur 08:00-22:00
- **Nur bei Eskalation** - Beeper nur bei CRIT-Eskalation

## Alternative Pin-Layouts

### Option 2: Einzelne Pins pro LED

| Funktion | GPIO Pin (BCM) |
|----------|----------------|
| LED Grün | 17 |
| LED Gelb | 27 |
| LED Rot | 22 |
| Beeper | 23 |

**Vorteil:** Einfach, klar getrennt

### Option 3: RGB-LED (später)

| Funktion | GPIO Pin (BCM) |
|----------|----------------|
| RGB LED (Rot) | 17 |
| RGB LED (Grün) | 27 |
| RGB LED (Blau) | 22 |
| Beeper | 23 |

**Hinweis:** Für Phase 1.5: Separate LEDs (einfacher)

## Troubleshooting

### LED leuchtet nicht

- Prüfe Widerstand (220Ω)
- Prüfe Polarisierung (Anode/Kathode)
- Prüfe GPIO-Pin (BCM-Nummerierung)
- Prüfe GND-Verbindung

### Beeper piept nicht

- Prüfe Beeper-Spannung (5V aktiv)
- Prüfe Transistor (falls verwendet)
- Prüfe GPIO-Pin (BCM-Nummerierung)
- Prüfe Tag-Zeitfenster (08:00-22:00)

### Falsche LED leuchtet

- Prüfe Pin-Zuordnung in Konfiguration
- Prüfe BCM vs. Physical Pin-Nummerierung
- Prüfe Severity-Mapping (INFO→Grün, WARN→Gelb, CRIT→Rot)
