#!/bin/bash
# Installation Script für prox-watch-watcher auf Raspberry Pi
# 
# Verwendung:
#   curl -fsSL https://raw.githubusercontent.com/tnt-nitro/prox_watch_monitoring_system/main/installer/install_watcher.sh | bash
#   Oder: wget -O- https://raw.githubusercontent.com/tnt-nitro/prox_watch_monitoring_system/main/installer/install_watcher.sh | bash

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
    echo -e "${YELLOW}Warnung: Dies scheint kein Raspberry Pi zu sein. Fortfahren? (j/n)${NC}"
    read -r response
    if [ "$response" != "j" ]; then
        exit 1
    fi
fi

echo -e "${GREEN}=== prox-watch-watcher Installation ===${NC}"

# 0. System aktualisieren (Basissetup)
echo -e "${YELLOW}System wird aktualisiert...${NC}"
apt-get update
apt-get upgrade -y

# 1. Git installieren (falls nicht vorhanden)
if ! command -v git &> /dev/null; then
    echo -e "${YELLOW}Git wird installiert...${NC}"
    apt-get update
    apt-get install -y git
else
    echo -e "${GREEN}Git ist bereits installiert: $(git --version)${NC}"
fi

# 2. Go installieren (falls nicht vorhanden)
if ! command -v go &> /dev/null; then
    echo -e "${YELLOW}Go wird installiert...${NC}"
    apt-get update
    apt-get install -y golang-go
else
    echo -e "${GREEN}Go ist bereits installiert: $(go version)${NC}"
fi

# 3. Repository klonen (oder aktualisieren)
INSTALL_DIR="/tmp/prox-watch-install"
REPO_URL="https://github.com/tnt-nitro/prox_watch_monitoring_system.git"

if [ -d "$INSTALL_DIR" ]; then
    echo -e "${YELLOW}Repository aktualisieren...${NC}"
    cd "$INSTALL_DIR"
    git pull || true
else
    echo -e "${YELLOW}Repository klonen...${NC}"
    git clone "$REPO_URL" "$INSTALL_DIR"
    cd "$INSTALL_DIR"
fi

# 4. Binary bauen
echo -e "${YELLOW}Binary wird gebaut...${NC}"
cd "$INSTALL_DIR"

# Abhängigkeiten auflösen
echo -e "${YELLOW}Abhängigkeiten werden aufgelöst...${NC}"
go mod tidy

# Prüfe, ob Hardware-GPIO aktiviert werden soll
BUILD_TAGS=""
if [ "${ENABLE_GPIO:-}" = "1" ]; then
    BUILD_TAGS="-tags raspberry"
    echo -e "${GREEN}Hardware-GPIO wird aktiviert${NC}"
fi

go build $BUILD_TAGS -o prox-watch-watcher ./cmd/watcher

# 4.1. LED-Test-Tool bauen (wenn GPIO aktiviert)
if [ "${ENABLE_GPIO:-}" = "1" ]; then
    echo -e "${YELLOW}LED-Test-Tool wird gebaut...${NC}"
    # Prüfe, ob cmd/ledtest existiert (nur wenn Branch auf GitHub ist)
    if [ -d "$INSTALL_DIR/cmd/ledtest" ]; then
        go build $BUILD_TAGS -o prox-watch-ledtest ./cmd/ledtest
        if [ -f prox-watch-ledtest ]; then
            install -m 755 prox-watch-ledtest /usr/local/bin/prox-watch-ledtest
            echo -e "${GREEN}LED-Test-Tool installiert${NC}"
        fi
    else
        echo -e "${YELLOW}Hinweis: cmd/ledtest nicht gefunden (Branch feature/gpio-test-tool noch nicht auf GitHub)${NC}"
        echo -e "${YELLOW}LED-Test-Tool wird übersprungen${NC}"
    fi
fi

# 5. Binary installieren
echo -e "${YELLOW}Binary wird installiert...${NC}"
install -m 755 prox-watch-watcher /usr/local/bin/prox-watch-watcher

# 6. User/Group erstellen
if ! id "prox-watch-watcher" &>/dev/null; then
    echo -e "${YELLOW}User/Group wird erstellt...${NC}"
    useradd -r -s /bin/false -d /var/lib/prox-watch-watcher prox-watch-watcher
else
    echo -e "${GREEN}User prox-watch-watcher existiert bereits${NC}"
fi

# 7. Verzeichnisse erstellen
echo -e "${YELLOW}Verzeichnisse werden erstellt...${NC}"
mkdir -p /var/lib/prox-watch-watcher
chown prox-watch-watcher:prox-watch-watcher /var/lib/prox-watch-watcher
chmod 700 /var/lib/prox-watch-watcher

