## Release Notes v0.2.0

**External Watcher (Phase 1) – Out-of-Band Monitoring**

## Status

✅ **Release-fähig** – Core + Externer Wächter (Watcher) implementiert, getestet und dokumentiert

## Neue Features in v0.2.0

### ✅ Externer Wächter (Watcher)

- **Neues Binary:** `prox-watch-watcher` (Raspberry Pi / Out-of-Band)
- **Zweck:** Erkennung von Proxmox-Totalausfällen, wenn der Core selbst keine Daten mehr liefern kann

### ✅ HealthCheck (Ping + HTTPS)

- **HealthChecker:**\s
  - Ping-Check (TCP-Connect als Ersatz für ICMP, kein externes `ping`-Binary)
  - HTTPS-Check (TLS-Handshake, Status-Code irrelevant – nur Erreichbarkeit zählt)
  - Kontextbasiertes Timeout (`target.timeout_seconds`)
- **Kombinationslogik:**
  - `mode: ping` → Erfolg, wenn Ping OK
  - `mode: https` → Erfolg, wenn HTTPS OK
  - `mode: ping+https` → Erfolg, wenn **beide** OK
- **Fehlercodes (abstrakt, ohne Host/IP):**
  - `timeout`
  - `connection_failed`
  - `tls_error`

### ✅ Counter + Severity-Eskalation

- **Counter (In-Memory):**
  - Kein Persistenzspeicher (bewusst, Phase 1)
  - `FailCount` wird bei Erfolg zurückgesetzt, bei Fehler erhöht
- **Severity-Logik (Watcher-spezifisch):**
  - `failCount == 0` → `INFO`
  - `failCount >= warn_threshold` (Default 3) → `WARN`
  - `failCount >= crit_threshold` (Default 10) → `CRIT`
- **Eskalationsregeln:**
  - Push nur bei Statuswechsel **nach oben**:
    - `INFO → WARN`
    - `WARN → CRIT`
  - Kein Push bei:
    - `WARN → WARN`
    - `CRIT → CRIT`
    - `CRIT → INFO` (Recovery)
    - `WARN → INFO` (Recovery)

### ✅ Push bei Eskalation

- **PushService (Watcher):**
  - Wiederverwendung des Core-Push-Adapters (`internal/push`)
  - Event-ID: `external.availability.down` (stabil, anonym)
  - Severity-basiertes Topic-Mapping:
    - `WARN` → `prox-watch-warn`
    - `CRIT` → `prox-watch-crit`
- **Push-Regeln:**
  - INFO → **kein Push**
  - WARN → Push bei Eskalation
  - CRIT → Push bei Eskalation
- **Fehlerverhalten:**
  - Push-Fehler sind **nicht-blockierend**
  - Kein Retry im gleichen Intervall
  - Keine Eskalation bei Push-Fehlern

### ✅ GPIO-Abstraktion (NoOp)

- **GPIO-Interface:**\s
  - `SetLED(sev)` – LED-Farbe basierend auf Severity
  - `Beep()` – Beeper bei CRIT-Eskalation
  - `Close()` – Ressourcen freigeben
- **Phase 1 Implementierung:** `NoOpGPIO`
  - Keine Hardware-Abhängigkeit
  - Immer erfolgreich (`nil`)
  - Kein Zustand, keine Nebenwirkungen
- **Design (für Phase 2 vorbereitet):**
  - LED:
    - INFO → Grün
    - WARN → Gelb
    - CRIT → Rot
  - Beeper:
    - Nur bei CRIT
    - Nur bei Eskalation (INFO→CRIT / WARN→CRIT)
    - Nur tagsüber (Tag/Nacht-Logik später)

### ✅ Watcher Runner (Event-Loop)

- **Single-Thread Event-Loop:**
  - Intervall: `watcher.interval_seconds` (Default 30s)
  - Kein Mutex, keine parallelen Goroutines pro Intervall
- **Datenfluss:**
  - Ticker → HealthCheck → Counter → Severity → Push (Eskalation) → GPIO-Update
- **Graceful Shutdown:**
  - Context-basiert (SIGINT/SIGTERM via systemd möglich)
  - Ticker wird sauber gestoppt
  - Keine Hintergrund-Goroutines nach Stop

