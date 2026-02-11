# systemd Service-Definition (Konzept)

## Ziel
Sicher. Robust. Minimal.

---

## 1️⃣ Service-Unit

### `prox-watch.service`

**Pfad:**
```
/etc/systemd/system/prox-watch.service
```

**Inhalt:**
```ini
[Unit]
Description=Prox Watch Monitoring System
Documentation=https://github.com/.../prox-watch
After=network.target systemd-journald.service
Wants=systemd-journald.service

[Service]
Type=simple
ExecStart=/usr/local/bin/prox-watch
Restart=on-failure
RestartSec=10
TimeoutStopSec=30

# User & Group
User=prox-watch
Group=prox-watch

# Working Directory
WorkingDirectory=/var/lib/prox-watch

# Security
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/prox-watch
ReadOnlyPaths=/var/log/journal

# Capabilities
CapabilityBoundingSet=
AmbientCapabilities=

# Network
RestrictAddressFamilies=AF_UNIX AF_INET AF_INET6
RestrictNamespaces=true

# Process
LimitNOFILE=1024
LimitNPROC=64

# Logging
StandardOutput=journal
StandardError=journal
SyslogIdentifier=prox-watch

[Install]
WantedBy=multi-user.target
```

---

## 2️⃣ User & Group

### Erstellung

**User:**
```bash
useradd -r -s /usr/sbin/nologin prox-watch
```

**Group:**
```bash
groupadd -r prox-watch
```

**Verzeichnis:**
```bash
mkdir -p /var/lib/prox-watch
chown prox-watch:prox-watch /var/lib/prox-watch
chmod 700 /var/lib/prox-watch
```

---

## 3️⃣ Security-Hardening

### `NoNewPrivileges`

**Zweck:**
- Verhindert Privilege-Escalation
- Keine SUID/SGID-Ausführung

**Wert:**
- `true`

---

### `PrivateTmp`

**Zweck:**
- Isolierter /tmp
- Keine Zugriffe auf andere Prozesse

**Wert:**
- `true`

---

### `ProtectSystem`

**Zweck:**
- Schreibschutz für System-Verzeichnisse
- Nur /var/lib/prox-watch schreibbar

**Wert:**
- `strict`

**Erlaubt:**
- `/var/lib/prox-watch` (ReadWritePaths)

---

### `ProtectHome`

**Zweck:**
- Kein Zugriff auf Home-Verzeichnisse
- Keine User-Daten

**Wert:**
- `true`

---

### `ReadWritePaths`

**Zweck:**
- Explizite Schreibpfade
- Minimaler Zugriff

**Werte:**
- `/var/lib/prox-watch`

---

### `ReadOnlyPaths`

**Zweck:**
- Explizite Lese-Pfade
- Nur Journal-Zugriff

**Werte:**
- `/var/log/journal`

---

## 4️⃣ Capabilities

### `CapabilityBoundingSet`

**Zweck:**
- Keine Capabilities
- Minimaler Zugriff

**Wert:**
- Leer (keine Capabilities)

---

### `AmbientCapabilities`

**Zweck:**
- Keine ambient Capabilities
- Keine Vererbung

**Wert:**
- Leer (keine Capabilities)

---

## 5️⃣ Network-Restrictions

### `RestrictAddressFamilies`

**Zweck:**
- Nur erlaubte Adressfamilien
- Keine exotischen Protokolle

**Werte:**
- `AF_UNIX` (Unix-Sockets)
- `AF_INET` (IPv4)
- `AF_INET6` (IPv6)

---

### `RestrictNamespaces`

**Zweck:**
- Keine Namespace-Erstellung
- Keine Container-Isolation

**Wert:**
- `true`

---

## 6️⃣ Process-Limits

### `LimitNOFILE`

**Zweck:**
- Maximale offene Dateien
- Ressourcen-Begrenzung

**Wert:**
- `1024`

---

### `LimitNPROC`

**Zweck:**
- Maximale Prozesse
- Keine Fork-Bombs

**Wert:**
- `64`

---

## 7️⃣ Restart-Verhalten

### `Restart`

**Zweck:**
- Automatischer Neustart bei Fehlern
- Kein Neustart bei normalem Shutdown

**Wert:**
- `on-failure`

---

### `RestartSec`

**Zweck:**
- Wartezeit vor Neustart
- Vermeidet Restart-Loops

**Wert:**
- `10` (Sekunden)

---

### `TimeoutStopSec`

**Zweck:**
- Maximale Wartezeit beim Stoppen
- Danach SIGKILL

**Wert:**
- `30` (Sekunden)

---

## 8️⃣ Logging

### `StandardOutput`

**Zweck:**
- stdout → journald
- Zentrale Logs

**Wert:**
- `journal`

---

### `StandardError`

**Zweck:**
- stderr → journald
- Fehler-Logs

**Wert:**
- `journal`

---

### `SyslogIdentifier`

**Zweck:**
- Identifikation in Logs
- Filterbar

**Wert:**
- `prox-watch`

---

## 9️⃣ Dependencies

### `After`

**Zweck:**
- Start-Reihenfolge
- Nach Netzwerk und Journal

**Werte:**
- `network.target`
- `systemd-journald.service`

---

### `Wants`

**Zweck:**
- Optionale Dependencies
- Kein Fehler bei Ausfall

**Werte:**
- `systemd-journald.service`

---

## 🔟 Installation

### Aktivierung

```bash
# Unit kopieren
cp prox-watch.service /etc/systemd/system/

# systemd neu laden
systemctl daemon-reload

# Service aktivieren
systemctl enable prox-watch.service

# Service starten
systemctl start prox-watch.service
```

---

## 1️⃣1️⃣ Status-Prüfung

### Kommandos

```bash
# Status
systemctl status prox-watch.service

# Logs
journalctl -u prox-watch.service -f

# Neustart
systemctl restart prox-watch.service

# Stoppen
systemctl stop prox-watch.service
```

---

## 1️⃣2️⃣ Timer-Unit (optional, später)

### `prox-watch.timer`

**Zweck:**
- Periodische Checks
- Externer Wächter

**Status:**
- Nicht in MVP v0.1

---

## 1️⃣3️⃣ Socket-Unit (optional, später)

### `prox-watch.socket`

**Zweck:**
- Socket-Aktivierung
- On-Demand-Start

**Status:**
- Nicht in MVP v0.1

---

## 1️⃣4️⃣ Hardening-Checkliste

### ✅ Implementiert

- [x] NoNewPrivileges
- [x] PrivateTmp
- [x] ProtectSystem
- [x] ProtectHome
- [x] ReadWritePaths (minimal)
- [x] ReadOnlyPaths (minimal)
- [x] CapabilityBoundingSet (leer)
- [x] RestrictAddressFamilies
- [x] RestrictNamespaces
- [x] Process-Limits
- [x] Dedicated User/Group

### ❌ Nicht in MVP

- [ ] AppArmor-Profile
- [ ] SELinux-Policy
- [ ] Seccomp-Filter
- [ ] SystemCall-Filter

---

## 1️⃣5️⃣ Fehlerbehandlung

### Start-Fehler

**Ursachen:**
- Fehlende Rechte
- Ungültige Konfiguration
- Fehlende Dependencies

**Verhalten:**
- systemd loggt Fehler
- Restart nach RestartSec
- Max. 5 Versuche (systemd-default)

---

## Status

- Service-Definition spezifiziert
- Hardening definiert
- Rechte minimal
- Sicherheit maximiert
