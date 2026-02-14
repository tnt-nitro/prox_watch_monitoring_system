#!/bin/bash
# Vollständige Installation für LED & Beeper Test Tool
# 
# Verwendung:
#   curl -fsSL https://raw.githubusercontent.com/tnt-nitro/prox_watch_monitoring_system/main/installer/install_ledtest.sh | sudo bash
#   Oder: wget -O- https://raw.githubusercontent.com/tnt-nitro/prox_watch_monitoring_system/main/installer/install_ledtest.sh | sudo bash

set -euo pipefail

# Farben für Output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Prüfe, ob als root ausgeführt
if [ "$EUID" -ne 0 ]; then 
    echo -e "${RED}Fehler: Dieses Skript muss als root ausgeführt werden (sudo)${NC}"
    exit 1
fi

# Prüfe, ob auf Raspberry Pi
if [ ! -f /proc/device-tree/model ] || ! grep -q "Raspberry Pi" /proc/device-tree/model 2>/dev/null; then
    echo -e "${RED}Fehler: Dieses Skript ist nur für Raspberry Pi gedacht${NC}"
    exit 1
fi

echo -e "${GREEN}=== LED & Beeper Test Tool - Vollständige Installation ===${NC}"

# 0. System aktualisieren (Basissetup)
echo -e "${YELLOW}System wird aktualisiert...${NC}"
apt-get update
apt-get upgrade -y

# 1. WS2812-Bibliothek installieren
echo -e "${YELLOW}=== Schritt 1: WS2812-Bibliothek installieren ===${NC}"

# Build-Tools installieren
echo -e "${YELLOW}Build-Tools werden installiert...${NC}"
apt-get install -y git gcc make build-essential scons

# rpi_ws281x Repository klonen oder aktualisieren
WS2812_DIR="/home/pi/rpi_ws281x"

if [ -d "$WS2812_DIR" ]; then
    echo -e "${YELLOW}rpi_ws281x existiert bereits, wird aktualisiert...${NC}"
    cd "$WS2812_DIR"
    git pull || true
else
    cd /home/pi
    git clone https://github.com/jgarff/rpi_ws281x.git
    cd "$WS2812_DIR"
fi

# Kompilieren
echo -e "${YELLOW}rpi_ws281x wird kompiliert...${NC}"
scons

# Prüfe, ob libws2811.a erstellt wurde
if [ -f "$WS2812_DIR/libws2811.a" ]; then
    echo -e "${GREEN}✓ libws2811.a erfolgreich erstellt${NC}"
else
    echo -e "${RED}Fehler: libws2811.a wurde nicht erstellt${NC}"
    exit 1
fi

# 2. Go installieren (falls nicht vorhanden)
echo -e "${YELLOW}=== Schritt 2: Go installieren ===${NC}"
if ! command -v go &> /dev/null; then
    echo -e "${YELLOW}Go wird installiert...${NC}"
    apt-get install -y golang-go
else
    echo -e "${GREEN}Go ist bereits installiert: $(go version)${NC}"
fi

# 3. Repository klonen
echo -e "${YELLOW}=== Schritt 3: Repository klonen ===${NC}"
REPO_DIR="/tmp/prox-watch-install"
REPO_URL="https://github.com/tnt-nitro/prox_watch_monitoring_system.git"
BRANCH="feature/gpio-test-tool"

# Altes Verzeichnis löschen (falls vorhanden und root gehört)
if [ -d "$REPO_DIR" ]; then
    echo -e "${YELLOW}Altes Repository-Verzeichnis wird entfernt...${NC}"
    rm -rf "$REPO_DIR"
fi

# Repository klonen
echo -e "${YELLOW}Repository klonen...${NC}"
git clone -b "$BRANCH" "$REPO_URL" "$REPO_DIR" 2>&1 || {
    echo -e "${YELLOW}Branch $BRANCH nicht gefunden, klone main und wechsle Branch...${NC}"
    git clone "$REPO_URL" "$REPO_DIR"
    cd "$REPO_DIR"
    git checkout "$BRANCH" 2>/dev/null || {
        echo -e "${RED}Fehler: Branch $BRANCH konnte nicht ausgecheckt werden${NC}"
        exit 1
    }
}

# Permissions korrigieren (falls als root erstellt)
chown -R pi:pi "$REPO_DIR" 2>/dev/null || true

# 4. Go-Bindings installieren
echo -e "${YELLOW}=== Schritt 4: Go-Bindings installieren ===${NC}"

# CGO aktivieren
export CGO_ENABLED=1

# Go-Bindings installieren
echo -e "${YELLOW}Go-Bindings werden installiert...${NC}"
cd "$REPO_DIR"
go get github.com/rpi-ws281x/rpi-ws281x-go
go mod tidy

# 5. Binary bauen
echo -e "${YELLOW}=== Schritt 5: Binary bauen ===${NC}"

# Prüfe, ob cmd/ledtest existiert
if [ ! -d "$REPO_DIR/cmd/ledtest" ]; then
    echo -e "${RED}Fehler: cmd/ledtest Verzeichnis nicht gefunden${NC}"
    echo -e "${YELLOW}Der Branch 'feature/gpio-test-tool' ist möglicherweise noch nicht auf GitHub gepusht.${NC}"
    echo -e "${YELLOW}Bitte pushe den Branch zu GitHub oder erstelle die Dateien manuell.${NC}"
    echo ""
    echo "Alternative: Dateien manuell erstellen:"
    echo "  cd $REPO_DIR"
    echo "  mkdir -p cmd/ledtest internal/gpio internal/neopixel internal/ui"
    echo "  # Dann die Dateien per SCP übertragen oder manuell erstellen"
    exit 1
fi

# Als pi-User bauen (falls als root)
if [ "$SUDO_USER" != "" ]; then
    sudo -u "$SUDO_USER" -E go build -tags raspberry -o prox-watch-ledtest ./cmd/ledtest
else
    go build -tags raspberry -o prox-watch-ledtest ./cmd/ledtest
fi

if [ ! -f "$REPO_DIR/prox-watch-ledtest" ]; then
    echo -e "${RED}Fehler: Binary wurde nicht erstellt${NC}"
    echo -e "${YELLOW}Build-Fehler prüfen:${NC}"
    if [ "$SUDO_USER" != "" ]; then
        sudo -u "$SUDO_USER" -E go build -v -tags raspberry ./cmd/ledtest 2>&1 | tail -20
    else
        go build -v -tags raspberry ./cmd/ledtest 2>&1 | tail -20
    fi
    exit 1
fi

# 6. Binary installieren
echo -e "${YELLOW}=== Schritt 6: Binary installieren ===${NC}"
install -m 755 "$REPO_DIR/prox-watch-ledtest" /usr/local/bin/prox-watch-ledtest

# 7. Zusammenfassung
echo -e "${GREEN}=== Installation abgeschlossen ===${NC}"
echo ""
echo "Das LED & Beeper Test Tool wurde erfolgreich installiert!"
echo ""
echo "Verwendung:"
echo "  sudo prox-watch-ledtest"
echo ""
echo "Hinweis: sudo ist erforderlich für GPIO-Zugriff."
echo ""
