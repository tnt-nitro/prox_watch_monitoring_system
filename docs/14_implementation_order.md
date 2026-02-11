# Implementierungsreihenfolge (Tasks für v0.1)

## Ziel
Logische Abhängigkeiten. Schrittweise. Testbar.

---

## 1️⃣ Phase 1: Grundlagen

### Task 1.1: Projekt-Setup

**Ziel:**
- Go-Modul initialisieren
- Verzeichnisstruktur anlegen
- Basis-Imports

**Schritte:**
1. `go mod init prox-watch`
2. Verzeichnisse erstellen (`cmd/`, `internal/`, `pkg/`)
3. Basis-Dateien anlegen
4. `.gitignore` aktualisieren

**Abhängigkeiten:**
- Keine

**Ergebnis:**
- Projektstruktur vorhanden
- Go-Modul funktionsfähig

---

### Task 1.2: Datenmodelle

**Ziel:**
- Enums definieren (Severity)
- Structs definieren (Event, CountState, etc.)
- Type-Definitionen

**Schritte:**
1. `internal/state/model.go` (Structs)
2. `internal/rules/severity.go` (Enums)
3. Gemeinsame Typen

**Abhängigkeiten:**
- Task 1.1

**Ergebnis:**
- Datenmodelle definiert
- Type-safe

---

### Task 1.3: Interfaces

**Ziel:**
- Interfaces definieren
- Verträge festlegen

**Schritte:**
1. `internal/journal/reader.go` (Interface)
2. `internal/pattern/matcher.go` (Interface)
3. `internal/state/store.go` (Interface)
4. `internal/push/adapter.go` (Interface)
5. `internal/rules/evaluator.go` (Interface)

**Abhängigkeiten:**
- Task 1.2

**Ergebnis:**
- Interfaces definiert
- Mock-fähig

---

## 2️⃣ Phase 2: Core-Komponenten

### Task 2.1: Config-Loader

**Ziel:**
- Konfiguration laden
- Validierung
- Guards

**Schritte:**
1. `internal/config/loader.go`
2. `internal/config/validate.go`
3. YAML-Parsing
4. Validierungsregeln
5. Guards (Repo-Pfad, IPs)

**Abhängigkeiten:**
- Task 1.3

**Ergebnis:**
- Config geladen
- Validierung funktioniert
- Guards aktiv

---

### Task 2.2: State-Store (SQLite)

**Ziel:**
- SQLite-Integration
- CRUD-Operationen
- Persistenz

**Schritte:**
1. `internal/state/store.go` (Implementierung)
2. Schema-Erstellung
3. Increment/Get/SetCooldown
4. Acknowledge-Funktionen
5. Tests (Mock + SQLite)

**Abhängigkeiten:**
- Task 1.2
- Task 1.3

**Ergebnis:**
- State-Store funktioniert
- Persistenz gewährleistet
- Tests vorhanden

---

### Task 2.3: Severity-Evaluator

**Ziel:**
- Zählregeln
- Zeitfenster
- Severity-Bewertung

**Schritte:**
1. `internal/rules/severity.go` (Implementierung)
2. Threshold-Logik
3. Window-Logik
4. Harte Fehler-Erkennung
5. Tests

**Abhängigkeiten:**
- Task 1.2

**Ergebnis:**
- Severity-Bewertung funktioniert
- Deterministisch
- Tests vorhanden

---

## 3️⃣ Phase 3: Log-Processing

### Task 3.1: Journal-Reader

**Ziel:**
- systemd journald Streaming
- Event-Generierung

**Schritte:**
1. `internal/journal/reader.go` (Implementierung)
2. journald-Integration
3. Streaming-Logik
4. Event-Metadaten-Extraktion
5. Tests (Mock)

**Abhängigkeiten:**
- Task 1.3

**Ergebnis:**
- Journal-Reader funktioniert
- Events generiert
- Tests vorhanden

---

### Task 3.2: Pattern-Matcher

