# GPIO Verkabelung - NeoPixel, Beeper, Button

## NeoPixel WS2812 Stick (8× WS2812)

### Pin-Belegung

| NeoPixel Pin | Funktion | Raspberry Pi | Hinweis |
|--------------|----------|--------------|---------|
| 5V | Versorgung | Pin 2 (5V) | ⚠️ **NICHT 3.3V verwenden!** |
| GND | Masse | Pin 6 (GND) | ⚠️ **Muss verbunden sein!** |
| DIN | Daten Eingang | Pin 12 (GPIO18 / PWM0) | Mit 330Ω Widerstand (empfohlen) |
| DOUT | Daten Ausgang | — | Nicht anschließen (für Erweiterung) |

### Verbindung

```
NeoPixel Stick          Raspberry Pi 3B+
─────────────────       ─────────────────
5V  ────────────────→  Pin 2 (5V)
GND ────────────────→  Pin 6 (GND)
DIN ──[330Ω]────────→  Pin 12 (GPIO18)
DOUT ────────────────  (nicht anschließen)
```

### ⚠️ Wichtige Hinweise

- **GND muss verbunden sein** - sonst funktioniert es nicht!
- **Kein 3.3V verwenden** - nur 5V!
- **Keine langen Kabel** - max. 30cm empfohlen
- **330Ω Widerstand** zwischen GPIO18 und DIN (empfohlen)
- **1000µF Elko** zwischen 5V und GND (optional, für Dauerbetrieb)

## Beeper KY-012 (Aktiver Buzzer)

### Pin-Belegung

| KY-012 Pin | Funktion | Raspberry Pi | Hinweis |
|------------|----------|--------------|---------|
| + / S | Signal | Pin 18 (GPIO24) | Aktiver Buzzer |
| GND | Masse | Pin 14 (GND) | |

### Verbindung

```
KY-012 Beeper          Raspberry Pi 3B+
─────────────────       ─────────────────
+ / S ──────────────→  Pin 18 (GPIO24)
GND ────────────────→  Pin 14 (GND)
```

**Hinweis:** KY-012 ist ein **aktiver Buzzer** - benötigt nur HIGH/LOW, kein PWM.

## Button (Taster)

### Pin-Belegung

| Button Pin | Funktion | Raspberry Pi | Hinweis |
|------------|----------|--------------|---------|
| COM | Gemeinsam | Pin 20 (GND) | Oder Pin 14 (GND) |
| NO | Normal Offen | Pin 16 (GPIO23) | Mit internem Pull-Up |

### Verbindung

```
Button                 Raspberry Pi 3B+
─────────────────       ─────────────────
COM ────────────────→  Pin 20 (GND)
NO  ────────────────→  Pin 16 (GPIO23)
```

**Hinweis:** 
- Interner Pull-Up aktiv
- Gedrückt = LOW (Pull-Up wird nach GND gezogen)

## GPIO-Pin-Farbcodierung

Für bessere Übersicht beim Hardware-Setup:

| Farbe | GPIO-Pins |
|-------|-----------|
| Orange | GPIO 2, 3 |
| Rot | GPIO 4, 17 |
| Pink | GPIO 27, 22 |
| Schwarz | GPIO 23 (Button) |
| Grün | GPIO 24 (Beeper) |
| Lila | GPIO 18 (NeoPixel) |
| Dunkelgrau | GND, 5V, 3.3V |
