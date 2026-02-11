# Release Notes v0.4.0

**v0.4.0 - Persistenz & Cooldown (Phase 2)**

## Status

✅ **Release-fähig** - Phase 2 (Persistenz & Cooldown) implementiert und getestet

## Features (neu in v0.4.0)

### ✅ Persistenter State (SQLite)

- **Single-Row-Tabelle:** Der Watcher-State wird in einer SQLite-Datenbank (`watcher_state.db`) gespeichert.
- **Minimaler State:** Nur notwendige Daten werden persistent gespeichert:
  - `FailCount`: Aktueller Fehlzähler
  - `CurrentSeverity`: Aktuelle Severity (INFO/WARN/CRIT)
  - `LastEscalation`: Timestamp der letzten Eskalation (für Cooldown)
- **Restart-sicher:** State wird beim Start geladen und bei Änderungen gespeichert.
- **Keine sensiblen Daten:** Keine Host/IP-Adressen oder andere sensible Informationen im State.

### ✅ Cooldown-Mechanismus

- **Konfigurierbar:** Cooldown-Dauer über `watcher.cooldown_seconds` konfigurierbar (Default: 600 Sekunden = 10 Minuten).
- **Push-Spam-Schutz:** Verhindert wiederholte Push-Benachrichtigungen bei gleicher Severity innerhalb des Cooldown-Zeitfensters.
- **Nur bei Eskalation:** Cooldown wird nur bei tatsächlichen Eskalationen (INFO→WARN, WARN→CRIT) aktiviert.
- **Restart-sicher:** Cooldown-Status wird über `LastEscalation` im persistenten State gespeichert.

### ✅ Konfiguration erweitert

- **`watcher.cooldown_seconds`:** Neue Konfigurationsoption für Cooldown-Dauer.
  - Default: 600 Sekunden (10 Minuten)
  - Min: 0 (Cooldown deaktiviert)
  - Max: 86400 Sekunden (24 Stunden)
  - Validierung: Negative Werte und Werte > 86400 werden abgelehnt.

### ✅ Integration Persistenz + Cooldown

- **Nahtlose Integration:** Persistenz und Cooldown sind vollständig in den Runner integriert.
- **Save nur bei Änderungen:** State wird nur bei tatsächlichen Änderungen gespeichert (kein Save-Spam).
- **Fehlerresistent:** DB-Fehler werden ignoriert (kein Panic), System läuft mit Default-State weiter.

### ✅ Erweiterte Integrationstests

- **Restart-Szenarien:** Tests für Neustart mit aktivem/abgelaufenem Cooldown.
- **Stabile Zustände:** Tests für stabile INFO (kein Save-Spam).
- **Flatternde Zustände:** Tests für flatternde Zustände mit Cooldown-Schutz.
- **Fehlerbehandlung:** Tests für DB-Fehler (kein Panic).

## Sicherheitsgarantien (v0.4.0)

### Harte Grenzen (Watcher)

- ❌ **Kein Auto-Reboot:** Keine automatischen System-Neustarts.
- ❌ **Kein Power-Cycle:** Automatische Hardware-Aktionen sind weiterhin nicht implementiert und bewusst gesperrt.
- ❌ **Kein Fernzugriff:** Das System bietet keine Remote-Steuerung oder -Konfiguration.
- ❌ **Kein Host/IP im State:** Der persistente State enthält keine Host- oder IP-Informationen.
- ❌ **Kein Push-Spam:** Cooldown-Mechanismus verhindert wiederholte Push-Benachrichtigungen.
- ❌ **Kein Doppel-Push nach Restart:** Cooldown-Status wird persistent gespeichert, verhindert doppelte Pushes nach Neustart.
- ❌ **Kein Save-Spam:** State wird nur bei tatsächlichen Änderungen gespeichert (nicht bei jedem Intervall).
- ❌ **Kein Panic bei DB-Fehler:** DB-Fehler werden ignoriert, System läuft mit Default-State weiter.

## Nicht enthalten in v0.4.0

- ❌ **Power-Cycle-Funktionalität:** Das Design ist vorhanden, aber die Implementierung ist bewusst gesperrt und erfordert manuelle Freigabe (geplant für Phase 3).
- ❌ **Systemd-Service für Watcher:** Die systemd Service-Unit für den Watcher ist noch nicht implementiert (optional für später).

## Datenschutz

- 🔒 **Offline-First:** Keine Cloud-Abhängigkeiten für den Watcher-Betrieb.
- 🔒 **No Telemetry:** Keine Nutzungsdaten werden gesammelt oder versendet.
- 🔒 **Privacy-by-Design:** Konfigurations-Guards verhindern Datenlecks (z.B. IP-Literale im Repo).
- 🔒 **Metadaten-Only:** Push-Nachrichten enthalten keine sensiblen Inhalte.
- 🔒 **Minimaler State:** Der persistente State enthält nur notwendige Metadaten, keine Host/IP-Informationen.

## Installation & Build

- **Standard-Build (mit Mock-GPIO):**
  ```bash
  go build -o prox-watch-watcher ./cmd/watcher
  ```
- **Hardware-Build (für Raspberry Pi mit echter GPIO):**
  ```bash
  go build -tags raspberry -o prox-watch-watcher ./cmd/watcher
  ```
Eine detaillierte Installationsanleitung finden Sie unter [cmd/watcher/README.md](cmd/watcher/README.md) und [docs/21_watcher_deployment.md](docs/21_watcher_deployment.md).

## Breaking Changes

Keine.

## Bekannte Einschränkungen

1. **Systemd-Service fehlt:** Die systemd Service-Unit für den Watcher ist noch nicht implementiert (optional für später).
2. **Power-Cycle nicht implementiert:** Das Design ist vorhanden, aber die Implementierung ist bewusst gesperrt (geplant für Phase 3).

## Migration von v0.3.0

- **Neue Konfigurationsoption:** `watcher.cooldown_seconds` wurde hinzugefügt (Default: 600).
- **Persistenter State:** Der Watcher erstellt automatisch eine SQLite-Datenbank (`watcher_state.db`) beim ersten Start.
- **Keine manuelle Migration erforderlich:** Der Watcher erstellt die Datenbank automatisch, wenn sie nicht existiert.

## Nächste Schritte (Phase 3)

- **Phase 3 – Gesicherter Power-Cycle (Watcher):** Design und Implementierung einer kontrollierten Power-Cycle-Funktionalität mit manueller Freigabe.
- **Systemd-Service für Watcher:** Optional: systemd Service-Unit für den Watcher ergänzen.
