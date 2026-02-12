# Watcher systemd Service Installation

## Übersicht

Dieses Dokument beschreibt die Installation und Konfiguration des systemd Service für `prox-watch-watcher`.

## Voraussetzungen

- **systemd** (Standard auf modernen Linux-Systemen)
- **Binary:** `prox-watch-watcher` muss unter `/usr/local/bin/prox-watch-watcher` installiert sein
- **Konfiguration:** Watcher-Konfiguration muss unter `/etc/prox-watch-watcher/watcher.yaml` existieren

## Installation

### 1. User/Group erstellen

```bash
# Dedicated User für Watcher erstellen
sudo useradd -r -s /bin/false -d /var/lib/prox-watch-watcher prox-watch-watcher

# Verzeichnis für State-DB und ARM-Datei erstellen
sudo mkdir -p /var/lib/prox-watch-watcher
sudo chown prox-watch-watcher:prox-watch-watcher /var/lib/prox-watch-watcher
sudo chmod 700 /var/lib/prox-watch-watcher
```

### 2. Konfigurationsverzeichnis erstellen

```bash
# Konfigurationsverzeichnis erstellen
sudo mkdir -p /etc/prox-watch-watcher
sudo chown root:root /etc/prox-watch-watcher
sudo chmod 755 /etc/prox-watch-watcher

# Konfiguration erstellen (siehe cmd/watcher/README.md)
sudo nano /etc/prox-watch-watcher/watcher.yaml
```

### 3. Service-Unit installieren

```bash
# Service-Unit kopieren
sudo cp installer/watcher.service /etc/systemd/system/prox-watch-watcher.service

# systemd neu laden
sudo systemctl daemon-reload

# Service aktivieren (Start beim Boot)
sudo systemctl enable prox-watch-watcher.service

# Service starten
sudo systemctl start prox-watch-watcher.service
```

### 4. Status prüfen

```bash
# Service-Status
sudo systemctl status prox-watch-watcher.service

# Logs anzeigen
sudo journalctl -u prox-watch-watcher.service -f

# State-DB prüfen
sudo -u prox-watch-watcher ls -la /var/lib/prox-watch-watcher/
```

## GPIO-Zugriff (falls Hardware-GPIO aktiviert)

### Option A: udev-Regel (empfohlen)

Erstellen Sie eine udev-Regel für GPIO-Zugriff:

```bash
# udev-Regel erstellen
sudo nano /etc/udev/rules.d/99-gpio-prox-watch.rules
```

Inhalt:
```
# GPIO-Zugriff für prox-watch-watcher
KERNEL=="gpiochip*", GROUP="gpio", MODE="0664"
SUBSYSTEM=="gpio", GROUP="gpio", MODE="0664"
```

```bash
# gpio-Gruppe erstellen (falls nicht vorhanden)
sudo groupadd -r gpio

# User zur gpio-Gruppe hinzufügen
sudo usermod -a -G gpio prox-watch-watcher

# udev-Regeln neu laden
sudo udevadm control --reload-rules
sudo udevadm trigger

# Service neu starten
sudo systemctl restart prox-watch-watcher.service
```

### Option B: gpio-Gruppe (falls vorhanden)

```bash
# User zur gpio-Gruppe hinzufügen
sudo usermod -a -G gpio prox-watch-watcher

# Service-Unit anpassen (SupplementaryGroups=gpio hinzufügen)
sudo nano /etc/systemd/system/prox-watch-watcher.service

# Service neu starten
sudo systemctl daemon-reload
sudo systemctl restart prox-watch-watcher.service
```

### Option C: Root (nicht empfohlen)

**⚠️ WARNUNG:** Nur für Tests, nicht für Produktion!

```bash
# Service-Unit anpassen (User=root)
sudo nano /etc/systemd/system/prox-watch-watcher.service

# Service neu starten
sudo systemctl daemon-reload
sudo systemctl restart prox-watch-watcher.service
```

