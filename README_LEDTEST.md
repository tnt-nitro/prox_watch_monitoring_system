# LED & Beeper Test Tool - Schnellstart

## Ein-Befehl-Installation

```bash
curl -fsSL https://raw.githubusercontent.com/tnt-nitro/prox_watch_monitoring_system/main/installer/install_ledtest.sh | sudo bash
```

**Fertig!** Das Tool ist installiert und kann mit `sudo prox-watch-ledtest` gestartet werden.

## Was wird installiert?

- ✅ System-Updates (`apt update && apt upgrade`)
- ✅ WS2812-Bibliothek (rpi_ws281x) für NeoPixel
- ✅ Go (falls nicht vorhanden)
- ✅ Repository von GitHub
- ✅ Go-Bindings für WS2812
- ✅ Binary-Kompilierung
- ✅ Installation nach `/usr/local/bin/prox-watch-ledtest`

## Verwendung

```bash
sudo prox-watch-ledtest
```

**Hinweis:** `sudo` ist erforderlich für GPIO-Zugriff.

## Hardware

- **NeoPixel (WS2812):** 5V (Pin 2), GND (Pin 6), DIN (Pin 12/GPIO18)
- **Beeper (KY-012):** + (Pin 18/GPIO24), GND (Pin 14)
- **Button:** NO (Pin 16/GPIO23), COM (GND)

Siehe [docs/30_gpio_wiring.md](docs/30_gpio_wiring.md) für Details.

## Funktionen

- **LED Test:** LED 0 Rot, LED 1 Grün, LED 2 Blau, Alle aus, Lauflicht
- **Beeper Test:** (wird implementiert)
- **GPIO Übersicht:** Dauerhaft sichtbare GPIO-Matrix

## Weitere Informationen

- **Vollständige Installation:** [docs/34_ledtest_github_installation.md](docs/34_ledtest_github_installation.md)
- **Architektur:** [docs/33_gpio_test_tool_overview.md](docs/33_gpio_test_tool_overview.md)
- **Verkabelung:** [docs/30_gpio_wiring.md](docs/30_gpio_wiring.md)