**Ziel:**
- Pattern-Matching
- Event-ID-Generierung
- Abstrakte Patterns

**Schritte:**
1. `internal/pattern/matcher.go` (Implementierung)
2. `internal/pattern/registry.go`
3. Pattern-Laden (Metadaten)
4. Matching-Logik
5. Event-ID-Generierung
6. Tests (Mock)

**Abhängigkeiten:**
- Task 1.3
- Task 3.1

**Ergebnis:**
- Pattern-Matching funktioniert
- Event-IDs generiert
- Tests vorhanden

---

## 4️⃣ Phase 4: Integration

### Task 4.1: Core-Runner

**Ziel:**
- Event-Loop
- Komponenten-Koordination
- Orchestrierung

**Schritte:**
1. `internal/core/runner.go`
2. `internal/core/lifecycle.go`
3. Event-Loop
4. Komponenten-Integration
5. Fehlerbehandlung
6. Tests (Integration)

**Abhängigkeiten:**
- Task 2.1
- Task 2.2
- Task 2.3
- Task 3.1
- Task 3.2

**Ergebnis:**
- Core-Runner funktioniert
- Event-Flow implementiert
- Tests vorhanden

---

### Task 4.2: Push-Adapter (ntfy)

**Ziel:**
- ntfy-Integration
- Topic-Mapping
- Metadaten-Versand

**Schritte:**
1. `internal/push/ntfy.go` (Implementierung)
2. HTTP-Client
3. Topic-Mapping
4. Message-Format
5. Retry-Logik (optional)
6. Tests (Mock)

**Abhängigkeiten:**
- Task 1.3
- Task 2.1

**Ergebnis:**
- Push-Adapter funktioniert
- Nachrichten versendet
- Tests vorhanden

---

## 5️⃣ Phase 5: CLI

### Task 5.1: CLI-Wizard (init)

**Ziel:**
- Konfigurations-Wizard
- Validierung
- Lokale Speicherung

**Schritte:**
1. `internal/cli/init.go`
2. Fragen-Abfrage
3. Validierung
4. Datei-Schreibung
5. Rechte setzen
6. Tests

**Abhängigkeiten:**
- Task 2.1

**Ergebnis:**
- Wizard funktioniert
- Config erstellt
- Tests vorhanden

---

### Task 5.2: CLI-Kommandos (status, ack)

**Ziel:**
- Status-Anzeige
- Acknowledge-Funktion

**Schritte:**
1. `internal/cli/status.go`
2. `internal/cli/ack.go`
3. State-Abfrage
4. Acknowledge-Setzung
5. Tests

**Abhängigkeiten:**
- Task 2.2
- Task 2.1

**Ergebnis:**
- CLI-Kommandos funktionieren
- Tests vorhanden

---

### Task 5.3: Main-Entry-Point

**Ziel:**
- CLI-Parsing
- Signal-Handling
- Daemon-Modus

**Schritte:**
1. `cmd/prox-watch/main.go`
2. CLI-Parser
3. Signal-Handler (SIGTERM, SIGINT)
4. Daemon-Modus
5. Foreground-Modus
6. Tests (Smoke)

**Abhängigkeiten:**
- Task 4.1
- Task 5.1
- Task 5.2

**Ergebnis:**
- Main funktioniert
- Daemon-Modus funktioniert
- Tests vorhanden

---

## 6️⃣ Phase 6: systemd & Installer

### Task 6.1: systemd Service-Unit

**Ziel:**
- Service-Definition
- Hardening
- Dependencies

**Schritte:**
1. `installer/prox-watch.service` erstellen
2. Hardening-Optionen
3. Dependencies
4. Test-Installation

**Abhängigkeiten:**
- Task 5.3

**Ergebnis:**
- Service-Unit vorhanden
- Hardening aktiv
- Testbar

---

### Task 6.2: Installer-Script

**Ziel:**
- One-Shot-Installer
- Offline
- Rollback

