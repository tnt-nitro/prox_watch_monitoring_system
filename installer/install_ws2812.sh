#!/bin/bash
# Installationsskript für WS2812 (rpi_ws281x) auf Raspberry Pi
# 
# Verwendung:
#   curl -fsSL https://raw.githubusercontent.com/tnt-nitro/prox_watch_monitoring_system/main/installer/install_ws2812.sh | sudo bash
#   Oder: wget -O- https://raw.githubusercontent.com/tnt-nitro/prox_watch_monitoring_system/main/installer/install_ws2812.sh | sudo bash

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

echo -e "${GREEN}=== WS2812 (rpi_ws281x) Installation ===${NC}"

# 1. System aktualisieren
echo -e "${YELLOW}System wird aktualisiert...${NC}"
apt-get update
apt-get upgrade -y

# 2. Build-Tools installieren
echo -e "${YELLOW}Build-Tools werden installiert...${NC}"
apt-get install -y git gcc make build-essential

# 3. rpi_ws281x Repository klonen
echo -e "${YELLOW}rpi_ws281x wird heruntergeladen...${NC}"
WS2812_DIR="/home/pi/rpi_ws281x"

if [ -d "$WS2812_DIR" ]; then
    echo -e "${YELLOW}Verzeichnis existiert bereits, wird aktualisiert...${NC}"
    cd "$WS2812_DIR"
    git pull || true
else
    cd /home/pi
    git clone https://github.com/jgarff/rpi_ws281x.git
    cd "$WS2812_DIR"
fi

# 4. Kompilieren
echo -e "${YELLOW}rpi_ws281x wird kompiliert...${NC}"
if command -v scons &> /dev/null; then
    scons
else
    echo -e "${YELLOW}scons wird installiert...${NC}"
    apt-get install -y scons
    scons
fi

# 5. Prüfe, ob libws2811.a erstellt wurde
if [ -f "$WS2812_DIR/libws2811.a" ]; then
    echo -e "${GREEN}✓ libws2811.a erfolgreich erstellt${NC}"
    echo -e "${GREEN}✓ WS2812-Bibliothek ist bereit${NC}"
else
    echo -e "${RED}Fehler: libws2811.a wurde nicht erstellt${NC}"
    exit 1
fi

# 6. Zusammenfassung
echo -e "${GREEN}=== Installation abgeschlossen ===${NC}"
echo ""
echo "Die WS2812-Bibliothek wurde erfolgreich installiert:"
echo "  - Verzeichnis: $WS2812_DIR"
echo "  - Bibliothek: $WS2812_DIR/libws2811.a"
echo ""
echo "Nächster Schritt: Go-Binding installieren (Schritt 2)"
echo ""
