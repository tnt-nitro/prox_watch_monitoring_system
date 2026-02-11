# Watcher Deployment-Anleitung (Raspberry Pi)

## Status

**Phase 1 - Dokumentation** ✅

## Übersicht

Diese Anleitung beschreibt die Installation und Konfiguration des **prox-watch-watcher** auf einem Raspberry Pi für Out-of-Band-Monitoring.

## Voraussetzungen

- **Raspberry Pi** (beliebiges Modell)
- **Go 1.21+** (für Build) oder vorgebautes Binary
- **systemd** (optional, für Service-Integration)
- **Netzwerk-Zugriff** zum Proxmox-Host

## Build

### Option 1: Build auf Raspberry Pi

```bash
# Go installieren (falls nicht vorhanden)
sudo apt update
sudo apt install golang-go

# Repository klonen
git clone <repository-url>
cd prox-watch-monitoring-system

# Build
go build -o prox-watch-watcher ./cmd/watcher
```

### Option 2: Cross-Compile (von anderem System)

```bash
# Für Raspberry Pi (ARM64)
GOOS=linux GOARCH=arm64 go build -o prox-watch-watcher ./cmd/watcher

# Für Raspberry Pi (ARM32)
GOOS=linux GOARCH=arm go build -o prox-watch-watcher ./cmd/watcher
```

## Installation

### 1. Binary installieren

```bash
# Binary kopieren
sudo cp prox-watch-watcher /usr/local/bin/
sudo chmod 755 /usr/local/bin/prox-watch-watcher

# Prüfen
prox-watch-watcher version
```

### 2. Konfiguration erstellen

```bash
# Verzeichnis erstellen
sudo mkdir -p /etc/prox-watch-watcher

# Beispiel-Konfiguration kopieren
sudo cp config/watcher.yaml.example /etc/prox-watch-watcher/watcher.yaml

# Konfiguration bearbeiten
sudo nano /etc/prox-watch-watcher/watcher.yaml
```

**Wichtig:** Ersetze `PLACEHOLDER` durch echten Hostname/IP (lokal, nicht im Repo!)

**Beispiel-Konfiguration:**
```yaml
watcher:
  interval_seconds: 30

target:
  mode: "ping+https"
  host: "proxmox.local"  # Lokal ersetzen!
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
  enabled: false  # Phase 1: Default false (NoOp)
  led_pin: 17
  beeper_pin: 27
  beeper_day_only: true

security:
  block_ip_literals: true
  require_manual_powercycle: true
```

### 3. systemd Service (optional, später)

**Service-Datei erstellen:**
```bash
sudo nano /etc/systemd/system/prox-watch-watcher.service
```

**Inhalt:**
```ini
[Unit]
Description=prox-watch Watcher (Out-of-Band Monitoring)
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/prox-watch-watcher run
Restart=on-failure
RestartSec=10
User=root
Group=root

# Security Hardening (optional)
NoNewPrivileges=yes
PrivateTmp=yes
ProtectSystem=strict
ProtectHome=yes

[Install]
WantedBy=multi-user.target
```

**Service aktivieren:**
```bash
sudo systemctl daemon-reload
sudo systemctl enable prox-watch-watcher
sudo systemctl start prox-watch-watcher
```

## Verwendung

### Manueller Start (Foreground)

```bash
prox-watch-watcher run
```

### Status anzeigen

```bash
prox-watch-watcher status
```

### Service-Verwaltung

```bash
# Status prüfen
sudo systemctl status prox-watch-watcher

# Logs anzeigen
sudo journalctl -u prox-watch-watcher -f

# Neustart
sudo systemctl restart prox-watch-watcher

# Stoppen
sudo systemctl stop prox-watch-watcher
```

## GPIO-Konfiguration

### LED & Beeper (Phase 1.5)

**Standard: GPIO ist deaktiviert (NoOp)**

Für Hardware-GPIO:
1. GPIO aktivieren: `gpio.enabled: true`
2. GPIO-Pins konfigurieren: `led_pin`, `beeper_pin`
3. Hardware-GPIO-Implementierung installieren (Build mit `-tags raspberry`)

Siehe [docs/22_gpio_hardware_architecture.md](docs/22_gpio_hardware_architecture.md) für Details.

### Power-Cycle (Phase 3)

**⚠️ KRITISCH: Power-Cycle steuert physische Hardware und kann zu Datenverlust führen.**

**Standard: Power-Cycle ist deaktiviert**

Für Power-Cycle:
1. **Lesen Sie zuerst:** [docs/24_powercycle_safety.md](docs/24_powercycle_safety.md)
2. Hardware-Verdrahtung prüfen (NO/NC-Relais)
3. `powercycle.enabled: true` setzen
4. `powercycle.relay_mode` korrekt konfigurieren (Pflicht!)
5. ARM-Datei erstellen: `sudo touch /var/lib/prox-watch/arm_powercycle`

**Kill-Switch (Schnellste Deaktivierung):**
```yaml
powercycle:
  enabled: false
```

Nach Änderung: `sudo systemctl restart prox-watch-watcher`

## Troubleshooting

### Keine Push-Benachrichtigungen

- Prüfe `push.enabled: true` in Konfiguration
- Prüfe ntfy-Server-Erreichbarkeit
- Prüfe Topics (`prox-watch-warn`, `prox-watch-crit`)
- Prüfe Firewall-Regeln

### Health-Check schlägt fehl

- Prüfe `target.host` und `target.port`
- Prüfe Netzwerk-Erreichbarkeit: `ping <host>`
- Prüfe HTTPS-Zugriff: `curl -k https://<host>:8006`
- Prüfe `target.timeout_seconds` (zu kurz?)

### Counter wird nicht zurückgesetzt

- Normal bei In-Memory State (Phase 1)
- Neustart → Counter wird zurückgesetzt
- Phase 2: Optional SQLite-Persistenz

### Service startet nicht

- Prüfe Logs: `sudo journalctl -u prox-watch-watcher`
- Prüfe Konfiguration: `prox-watch-watcher run` (manuell)
- Prüfe Berechtigungen: Binary muss ausführbar sein

## Sicherheitshinweise

- **Keine IP-Literale im Repository** - Konfiguration lokal speichern
- **Keine Host/IP in Logs** - Abstrakte Fehlermeldungen
- **Kein Auto-Reboot** - Keine automatischen System-Neustarts
- **Kein Power-Cycle** - Keine automatischen Hardware-Aktionen (Phase 1)

## Nächste Schritte

- Phase 2: Hardware-GPIO-Implementierung
- Phase 2: SQLite-Persistenz (optional)
- Phase 2: Cooldown-Mechanismus
- Phase 2: Power-Cycle (mit manueller Freigabe)
