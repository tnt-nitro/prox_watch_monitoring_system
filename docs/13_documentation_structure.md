# Dokumentations-MVP (README & Docs-Outline)

## Ziel
Struktur klar. Inhalte später. Übersichtlich.

---

## 1️⃣ Root-README.md

### Struktur

**Sektionen:**
1. **Projektname & Status**
   - Name: prox_watch_monitoring_system
   - Status: WIP - Architektur offen (aktuell: Phase 0 abgeschlossen)

2. **Was ist es / was nicht**
   - Was: Offline-First Monitoring für Proxmox
   - Was nicht: Keine Cloud, keine Telemetrie, keine automatische Recovery

3. **Schnellstart (später)**
   - Installation
   - Konfiguration
   - Start

4. **Datenschutz-Prinzipien**
   - Offline-First
   - No Telemetry
   - No Cloud
   - Lokale Speicherung

5. **Architektur-Übersicht**
   - Hybrid-Ansatz (N4)
   - Komponenten
   - Datenfluss (kurz)

6. **Dokumentation**
   - Link zu `docs/`
   - Übersicht der Dokumente

7. **Contributing**
   - Link zu `CONTRIBUTING.md`
   - Datenschutz-Regeln

8. **Security**
   - Link zu `SECURITY.md`
   - Grundprinzipien

9. **Roadmap**
   - Phase 0: ✅ Abgeschlossen
   - Phase 1: 🟡 In Arbeit
   - Phase 2-4: ⏳ Geplant

10. **Lizenz**
    - (später)

---

## 2️⃣ docs/ - Übersicht

### Struktur

```
docs/
├─ 00_vision.md              # Vision & Scope
├─ 01_architecture.md         # Architektur (aktuell nur Überschriften)
├─ 02_architecture_options.md # Architektur-Vergleich
├─ 03_mvp_scope.md           # MVP-Scope v0.1
├─ 04_process_model.md       # Prozessmodell
├─ 05_module_structure.md    # Go-Modulstruktur
├─ 06_interfaces.md          # Interfaces
├─ 07_data_models.md         # Datenmodelle
├─ 08_config_schema.md       # Konfigurationsschema
├─ 09_validation_rules.md    # Validierungsregeln
├─ 10_systemd_service.md     # systemd Service
├─ 11_installer_concept.md  # Installer-Konzept
├─ 12_test_strategy.md       # Teststrategie
├─ 13_documentation_structure.md # Diese Datei
└─ diagrams/
   └─ prox_watch_overview.png # Übersichtsdiagramm (Platzhalter)
```

---

## 3️⃣ docs/00_vision.md

### Inhalt

**Sektionen:**
1. **Problemstellung**
   - Was wird überwacht
   - Warum offline
   - Warum datenschutz-konform

2. **Zielsetzung**
   - Hauptziele
   - Erfolgskriterien

3. **Nicht-Ziele**
   - Was nicht implementiert wird
   - Was explizit ausgeschlossen ist

4. **Grundprinzipien**
   - Offline-First
   - Datenschutz
   - Determinismus

---

## 4️⃣ docs/01_architecture.md

### Inhalt (aktuell nur Überschriften)

**Sektionen:**
1. **Komponenten**
   - (später ausfüllen)

2. **Datenfluss**
   - (später ausfüllen)

3. **Fehlerszenarien**
   - (später ausfüllen)

**Status:**
- Architektur noch offen
- Wird nach Implementierung ausgefüllt

---

## 5️⃣ docs/02_architecture_options.md

### Inhalt

**Sektionen:**
1. **Option 1: systemd-nativ**
2. **Option 2: Zabbix**
3. **Option 3: Prometheus + Loki**
4. **Option 4: Hybrid (gewählt)**
5. **Vergleichsmatrix**
6. **Entscheidung: Hybrid (N4)**

---

## 6️⃣ docs/03_mvp_scope.md

### Inhalt

**Sektionen:**
1. **Zwingend enthalten (MVP)**
2. **Nicht enthalten (v0.1)**
3. **MVP-Ausgabe**
4. **Erfolgskriterien**
5. **Nächste Schritte**

---

## 7️⃣ Technische Dokumentation (docs/04-12)

### Struktur

**Jede Datei:**
- Ziel
- Spezifikation
- Beispiele (anonym)
- Status

**Themen:**
- Prozessmodell
- Modulstruktur
- Interfaces
- Datenmodelle
- Konfigurationsschema
- Validierungsregeln
- systemd Service
- Installer-Konzept
- Teststrategie

