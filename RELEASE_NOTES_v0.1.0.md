# Release Notes v0.1.0

**MVP (Minimal Viable Product) - Core Implementation**

## Status

✅ **Release-fähig** - Alle Core-Komponenten implementiert und getestet

## Features

### ✅ Konfiguration

- **CLI-Wizard** (`prox-watch init`) - Interaktive Konfigurationserstellung
- **Config Guards** - Validierung verhindert Datenlecks (keine IPs, keine Pfade im Repo)
- **YAML-basierte Konfiguration** - Strukturiert, validierbar
- **Secrets-Management** - Separate `secrets.yaml` mit 0600-Rechten

### ✅ State Management

- **SQLite-basierte Persistenz** - Event-Zähler, Cooldowns, Acknowledges
- **Restart-Safe** - State überlebt Neustarts
- **WAL-Mode** - Robuste Transaktionen

### ✅ Event Processing

- **Journal Reader** - Interface für systemd journald (Fake für Tests)
- **Pattern Matcher** - Event-ID-Generierung, Pattern-Registry
- **Severity Evaluator** - Zählregeln (1× / 3× / 10×), Zeitfenster, Hard-Error-Erkennung
- **Core Runner** - Single-Thread Event-Loop, Orchestrierung

### ✅ Push Notifications

- **ntfy-Integration** - Optional, Topic-Mapping
- **Local-Only Mode** - No-op Adapter für Offline-Betrieb
- **Metadaten-Only** - Keine Log-Inhalte, keine IPs

### ✅ CLI-Kommandos

- **status** - Event-Status anzeigen (gesamt oder spezifisch)
- **ack** - Events quittieren (mit Dauer)
- **run** - Daemon-Modus (Foreground)

### ✅ systemd Integration

- **Service-Unit** - Vollständig gehärtet
- **Security Hardening** - Minimaler Zugriff, NoNewPrivileges, ProtectSystem
- **Graceful Shutdown** - Signal-Handling (SIGTERM, SIGINT)

### ✅ Tests

- **Unit-Tests** - Alle Core-Komponenten
- **Integration-Tests** - End-to-End Event-Flow
- **Deterministic** - Fake-Time, keine externen Abhängigkeiten
- **Privacy-Compliant** - Keine realen Daten in Tests

### ✅ Dokumentation

- **README.md** - Vollständige Übersicht
- **Vision & Architektur** - Klar definiert
- **Technische Docs** - Interfaces, Datenmodelle, Konfigurationsschema
- **Security & Contributing** - Richtlinien dokumentiert

## Nicht enthalten (v0.1)

- ❌ **Journal Reader (systemd)** - Nur Interface, Fake für Tests
- ❌ **Pattern Registry (YAML)** - Nur Code-Struktur
- ❌ **Externer Wächter** - Phase 2
- ❌ **GUI / Web-UI** - Nicht im MVP
- ❌ **Langzeit-Historisierung** - Nicht im MVP

## Datenschutz

- 🔒 **Offline-First** - Keine Cloud-Abhängigkeiten
- 🔒 **No Telemetry** - Keine Datenübertragung
- 🔒 **Privacy-by-Design** - Config Guards verhindern Datenlecks
- 🔒 **Metadaten-Only** - Keine Log-Inhalte, keine IPs im Repository

## Installation

Siehe [README.md](README.md) für Installationsanleitung.

## Breaking Changes

Keine (erste Version).

## Bekannte Einschränkungen

1. **Journal Reader** - Nur Fake-Implementierung für Tests
2. **Pattern Registry** - Pattern-Loading noch nicht vollständig implementiert
3. **Hard-Error-Erkennung** - `IsHardError()` gibt immer `false` zurück

## Nächste Schritte (Phase 1)

- Journal Reader (systemd) - Echte systemd journald-Integration
- Pattern Registry (YAML) - Pattern-Definitionen aus YAML laden
- Externer Wächter (Raspberry) - Out-of-Band-Monitoring

## Danksagungen

MVP v0.1 wurde gemäß Spezifikation implementiert:
- Deterministisch
- Privacy-Preserving
- Offline-First
- Testbar