mkdir -p /etc/prox-watch-watcher
chown root:root /etc/prox-watch-watcher
chmod 755 /etc/prox-watch-watcher

# 8. Konfiguration erstellen (falls nicht vorhanden)
if [ ! -f /etc/prox-watch-watcher/watcher.yaml ]; then
    echo -e "${YELLOW}Beispiel-Konfiguration wird erstellt...${NC}"
    cat > /etc/prox-watch-watcher/watcher.yaml << 'EOF'
watcher:
  interval_seconds: 30
  cooldown_seconds: 600

target:
  mode: "ping+https"
  host: "PLACEHOLDER"  # Ersetzen durch echten Hostname/IP! OHNE Portnummer (Port wird separat konfiguriert)
  port: 8006
  timeout_seconds: 5

thresholds:
  warn: 3
  crit: 10

push:
  enabled: true
  adapter: "ntfy"
  topics:
    warn: "prox-watch-warn"
    crit: "prox-watch-crit"

gpio:
  enabled: false
  backend: "noop"
  led_pin_green: 17
  led_pin_yellow: 27
  led_pin_red: 22
  beeper_pin: 23
  beeper_day_only: true
  beeper_window_start: "08:00"
  beeper_window_end: "22:00"
  beeper_max_ms: 1000

powercycle:
  enabled: false
  gpio_pin: 24
  relay_active_high: false
  relay_mode: ""  # MUSS gesetzt werden, wenn enabled=true! Werte: "cut_power_on_active" (NO-Relais) oder "cut_power_on_inactive" (NC-Relais). Siehe docs/24_powercycle_safety.md
  max_attempts: 1
  min_downtime_seconds: 15
  retry_after_seconds: 900
  require_manual_arm: true
  arm_file_path: "/var/lib/prox-watch-watcher/arm_powercycle"

security:
  block_ip_literals: true
EOF
    echo -e "${YELLOW}WICHTIG: Bearbeiten Sie /etc/prox-watch-watcher/watcher.yaml und ersetzen Sie PLACEHOLDER!${NC}"
else
    echo -e "${GREEN}Konfiguration existiert bereits${NC}"
fi

# 9. systemd Service installieren
echo -e "${YELLOW}systemd Service wird installiert...${NC}"
if [ -f "$INSTALL_DIR/installer/watcher.service" ]; then
    cp "$INSTALL_DIR/installer/watcher.service" /etc/systemd/system/prox-watch-watcher.service
    systemctl daemon-reload
    systemctl enable prox-watch-watcher.service
    echo -e "${GREEN}Service installiert und aktiviert${NC}"
else
    echo -e "${RED}Fehler: watcher.service nicht gefunden${NC}"
    exit 1
fi

# 10. GPIO-Zugriff konfigurieren (optional)
if [ "${ENABLE_GPIO:-}" = "1" ]; then
    echo -e "${YELLOW}GPIO-Zugriff wird konfiguriert...${NC}"
    
    # Prüfe, ob gpio-Gruppe existiert
    if ! getent group gpio > /dev/null 2>&1; then
        groupadd -r gpio
    fi
    
    # User zur gpio-Gruppe hinzufügen
    usermod -a -G gpio prox-watch-watcher
    
    # udev-Regel erstellen (falls nicht vorhanden)
    if [ ! -f /etc/udev/rules.d/99-gpio-prox-watch.rules ]; then
        cat > /etc/udev/rules.d/99-gpio-prox-watch.rules << 'EOF'
# GPIO-Zugriff für prox-watch-watcher
KERNEL=="gpiochip*", GROUP="gpio", MODE="0664"
SUBSYSTEM=="gpio", GROUP="gpio", MODE="0664"
EOF
        udevadm control --reload-rules
        udevadm trigger
        echo -e "${GREEN}udev-Regel erstellt${NC}"
    fi
fi

# 11. Zusammenfassung
echo -e "${GREEN}=== Installation abgeschlossen ===${NC}"
echo ""
echo "Nächste Schritte:"
echo "1. Konfiguration bearbeiten:"
echo "   sudo nano /etc/prox-watch-watcher/watcher.yaml"
echo ""
echo "2. Service starten:"
echo "   sudo systemctl start prox-watch-watcher.service"
echo ""
echo "3. Status prüfen:"
echo "   sudo systemctl status prox-watch-watcher.service"
echo ""
echo "4. Logs anzeigen:"
echo "   sudo journalctl -u prox-watch-watcher.service -f"
echo ""

# Aufräumen (optional)
if [ "${CLEANUP:-}" = "1" ]; then
    echo -e "${YELLOW}Aufräumen...${NC}"
    rm -rf "$INSTALL_DIR"
fi

echo -e "${GREEN}Fertig!${NC}"
