# prox-watch

**Proxmox Monitoring System** - Offline-First, Privacy-Preserving, Deterministic

## Status

✅ **MVP v0.1.0** - Core implementation complete
✅ **v0.2.0** - External Watcher (Phase 1) complete
✅ **v0.3.0** - Hardware GPIO (Phase 1.5) complete
✅ **v0.4.0** - Persistenz & Cooldown (Phase 2) complete

## Was ist es / was nicht

### Was es ist

- **Offline-First Monitoring** für Proxmox-Server
- **Event-Driven** Log-Analyse über systemd journald
- **Privacy-Preserving** - keine Log-Inhalte, keine IPs, keine Hostnames im Repository
- **Deterministic** - reproduzierbare Severity-Bewertung
- **Local-Only** - vollständig offline betreibbar

### Was es nicht ist

- ❌ Keine Cloud-Dienste
- ❌ Keine Telemetrie
- ❌ Keine automatische Recovery
- ❌ Keine GUI/Web-UI (MVP)
- ❌ Keine Langzeit-Historisierung (MVP)

## Schnellstart

### Installation

1. **Konfiguration initialisieren:**
   ```bash
   sudo prox-watch init
   ```

2. **systemd Service installieren:**
   ```bash
   sudo cp installer/prox-watch.service /etc/systemd/system/
   sudo systemctl daemon-reload
   sudo systemctl enable prox-watch.service
   sudo systemctl start prox-watch.service
   ```

3. **Status prüfen:**
   ```bash
   sudo prox-watch status
   sudo systemctl status prox-watch.service
   ```

### Konfiguration

- **Config:** `/var/lib/prox-watch/config.yaml`
- **State DB:** `/var/lib/prox-watch/state.db`
- **Secrets:** `/var/lib/prox-watch/secrets.yaml`

Siehe [config/config.yaml.example](config/config.yaml.example) für Beispiel-Konfiguration.

## Datenschutz-Prinzipien

- 🔒 **Offline-First** - keine externen Abhängigkeiten
- 🔒 **No Telemetry** - keine Datenübertragung
- 🔒 **No Cloud** - vollständig lokal
- 🔒 **Metadaten-Only** - keine Log-Inhalte im Repository
- 🔒 **Privacy-by-Design** - Config Guards verhindern Datenlecks

Siehe [SECURITY.md](SECURITY.md) für Details.

## Architektur

### Hybrid-Ansatz (N4)

- **Lean Core** (`prox-watch`) - Event-Processing, State-Management, Push
- **Optional Observability** - externe Tools (Zabbix, Prometheus, etc.) möglich

### Komponenten

1. **Journal Reader** - systemd journald Streaming
2. **Pattern Matcher** - Event-ID-Generierung
3. **State Store** - SQLite-basierte Persistenz
4. **Severity Evaluator** - Zählregeln & Zeitfenster
5. **Push Adapter** - ntfy-Integration (optional)

### Datenfluss

```
Journal → Pattern → State → Severity → Push
```

Siehe [docs/](docs/) für detaillierte Architektur-Dokumentation.

## Dokumentation

- [docs/00_vision.md](docs/00_vision.md) - Vision & Scope
- [docs/03_mvp_scope.md](docs/03_mvp_scope.md) - MVP v0.1 Scope
- [docs/05_module_structure.md](docs/05_module_structure.md) - Go-Modulstruktur
- [docs/06_interfaces.md](docs/06_interfaces.md) - Interfaces
- [docs/08_config_schema.md](docs/08_config_schema.md) - Konfigurationsschema
- [docs/10_systemd_service.md](docs/10_systemd_service.md) - systemd Service
- [docs/12_test_strategy.md](docs/12_test_strategy.md) - Teststrategie

## Contributing

Bitte beachten Sie die **Datenschutz-Regeln**:

- ❌ Keine IP-Adressen
- ❌ Keine realen Hostnames
- ❌ Keine Log-Inhalte
- ❌ Keine Secrets

Siehe [CONTRIBUTING.md](CONTRIBUTING.md) für Details.

## Security

- **Hardening** - systemd Service mit minimalen Rechten
- **Config Guards** - Validierung verhindert Datenlecks
- **Offline-First** - keine Netzwerk-Abhängigkeiten (außer ntfy optional)

Siehe [SECURITY.md](SECURITY.md) für Details.

## Roadmap

### ✅ Phase 0: Architektur & Spezifikation
- Repository-Struktur
- Interfaces & Datenmodelle
- Konfigurationsschema

### ✅ Phase 1: MVP v0.1 (Core)
- Config Loader + Guards
- State Store (SQLite)
- Severity Rules
- Journal Reader (Fake)
- Pattern Matcher
- Core Runner
- Push Adapter (ntfy)
- CLI-Kommandos (status, ack)
- systemd Service-Unit
- Tests (Unit + Komponenten)
- Docs (minimal)

### ✅ Phase 1.0: Externer Wächter (Watcher) - v0.2.0
- Health-Check (Ping + HTTPS)
- Counter (In-Memory)
- Severity-Evaluierung
- Push-Benachrichtigungen
- GPIO-Interface (NoOp)
- Runner (Event-Loop)
- Integration-Tests

### ✅ Phase 1.5: Hardware GPIO - v0.3.0
- Raspberry GPIO (periph.io Integration)
- LED-Statusanzeige (INFO/WARN/CRIT)
- Beeper mit Eskalations-Trigger, Zeitfenster, Maximaldauer
- Atomic Concurrency-Schutz
- Build-Tags (raspberry / default mock)
- Vollständige Hardware-Testabdeckung (MockPin)

Siehe [cmd/watcher/README.md](cmd/watcher/README.md) und [RELEASE_NOTES_v0.3.0.md](RELEASE_NOTES_v0.3.0.md) für Details.

### ⏳ Phase 2: Erweiterungen
- Journal Reader (systemd)
- Pattern Registry (YAML)
- Hardware-GPIO (Raspberry Pi)
- SQLite-Persistenz (Watcher)
- Cooldown-Mechanismus
- Power-Cycle (mit manueller Freigabe)

### ⏳ Phase 3: Observability (optional)
- Zabbix-Integration
- Prometheus-Export
- Grafana-Dashboards

## Lizenz

(TBD)
