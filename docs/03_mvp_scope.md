# MVP Scope (v0.1) - Minimaler Core

## Ziel

Funktionsfähiger Kern ohne Extras.
Offline-First, datenschutz-konform, deterministisch.

---

## ✅ Zwingend enthalten (MVP)

### 1. Konfiguration

- CLI-Wizard (`prox-watch init`)
- Lokale Speicherung (außerhalb Repo)
- Validierung (keine IPs, keine Pfade im Repo)
- Beispiel-Dateien im Repo (`config.yaml.example`)

### 2. Log-Reader

- systemd journald Streaming
- Realtime-Monitoring
- Keine Log-Speicherung (nur Pattern-Matching)

### 3. Pattern-Engine (abstrakt)

- Pattern-Definitionen (nur Metadaten im Repo)
- Lokale Regex-Zuordnung (nicht versioniert)
- Pattern-Matching gegen Log-Zeilen
- Event-ID-Generierung

### 4. State-Management

- SQLite-Datenbank (`state.db`)
- Event-Zähler
- Zeitfenster-Tracking
- Cooldown-Verwaltung
- Persistenz über Neustarts

### 5. Severity & Rules

- INFO / WARN / CRIT
- Zählregeln (1× / 3× / 10×)
- Zeitfenster (10 min / 15 min)
- Harte Fehler (sofort CRIT)

### 6. Push-Adapter

- ntfy-Integration (lokal/selbst gehostet oder freier Server)
- Topic-Mapping (info/warn/crit)
- Metadaten-Only (keine Log-Inhalte)
- Offline-First (keine Cloud-Zwänge)

### 7. Basis-Ereignisfluss

- Log → Pattern → Zähler → Severity → Push
- Cooldown-Logik
- Fehlerbehandlung (Fail-Safe)

---

## ❌ Nicht enthalten (v0.1)

### Externer Wächter

- Raspberry-Komponente
- Hardware-Signale (LED/Beeper)
- Power-Cycle-Funktionalität

### GUI / Web-UI

- Keine grafische Oberfläche
- Keine Web-Server

### Historisierung

- Keine Langzeit-Speicherung
- Keine Trend-Analyse

### Erweiterte Features

- Keine Korrelation
- Keine Eskalation
- Keine Acknowledge-Funktion
- Keine Profile-Verwaltung

### Observability-Add-ons

- Kein Zabbix
- Kein Prometheus
- Kein Grafana

---

## MVP-Ausgabe

### Funktionen

1. ✅ Konfiguration anlegen (CLI)
2. ✅ Logs lesen (journald)
3. ✅ Patterns matchen
4. ✅ Events zählen
5. ✅ Severity bewerten
6. ✅ Push senden (ntfy)

### Dateien (lokal, nicht versioniert)

- `config.yaml`
- `secrets.yaml`
- `state.db`

### Dateien (Repo)

- `config/config.yaml.example`
- Pattern-Definitionen (Metadaten)
- Dokumentation

---

## Erfolgskriterien (MVP)

1. ✅ System startet ohne Fehler
2. ✅ Konfiguration wird lokal gespeichert
3. ✅ Logs werden gelesen
4. ✅ Patterns werden getroffen
5. ✅ Events werden gezählt
6. ✅ Push wird bei CRIT gesendet
7. ✅ Keine Datenlecks
8. ✅ Vollständig offline betreibbar

---

## Nächste Schritte (nach MVP)

- Externer Wächter (Raspberry)
- Acknowledge-Funktion
- Erweiterte Pattern-Typen
- Optional: GUI
- Optional: Observability-Add-ons

---

## Status

- MVP-Scope definiert
- Implementierung noch nicht gestartet
- Architektur: Hybrid (N4)