### ✅ Tests (Watcher)

- **Unit-Tests:**
  - Counter (Increment/Reset)
  - Severity-Logik (Thresholds & Priorität)
  - HealthChecker (Ping/HTTPS/Timeout)
  - PushService (Eskalationslogik)
  - GPIO (NoOp-Verhalten)
- **Integration-Tests:**
  - Stabil OK (kein Push)
  - WARN-Eskalation (ein Push)
  - CRIT-Eskalation (Push + Beep)
  - Recovery (kein Push bei Verbesserung)
  - Flattern (Fail/Success-Muster ohne Push-Spam)
  - Determinismus (gleiche Inputs → gleiche Outputs)
  - Keine Host/IP im Output

### ✅ Dokumentation (Watcher)

- `cmd/watcher/README.md`
  - Zweck, Architektur, Datenfluss, Sicherheitsgrenzen
  - Deployment-Anleitung für Raspberry Pi
  - Phase-1 Scope (implementiert vs. nicht implementiert)
- `docs/15_watcher_config_schema.md`
- `docs/16_watcher_health_engine.md`
- `docs/17_watcher_counter_severity.md`
- `docs/18_watcher_gpio.md`
- `docs/19_watcher_runner.md`
- `docs/21_watcher_deployment.md`

## Sicherheitsgarantien (v0.2.0)

### Harte Grenzen (Watcher)

- ❌ **Kein Power-Cycle**
- ❌ **Keine Persistenz** (Watcher-State ist In-Memory)
- ❌ **Kein Fernzugriff**
- ❌ **Keine automatische Reparatur**
- ❌ **Kein Auto-Reboot**
- ❌ **Kein Logging von Host/IP**

### Push-Sicherheit

- **Event-ID:** `external.availability.down` (anonym, ohne Host/IP)
- **Payload (Push):**
  - `event_id`
  - `severity`
  - `timestamp`
- **Keine:**\s
  - IP-Adressen
  - Hostnamen
  - Log-Inhalte

## Nicht enthalten in v0.2.0

- ❌ **Hardware-GPIO (LED/Beeper real)** – vorbereitet, aber nicht implementiert
- ❌ **Cooldown-Mechanismus** – kein Cooldown für wiederholte CRIT-Pushes
- ❌ **Persistenter State (Watcher)** – kein SQLite, nur In-Memory
- ❌ **Power-Cycle/Steckdosensteuerung** – bewusst gesperrt
- ❌ **Automatische Recovery/Selbstheilung**

## Kompatibilität

- **Core (v0.1.0)** bleibt unverändert:
  - Event-Processing über systemd journald
  - SQLite-State
  - CLI (init, status, ack)
  - systemd Service
- **Watcher (v0.2.0)** ist **ergänzend**, nicht ersetzend:
  - Unabhängiges Binary
  - Kann separat deployed werden (z.B. Raspberry Pi)

## Installation & Deployment

- Core:
  - Siehe `RELEASE_NOTES_v0.1.0.md` und `README.md`
- Watcher:
  - Siehe `cmd/watcher/README.md` und `docs/21_watcher_deployment.md`

## Bekannte Einschränkungen

1. **Watcher-State ist nicht persistent:**
   - Neustart setzt Counter/Severity zurück
   - Geplante Verbesserung in Phase 2 (SQLite + Cooldown)
2. **Hardware-GPIO fehlt:**
   - NoOpGPIO als Platzhalter
   - Phase 1.5/2: echte LED/Beeper-Unterstützung
3. **Journal Reader (systemd) im Core:**
   - Noch immer Fake/Platzhalter (wie in v0.1.0 beschrieben)

## Nächste Schritte (Phase 1.5 / Phase 2)

- **Phase 1.5 – Hardware GPIO (Watcher):**
  - Implementierung echter LED/Beeper-Ansteuerung (Raspberry Pi GPIO)
  - Tag/Nacht-Logik für Beeper
- **Phase 2 – Persistenz & Cooldown:**
  - SQLite-State für Watcher
  - Cooldown-Mechanismus zur weiteren Spam-Reduktion
  - Optional: Hot-Standby-Watcher