## Security Hardening

Die Service-Unit verwendet umfangreiche systemd Hardening-Optionen:

- **NoNewPrivileges:** Verhindert Privilege-Escalation
- **ProtectSystem=strict:** Schreibzugriff nur auf explizit erlaubte Pfade
- **ReadWritePaths:** Nur `/var/lib/prox-watch-watcher` (State-DB, ARM-Datei)
- **ReadOnlyPaths:** Nur `/etc/prox-watch-watcher` (Konfiguration)
- **MemoryDenyWriteExecute:** Verhindert Code-Injection
- **RestrictNamespaces:** Verhindert Namespace-Manipulation
- **LimitNOFILE/LimitNPROC:** Ressourcen-Limits

## ARM-Datei-Pfad

Die ARM-Datei für Power-Cycle muss innerhalb von `/var/lib/prox-watch-watcher/` liegen:

```bash
# ARM-Datei erstellen (für Power-Cycle)
sudo touch /var/lib/prox-watch-watcher/arm_powercycle
sudo chown prox-watch-watcher:prox-watch-watcher /var/lib/prox-watch-watcher/arm_powercycle
sudo chmod 600 /var/lib/prox-watch-watcher/arm_powercycle
```

**Wichtig:** Der Service hat nur Schreibzugriff auf `/var/lib/prox-watch-watcher/`. Die ARM-Datei muss dort liegen.

## Restart-Verhalten

- **Restart=on-failure:** Service wird nur bei Crash neu gestartet
- **Kein Restart bei Exit 0:** Normales Beenden führt nicht zu Restart
- **StartLimitIntervalSec=300:** Max. 5 Restarts innerhalb von 5 Minuten
- **StartLimitBurst=5:** Max. 5 Restarts in Folge

## Troubleshooting

### Service startet nicht

```bash
# Logs prüfen
sudo journalctl -u prox-watch-watcher.service -n 50

# Konfiguration prüfen
sudo -u prox-watch-watcher /usr/local/bin/prox-watch-watcher run --config-path /etc/prox-watch-watcher/watcher.yaml

# Berechtigungen prüfen
sudo -u prox-watch-watcher ls -la /var/lib/prox-watch-watcher/
sudo -u prox-watch-watcher cat /etc/prox-watch-watcher/watcher.yaml
```

### GPIO-Zugriff funktioniert nicht

```bash
# GPIO-Gruppe prüfen
groups prox-watch-watcher

# udev-Regeln prüfen
sudo udevadm info /dev/gpiochip0

# Service-Logs prüfen
sudo journalctl -u prox-watch-watcher.service | grep -i gpio
```

### State-DB wird nicht geschrieben

```bash
# Berechtigungen prüfen
sudo -u prox-watch-watcher ls -la /var/lib/prox-watch-watcher/

# Schreibzugriff testen
sudo -u prox-watch-watcher touch /var/lib/prox-watch-watcher/test
sudo -u prox-watch-watcher rm /var/lib/prox-watch-watcher/test
```

## Deinstallation

```bash
# Service stoppen und deaktivieren
sudo systemctl stop prox-watch-watcher.service
sudo systemctl disable prox-watch-watcher.service

# Service-Unit entfernen
sudo rm /etc/systemd/system/prox-watch-watcher.service
sudo systemctl daemon-reload

# User/Group entfernen (optional)
sudo userdel prox-watch-watcher
sudo groupdel prox-watch-watcher

# Daten entfernen (optional, VORSICHT: Löscht State-DB!)
sudo rm -rf /var/lib/prox-watch-watcher
```

## Siehe auch

- [docs/21_watcher_deployment.md](../docs/21_watcher_deployment.md) - Watcher Deployment-Anleitung
- [docs/24_powercycle_safety.md](../docs/24_powercycle_safety.md) - Power-Cycle Sicherheitsdokumentation
- [cmd/watcher/README.md](../cmd/watcher/README.md) - Watcher-Konfiguration
