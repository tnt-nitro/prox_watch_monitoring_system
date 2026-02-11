# Vision & Scope

## Problemstellung

### Was wird überwacht

- **Proxmox-Server** - Host-System, VMs, Container
- **systemd journald** - System-Logs, Kernel-Logs, Service-Logs
- **Ereignisse** - Netzwerk-Ausfälle, Hardware-Fehler, Service-Crashes

### Warum offline

- **Datenschutz** - keine Cloud-Abhängigkeiten
- **Autonomie** - vollständig lokal betreibbar
- **Zuverlässigkeit** - keine Netzwerk-Ausfälle als Single Point of Failure

### Warum datenschutz-konform

- **Privacy-by-Design** - keine sensiblen Daten im Repository
- **Metadaten-Only** - keine Log-Inhalte, keine IPs, keine Hostnames
- **Lokale Speicherung** - alle Daten bleiben auf dem Server

## Zielsetzung

### Hauptziele

1. **Event-Erkennung** - Pattern-Matching gegen systemd journald
2. **Severity-Bewertung** - INFO / WARN / CRIT basierend auf Zählregeln
3. **Push-Benachrichtigungen** - ntfy-Integration (optional)
4. **State-Persistenz** - SQLite-basierte Event-Historie
5. **Offline-First** - vollständig ohne Cloud-Dienste

### Erfolgskriterien

- ✅ System startet ohne Fehler
- ✅ Konfiguration wird lokal gespeichert
- ✅ Logs werden gelesen
- ✅ Patterns werden getroffen
- ✅ Events werden gezählt
- ✅ Push wird bei CRIT gesendet
- ✅ Keine Datenlecks
- ✅ Vollständig offline betreibbar

## Nicht-Ziele

### Was nicht implementiert wird

- ❌ **Externer Wächter** (Raspberry) - Phase 2
- ❌ **GUI / Web-UI** - nicht im MVP
- ❌ **Langzeit-Historisierung** - nicht im MVP
- ❌ **Automatische Recovery** - explizit ausgeschlossen
- ❌ **Cloud-Integration** - explizit ausgeschlossen

### Was explizit ausgeschlossen ist

- ❌ **Telemetrie** - keine Datenübertragung
- ❌ **Automatische Updates** - manuell
- ❌ **Remote-Zugriff** - lokal nur
- ❌ **Log-Speicherung** - nur Metadaten

## Grundprinzipien

### Offline-First

- Alle Komponenten lokal
- Keine externen Abhängigkeiten (außer ntfy optional)
- Funktioniert ohne Internet

### Datenschutz

- **Privacy-by-Design** - keine sensiblen Daten im Repository
- **Config Guards** - Validierung verhindert Datenlecks
- **Metadaten-Only** - keine Log-Inhalte, keine IPs

### Determinismus

- Reproduzierbare Severity-Bewertung
- Deterministische Event-ID-Generierung
- Fake-Time für Tests
