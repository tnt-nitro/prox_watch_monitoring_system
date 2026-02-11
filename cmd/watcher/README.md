# prox-watch-watcher

Externer Wächter für Out-of-Band-Monitoring (Raspberry Pi).

## Status

**Phase 1 - Implementiert** ✅

## Zweck

Der **prox-watch-watcher** ist ein unabhängiges Monitoring-System, das **Out-of-Band** arbeitet. Er erkennt Totalausfälle, wenn Proxmox selbst nicht mehr antwortet.

**Hauptfunktionen:**
- ✅ Erreichbarkeitsprüfung (Ping + HTTPS)
- ✅ Fehlzähler (WARN ≥3, CRIT ≥10)
- ✅ Push-Benachrichtigungen (ntfy)
- ✅ Lokale Signale (LED + Beeper, optional)

**Nicht enthalten (Phase 1):**
- ❌ Log-Analyse
- ❌ Deep-Monitoring
- ❌ Automatische Reparatur
- ❌ Power-Cycle (nur Design, keine Umsetzung)

## Architektur

### Komponenten

```
prox-watch-watcher (Single Binary)
 ├─ HealthChecker      # Ping + HTTPS-Check
 ├─ Counter            # Fehlzähler (In-Memory)
 ├─ Severity Evaluator # INFO/WARN/CRIT
 ├─ PushService        # Push-Benachrichtigungen
 ├─ GPIO               # LED/Beeper (NoOp in Phase 1)
 └─ Runner             # Event-Loop (30s Intervall)
```

### Datenfluss

```
Ticker (30s)
 └─ Health Check
     ├─ Success → Counter.Reset() → Severity = INFO
     └─ Failure → Counter.Increment() → Severity evaluieren
         └─ Severity bestimmen (WARN/CRIT basierend auf Thresholds)
             ├─ Push (nur bei Eskalation: INFO→WARN, WARN→CRIT)
             └─ GPIO Update (LED + Beep bei CRIT-Eskalation)
```

### Unterschiede zum Core

| Aspekt | Core (prox-watch) | Watcher (prox-watch-watcher) |
|--------|-------------------|------------------------------|
| **Datenquelle** | systemd Journal | Health-Check (Ping/HTTPS) |
| **State** | SQLite (persistent) | In-Memory (nicht persistent) |
| **Event-IDs** | Dynamisch (Pattern-basiert) | Fest: `external.availability.down` |
| **Intervall** | Echtzeit (Journal-Stream) | 30s (konfigurierbar) |
| **GPIO** | Nicht vorhanden | LED + Beeper (optional) |
| **Deployment** | Proxmox-Host | Raspberry Pi (extern) |

## Sicherheitsgrenzen

### Absolute Verbote

- ❌ **Kein Auto-Reboot** - Keine automatischen System-Neustarts
- ❌ **Keine Persistenz** - Keine dauerhafte Speicherung (Phase 1: In-Memory)
- ❌ **Kein Fernzugriff** - Keine Remote-Steuerung ohne lokale Freigabe
- ❌ **Kein Logging sensibler Daten** - Keine Host/IP in Logs oder Push-Nachrichten
- ❌ **Kein Power-Cycle** - Keine automatischen Hardware-Aktionen (Phase 1)

### Erlaubt

- ✔ **Push-Benachrichtigungen** (Metadaten-only, keine Host/IP)
- ✔ **LED-Statusanzeige** (lokal, keine Datenübertragung)
- ✔ **Beeper** (nur CRIT, nur Tag, nur bei Eskalation)

## Limitierungen (Phase 1)

### Bewusst nicht implementiert

1. **Power-Cycle:**
   - Design vorhanden, aber **nicht implementiert**
   - Erfordert manuelle Freigabe (später)
   - Max. 1× / 24h (später)

2. **Persistenz:**
   - State ist In-Memory
   - Neustart → Counter wird zurückgesetzt
   - Phase 2: Optional SQLite-Persistenz

3. **Hardware-GPIO:**
   - Phase 1: Nur NoOp-GPIO (immer erfolgreich)
   - Hardware-Implementierung (Raspberry Pi) für Phase 2 geplant