**Schritte:**
1. `installer/install.sh` erstellen
2. Voraussetzungen prüfen
3. User/Verzeichnisse
4. Binary installieren
5. Service installieren
6. Wizard ausführen
7. Rollback-Logik
8. Tests

**Abhängigkeiten:**
- Task 6.1
- Task 5.1

**Ergebnis:**
- Installer funktioniert
- Offline
- Rollback möglich

---

## 7️⃣ Phase 7: Tests & Dokumentation

### Task 7.1: Unit-Tests

**Ziel:**
- Coverage ≥ 90% (Core)
- Alle Komponenten getestet

**Schritte:**
1. Tests für alle Komponenten
2. Mock-Objekte
3. Edge-Cases
4. Coverage-Report

**Abhängigkeiten:**
- Alle vorherigen Tasks

**Ergebnis:**
- Tests vorhanden
- Coverage-Ziel erreicht

---

### Task 7.2: Integration-Tests

**Ziel:**
- End-to-End-Tests
- Komponenten-Integration

**Schritte:**
1. Integration-Tests
2. Mock-Interfaces
3. Event-Flow-Tests
4. Fehler-Szenarien

**Abhängigkeiten:**
- Task 7.1

**Ergebnis:**
- Integration-Tests vorhanden
- Event-Flow getestet

---

### Task 7.3: Dokumentation aktualisieren

**Ziel:**
- README.md ausfüllen
- docs/ aktualisieren
- Beispiele (anonym)

**Schritte:**
1. README.md vervollständigen
2. docs/01_architecture.md ausfüllen
3. API-Dokumentation
4. Beispiele hinzufügen

**Abhängigkeiten:**
- Alle vorherigen Tasks

**Ergebnis:**
- Dokumentation vollständig
- Beispiele vorhanden

---

## 8️⃣ Implementierungsreihenfolge (Zusammenfassung)

### Woche 1: Grundlagen
- Task 1.1: Projekt-Setup
- Task 1.2: Datenmodelle
- Task 1.3: Interfaces

### Woche 2: Core-Komponenten
- Task 2.1: Config-Loader
- Task 2.2: State-Store
- Task 2.3: Severity-Evaluator

### Woche 3: Log-Processing
- Task 3.1: Journal-Reader
- Task 3.2: Pattern-Matcher

### Woche 4: Integration
- Task 4.1: Core-Runner
- Task 4.2: Push-Adapter

### Woche 5: CLI
- Task 5.1: CLI-Wizard
- Task 5.2: CLI-Kommandos
- Task 5.3: Main-Entry-Point

### Woche 6: systemd & Installer
- Task 6.1: systemd Service
- Task 6.2: Installer-Script

### Woche 7: Tests & Dokumentation
- Task 7.1: Unit-Tests
- Task 7.2: Integration-Tests
- Task 7.3: Dokumentation

---

## 9️⃣ Kritischer Pfad

**Minimal für MVP:**
1. Task 1.1-1.3 (Grundlagen)
2. Task 2.1-2.3 (Core)
3. Task 3.1-3.2 (Log-Processing)
4. Task 4.1 (Core-Runner)
5. Task 5.3 (Main)

**Optional für MVP:**
- Task 4.2 (Push kann später)
- Task 5.1-5.2 (CLI kann später)
- Task 6.1-6.2 (Installer kann später)

---

## 🔟 Abhängigkeits-Graph

```
1.1 (Setup)
  ↓
1.2 (Modelle) → 1.3 (Interfaces)
  ↓              ↓
2.1 (Config)   2.2 (State)   2.3 (Severity)
  ↓              ↓              ↓
3.1 (Journal)  3.2 (Pattern)
  ↓              ↓
4.1 (Runner) ← 4.2 (Push)
  ↓
5.1 (Wizard)   5.2 (CLI)   5.3 (Main)
  ↓              ↓           ↓
6.1 (Service)  6.2 (Installer)
  ↓
7.1-7.3 (Tests & Docs)
```

---

## Status

- Implementierungsreihenfolge definiert
- Abhängigkeiten klar
- Schrittweise umsetzbar
- Testbar