---

## 8️⃣ installer/README.md

### Inhalt

**Sektionen:**
1. **Ziel**
   - One-Shot-Installer
   - Offline
   - Reproduzierbar

2. **Voraussetzungen**
   - Root-Rechte
   - systemd
   - journald

3. **Installation**
   - Ablauf
   - Optionen
   - Verifizierung

4. **Deinstallation**
   - Ablauf
   - Optionen

5. **Troubleshooting**
   - Häufige Fehler
   - Lösungen

**Status:**
- Aktuell: nicht implementiert
- Konzept: `docs/11_installer_concept.md`

---

## 9️⃣ monitoring/README.md

### Inhalt

**Sektionen:**
1. **Ziel**
   - Sammelbegriff für Monitoring-Komponenten

2. **Komponenten**
   - Logs
   - Status
   - Cron
   - Timer

3. **Keine Tool-Nennung**
   - Tool-agnostisch
   - Architektur offen

**Status:**
- Aktuell: nur Platzhalter
- Wird später ausgefüllt

---

## 🔟 external-watch/README.md

### Inhalt

**Sektionen:**
1. **Ziel**
   - Externer Wächter (z. B. Raspberry)

2. **Idee**
   - Out-of-Band-Monitoring
   - Unabhängig vom Proxmox

3. **Hardware**
   - (noch offen)

4. **Netzwerk**
   - (noch offen)

**Status:**
- Aktuell: nur Konzept
- Hardware/Netzwerk noch offen

---

## 1️⃣1️⃣ scripts/README.md

### Inhalt

**Sektionen:**
1. **Ziel**
   - Platz für Glue-Skripte

2. **Shell / Python offen**
   - Keine Festlegung
   - Flexibel

3. **Beispiele (später)**
   - Backup-Skripte
   - Wartungs-Skripte
   - Utilities

**Status:**
- Aktuell: leer
- Wird bei Bedarf gefüllt

---

## 1️⃣2️⃣ SECURITY.md

### Inhalt

**Sektionen:**
1. **Grundprinzipien**
   - Offline-First
   - No Telemetry
   - No Cloud
   - No Phone-Home

2. **Verbotene Inhalte im Repository**
3. **Erlaubte Inhalte**
4. **Konfiguration**
5. **Meldung von Sicherheitsproblemen**

**Status:**
- ✅ Bereits erstellt (Schritt B0)

---

## 1️⃣3️⃣ CONTRIBUTING.md

### Inhalt

**Sektionen:**
1. **Datenschutz (Pflicht)**
2. **Konfigurationsdateien**
3. **Code**
4. **Tests / Beispiele**
5. **Verstöße**

**Status:**
- ✅ Bereits erstellt (Schritt B0)

---

## 1️⃣4️⃣ Dokumentations-Prioritäten

### MVP v0.1 (Pflicht)

1. ✅ Root-README.md (Grundstruktur)
2. ✅ docs/00_vision.md
3. ✅ docs/03_mvp_scope.md
4. ✅ installer/README.md (Grundstruktur)
5. ✅ SECURITY.md
6. ✅ CONTRIBUTING.md

### Nach MVP (optional)

1. docs/01_architecture.md (ausfüllen)
2. API-Dokumentation
3. Pattern-Registry-Dokumentation
4. Troubleshooting-Guide
5. FAQ

---

## 1️⃣5️⃣ Dokumentations-Standards

### Format

- Markdown (.md)
- Klare Struktur
- Code-Beispiele (anonym)
- Keine realen Daten

### Sprache

- Deutsch (Hauptdokumentation)
- Englisch (optional, später)

### Links

- Relative Links innerhalb Repo
- Keine externen Abhängigkeiten

---

## 1️⃣6️⃣ README.md Template (Struktur)

```markdown
# prox_watch_monitoring_system

## Status
WIP - Architektur offen

## Was ist es / was nicht
...

## Schnellstart
...

## Datenschutz-Prinzipien
...

## Architektur
...

## Dokumentation
Siehe [docs/](docs/)

## Contributing
Siehe [CONTRIBUTING.md](CONTRIBUTING.md)

## Security
Siehe [SECURITY.md](SECURITY.md)

## Roadmap
...
```

---

## Status

- Dokumentationsstruktur definiert
- Inhalte später
- Übersichtlich
- Erweiterbar