4. **Cooldown:**
   - Phase 1: Kein Cooldown-Mechanismus
   - Push wird bei jedem Statuswechsel ausgelöst

## Deployment (Raspberry Pi)

### Voraussetzungen

- Raspberry Pi (beliebiges Modell)
- Go 1.21+ (für Build)
- systemd (optional, für Service)

### Build

```bash
# Im Projekt-Root
go build -o prox-watch-watcher ./cmd/watcher
```

### Installation

1. **Binary kopieren:**
   ```bash
   sudo cp prox-watch-watcher /usr/local/bin/
   sudo chmod 755 /usr/local/bin/prox-watch-watcher
   ```

2. **Konfiguration erstellen:**
   ```bash
   sudo mkdir -p /etc/prox-watch-watcher
   sudo cp config/watcher.yaml.example /etc/prox-watch-watcher/watcher.yaml
   sudo nano /etc/prox-watch-watcher/watcher.yaml
   ```
   
   **Wichtig:** Ersetze `PLACEHOLDER` durch echten Hostname/IP (lokal, nicht im Repo!)

3. **systemd Service (optional, später):**
   ```bash
   sudo cp installer/prox-watch-watcher.service /etc/systemd/system/
   sudo systemctl daemon-reload
   sudo systemctl enable prox-watch-watcher
   sudo systemctl start prox-watch-watcher
   ```

### Konfiguration

**Beispiel `watcher.yaml`:**
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

**Wichtig:**
- `host`: Lokal ersetzen (keine IP im Repo!)
- `gpio.enabled: false` → NoOp-GPIO (Phase 1)
- `push.enabled: true` → Push-Benachrichtigungen aktivieren

### Verwendung

```bash
# Manueller Start (Foreground)
./prox-watch-watcher run

# Status anzeigen
./prox-watch-watcher status

# Version anzeigen
./prox-watch-watcher version
```

## Phase-1 Scope

### Implementiert (v0.1)

- ✅ Health-Check (Ping + HTTPS)
- ✅ Counter (In-Memory)
- ✅ Severity-Evaluierung (INFO/WARN/CRIT)
- ✅ Push-Benachrichtigungen (ntfy)
- ✅ GPIO-Interface (NoOp-Implementierung)
- ✅ Event-Loop (30s Intervall)
- ✅ Graceful Shutdown

### Nicht implementiert (Phase 1)

- ❌ **Power-Cycle** - Bewusst gesperrt, Design vorhanden
- ❌ **Hardware-GPIO** - Nur NoOp, Hardware-Implementierung später
- ❌ **Persistenz** - In-Memory, SQLite später optional
- ❌ **Cooldown** - Push bei jedem Statuswechsel

### Phase-2 Erweiterung (geplant)

- 🔵 Hardware-GPIO (Raspberry Pi)
- 🔵 SQLite-Persistenz (optional)
- 🔵 Cooldown-Mechanismus
- 🔵 Power-Cycle (mit manueller Freigabe)
- 🔵 systemd Service-Unit

## Troubleshooting

### Keine Push-Benachrichtigungen

- Prüfe `push.enabled: true` in Konfiguration
- Prüfe ntfy-Server-Erreichbarkeit
- Prüfe Topics (`prox-watch-warn`, `prox-watch-crit`)

### Health-Check schlägt fehl

- Prüfe `target.host` und `target.port`
- Prüfe Netzwerk-Erreichbarkeit
- Prüfe `target.timeout_seconds` (zu kurz?)

### Counter wird nicht zurückgesetzt

- Normal bei In-Memory State
- Neustart → Counter wird zurückgesetzt
- Phase 2: Optional SQLite-Persistenz

## Weitere Dokumentation

- [Architektur-Übersicht](../../docs/01_architecture.md)
- [Watcher-Konfigurationsschema](../../docs/15_watcher_config_schema.md)
- [Health-Check-Engine](../../docs/16_watcher_health_engine.md)
- [Counter & Severity-Flow](../../docs/17_watcher_counter_severity.md)
- [GPIO-Abstraktion](../../docs/18_watcher_gpio.md)
- [Runner-Modell](../../docs/19_watcher_runner.md)
