#!/bin/bash
# Automatisches Setup-Skript für GPIO Test Tool
# 
# Verwendung:
#   curl -fsSL https://raw.githubusercontent.com/tnt-nitro/prox_watch_monitoring_system/main/installer/setup_gpio_test_tool.sh | bash
#   Oder: wget -O- https://raw.githubusercontent.com/tnt-nitro/prox_watch_monitoring_system/main/installer/setup_gpio_test_tool.sh | bash

set -euo pipefail

# Farben für Output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== GPIO Test Tool - Automatisches Setup ===${NC}"

# 1. Prüfe, ob im Projekt-Verzeichnis
if [ ! -f "go.mod" ]; then
    echo -e "${RED}Fehler: go.mod nicht gefunden. Bitte im Projekt-Verzeichnis ausführen.${NC}"
    echo "Beispiel: cd ~/prox_watch_monitoring_system && bash installer/setup_gpio_test_tool.sh"
    exit 1
fi

PROJECT_DIR="$(pwd)"
echo -e "${GREEN}Projekt-Verzeichnis: $PROJECT_DIR${NC}"

# 2. Schritt 1: WS2812-Bibliothek installieren
echo -e "${YELLOW}=== Schritt 1: WS2812-Bibliothek installieren ===${NC}"
if [ -f "installer/install_ws2812.sh" ]; then
    echo -e "${YELLOW}WS2812-Installationsskript wird ausgeführt...${NC}"
    sudo bash installer/install_ws2812.sh
else
    echo -e "${YELLOW}WS2812-Installationsskript nicht gefunden, manuelle Installation...${NC}"
    echo "Bitte manuell ausführen: sudo bash installer/install_ws2812.sh"
    read -p "Fortfahren? (j/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Jj]$ ]]; then
        exit 1
    fi
fi

# 3. Schritt 2: Go-Bindings installieren
echo -e "${YELLOW}=== Schritt 2: Go-Bindings installieren ===${NC}"

# CGO prüfen
if [ "$(go env CGO_ENABLED)" != "1" ]; then
    echo -e "${YELLOW}CGO wird aktiviert...${NC}"
    export CGO_ENABLED=1
fi

# Go Binding installieren
echo -e "${YELLOW}Go Binding wird installiert...${NC}"
go get github.com/rpi-ws281x/rpi-ws281x-go || {
    echo -e "${RED}Fehler beim Installieren des Go-Bindings${NC}"
    exit 1
}

# Abhängigkeiten aktualisieren
echo -e "${YELLOW}Abhängigkeiten werden aktualisiert...${NC}"
go mod tidy

# 4. Schritt 3: Build-Test
echo -e "${YELLOW}=== Schritt 3: Build-Test ===${NC}"

# Prüfe, ob cmd/ledtest existiert
if [ ! -d "cmd/ledtest" ]; then
    echo -e "${RED}Fehler: cmd/ledtest Verzeichnis nicht gefunden${NC}"
    echo -e "${YELLOW}Hinweis: Der Branch feature/gpio-test-tool muss auf GitHub sein${NC}"
    echo "Oder die Dateien müssen manuell erstellt werden."
    exit 1
fi

# Build-Test
echo -e "${YELLOW}LED-Test-Tool wird gebaut...${NC}"
if go build -tags raspberry ./cmd/ledtest; then
    echo -e "${GREEN}✓ Build erfolgreich!${NC}"
    if [ -f "ledtest" ]; then
        echo -e "${GREEN}✓ Binary erstellt: $(pwd)/ledtest${NC}"
    fi
else
    echo -e "${RED}✗ Build fehlgeschlagen${NC}"
    echo "Prüfe:"
    echo "  - Ist libws2811.a vorhanden? (Schritt 1)"
    echo "  - Ist CGO_ENABLED=1?"
    echo "  - Sind alle Build-Tools installiert?"
    exit 1
fi

# 5. Zusammenfassung
echo -e "${GREEN}=== Setup abgeschlossen ===${NC}"
echo ""
echo "Das LED-Test-Tool wurde erfolgreich gebaut."
echo "Ausführen mit:"
echo "  sudo ./ledtest"
echo ""
