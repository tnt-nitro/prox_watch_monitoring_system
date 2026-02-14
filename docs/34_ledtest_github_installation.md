# LED & Beeper Test Tool - GitHub Installation

## Einfache Installation (Ein-Befehl)

```bash
curl -fsSL https://raw.githubusercontent.com/tnt-nitro/prox_watch_monitoring_system/main/installer/install_ledtest.sh | sudo bash
```

**Das war's!** Das Skript installiert automatisch:
- ✅ System-Updates (`apt update && apt upgrade`)
- ✅ WS2812-Bibliothek (rpi_ws281x)
- ✅ Go (falls nicht vorhanden)
- ✅ Repository von GitHub
- ✅ Go-Bindings
- ✅ Binary-Kompilierung
- ✅ Installation nach `/usr/local/bin/prox-watch-ledtest`

## Verwendung

Nach der Installation:

```bash
sudo prox-watch-ledtest
```

**Hinweis:** `sudo` ist erforderlich für GPIO-Zugriff.

## Manuelle Installation (Schritt für Schritt)

Falls du die Installation manuell durchführen möchtest:

### Schritt 1: System aktualisieren

```bash
sudo apt update
sudo apt upgrade -y
```

### Schritt 2: WS2812-Bibliothek installieren

```bash
curl -fsSL https://raw.githubusercontent.com/tnt-nitro/prox_watch_monitoring_system/main/installer/install_ws2812.sh | sudo bash
```

### Schritt 3: Repository klonen

```bash
cd ~
git clone https://github.com/tnt-nitro/prox_watch_monitoring_system.git
cd prox_watch_monitoring_system
```

### Schritt 4: Go-Bindings installieren

```bash
export CGO_ENABLED=1
go get github.com/rpi-ws281x/rpi-ws281x-go
go mod tidy
```

### Schritt 5: Binary bauen

```bash
go build -tags raspberry -o prox-watch-ledtest ./cmd/ledtest
```

### Schritt 6: Installieren

```bash
sudo install -m 755 prox-watch-ledtest /usr/local/bin/prox-watch-ledtest
```

## Hardware-Verkabelung

Siehe [docs/30_gpio_wiring.md](docs/30_gpio_wiring.md) für die vollständige Verkabelungsanleitung.

**Kurzfassung:**
- **NeoPixel:** 5V (Pin 2), GND (Pin 6), DIN (Pin 12/GPIO18)
- **Beeper:** + (Pin 18/GPIO24), GND (Pin 14)
- **Button:** NO (Pin 16/GPIO23), COM (GND)

## Fehlerbehebung

### Binary nicht gefunden

```bash
# Prüfe, ob Binary erstellt wurde
ls -lh prox-watch-ledtest

# Falls nicht, prüfe Build-Fehler
go build -v -tags raspberry ./cmd/ledtest
```

### LED leuchtet nicht

1. **Hardware-Verkabelung prüfen** (siehe oben)
2. **Mit sudo ausführen:** `sudo prox-watch-ledtest`
3. **Kernel-Logs prüfen:** `sudo dmesg | tail -20`
4. **WS2812-Bibliothek prüfen:** `ls -lh ~/rpi_ws281x/libws2811.a`

### Build-Fehler

```bash
# Go-Cache löschen
go clean -cache

# Dependencies aktualisieren
go mod tidy

# Neu kompilieren
go build -tags raspberry ./cmd/ledtest
```

## Weitere Informationen

- **Architektur:** [docs/33_gpio_test_tool_overview.md](docs/33_gpio_test_tool_overview.md)
- **Verkabelung:** [docs/30_gpio_wiring.md](docs/30_gpio_wiring.md)
- **GPIO-Matrix:** [docs/31_gpio_matrix.md](docs/31_gpio_matrix.md)
