# Release Notes v0.5.0

**v0.5.0 - Gesicherter Power-Cycle (Phase 3)**

## Status

✅ **Release-fähig** - Gesicherter Power-Cycle (Phase 3) implementiert und getestet

## Features (neu in v0.5.0)

### ✅ GPIO Power-Cycle (Relais)

- **Hardware-Steuerung:** Vollständige GPIO-Integration für Relais-basierten Power-Cycle
- **NO/NC-kompatibel:** Unterstützung für Normally Open (NO) und Normally Closed (NC) Relais über `relay_mode`
- **Schaltsequenz:** Power OFF → Warte (min_downtime) → Power ON
- **Ein Toggle pro Attempt:** Garantiert, dass nur ein Power-Cycle-Versuch pro Eskalation durchgeführt wird
- **Deterministisch:** Pin-Sequenz ist vollständig deterministisch und testbar

### ✅ ARM-Mechanismus (Manuelle Freigabe)

- **ARM-Datei:** Power-Cycle erfordert explizite manuelle Freigabe über ARM-Datei (`/var/lib/prox-watch/arm_powercycle`)
- **Automatische Entfernung:** ARM-Datei wird nach erfolgreichem Attempt automatisch entfernt
- **Fehlerbehandlung:** Bei Fehler bleibt ARM-Datei erhalten (für Diagnose)
- **Sicherheit:** Kein Power-Cycle ohne ARM-Datei, auch wenn alle anderen Bedingungen erfüllt sind

### ✅ Severity-Edge-Trigger

- **Edge-Erkennung:** Power-Cycle wird nur ausgelöst, wenn Severity von **< CRIT → CRIT** wechselt
- **CRIT-Sturm-Schutz:** Verhindert Attempts bei dauerhaftem CRIT (kein Neustart-Loop)
- **Restart-Sicherheit:** Nach Restart wird kein Attempt durchgeführt, wenn bereits CRIT vorliegt (ohne Edge)
- **Neue Eskalation:** CRIT → WARN → CRIT erlaubt neuen Attempt (wenn ARM + Retry ok)

### ✅ Retry-Window & Max-Attempts

- **Retry-Cooldown:** Konfigurierbarer Cooldown zwischen Versuchen (`retry_after_seconds`, Default: 900 = 15 Minuten)
- **Max Attempts:** Konfigurierbare maximale Anzahl Versuche (`max_attempts`, Default: 1)
- **Persistenz:** Retry-Status und Attempt-Count werden persistent gespeichert
- **Restart-sicher:** Limits bleiben nach Restart erhalten

### ✅ Erweiterte Persistenz

- **PowerAttempts:** Anzahl durchgeführter Power-Cycle-Versuche wird persistent gespeichert
- **LastPowerAttempt:** Timestamp des letzten Versuchs wird gespeichert
- **Schema-Migration:** Automatische Migration bestehender Datenbanken (kein Datenverlust)
- **Single-Row-Tabelle:** Weiterhin nur eine Tabelle mit genau einer Zeile (minimaler State)

### ✅ Sicherheitsdokumentation

- **Vollständige Dokumentation:** [docs/24_powercycle_safety.md](docs/24_powercycle_safety.md)
- **Risikoabschnitt:** Explizite Warnungen vor Datenverlust und Hardware-Schäden
- **Wiring-Hinweise:** Detaillierte Anleitung für NO/NC-Relais-Verdrahtung
- **ARM-Prozess:** Schritt-für-Schritt-Anleitung für Aktivierung/Deaktivierung
- **Kill-Switch:** Schnellste Deaktivierung über `powercycle.enabled: false`
- **Fehlerbehandlung:** Diagnose-Anleitung für häufige Fehler

## Sicherheitsgarantien (v0.5.0)

### Harte Grenzen (Power-Cycle)

- 🔒 **Kein Relais-Sturm:** Severity-Edge-Trigger verhindert Attempts bei dauerhaftem CRIT
- 🔒 **Kein Toggle-Loop:** Nur ein Toggle pro Attempt, kein automatisches Zurücksetzen
- 🔒 **Kein Attempt ohne ARM:** ARM-Datei ist zwingend erforderlich
- 🔒 **Kein Attempt bei dauerhaftem CRIT:** Nur Edge-Trigger (INFO/WARN → CRIT)
- 🔒 **Kein Attempt nach Restart ohne Edge:** Restart mit gespeichertem CRIT führt nicht zu Attempt
- 🔒 **Kein Panic bei Fehlern:** Alle Fehler werden abgefangen, kein System-Crash
- 🔒 **Max Attempts Limit:** Konfigurierbares Limit verhindert unbegrenzte Versuche
- 🔒 **Retry-Cooldown:** Zeitbasierter Schutz gegen wiederholte Versuche
- 🔒 **Min Downtime:** Garantiert ausreichend Zeit für sauberes Herunterfahren

### Konfigurations-Guards

- 🔒 **relay_mode Pflicht:** Power-Cycle erfordert explizite `relay_mode`-Konfiguration (NO/NC)
- 🔒 **Validierung:** Alle Konfigurationsparameter werden streng validiert
- 🔒 **Default deaktiviert:** Power-Cycle ist standardmäßig deaktiviert (`enabled: false`)

## Nicht enthalten in v0.5.0

- ❌ **Automatische Recovery:** Power-Cycle ist kein Selbstheilungsmechanismus
- ❌ **Remote-Steuerung:** Keine Fernsteuerung des Power-Cycle (nur lokal)
- ❌ **GUI/Web-UI:** Keine grafische Oberfläche für Power-Cycle-Konfiguration

## Datenschutz

- 🔒 **Offline-First:** Keine Cloud-Abhängigkeiten für Power-Cycle
- 🔒 **No Telemetry:** Keine Nutzungsdaten werden gesammelt oder versendet
- 🔒 **Privacy-by-Design:** Konfigurations-Guards verhindern Datenlecks
- 🔒 **Lokale Speicherung:** Power-Cycle-State wird ausschließlich lokal gespeichert

## Installation & Build

- **Standard-Build (ohne Hardware-GPIO):**
  ```bash
  go build -o prox-watch-watcher ./cmd/watcher
  ```

- **Hardware-Build (für Raspberry Pi mit echter GPIO):**
  ```bash
  go build -tags raspberry -o prox-watch-watcher ./cmd/watcher
  ```

Eine detaillierte Installationsanleitung finden Sie unter [cmd/watcher/README.md](cmd/watcher/README.md) und [docs/21_watcher_deployment.md](docs/21_watcher_deployment.md).

**⚠️ WICHTIG:** Lesen Sie [docs/24_powercycle_safety.md](docs/24_powercycle_safety.md) **VOR** der Aktivierung von Power-Cycle!

## Breaking Changes

Keine.

## Bekannte Einschränkungen

1. **Hardware-Abhängigkeit:** Power-Cycle erfordert physische Relais-Hardware
2. **Verdrahtung:** `relay_mode` muss korrekt konfiguriert werden (NO vs. NC)
3. **Test-Umgebung:** Power-Cycle sollte in einer sicheren Test-Umgebung getestet werden

## Migration von v0.4.0

1. **Datenbank-Migration:** Automatisch beim Start (keine manuelle Aktion erforderlich)
2. **Konfiguration:** Neue `powercycle`-Sektion hinzufügen (optional, standardmäßig deaktiviert)
3. **ARM-Datei:** Erstellen Sie die ARM-Datei nur, wenn Sie Power-Cycle aktivieren möchten

## Nächste Schritte (Phase 4)

- **Phase 4 – Komfort & Erweiterung:** GUI, Profile, Visualisierung (optional)
- **Weitere Hardware-Integrationen:** Zusätzliche Sensoren, erweiterte GPIO-Funktionen

## Evolutionslinie

- **v0.1.0** → Core (Event-Processing, State-Management, Push)
- **v0.2.0** → Watcher (Health-Check, Counter, Severity, GPIO-Interface)
- **v0.3.0** → GPIO (Hardware-LED/Beeper, periph.io Integration)
- **v0.4.0** → Persistenz & Cooldown (SQLite-State, Push-Spam-Schutz)
- **v0.5.0** → Power-Cycle (Safety-Grade, Edge-Trigger, ARM-Mechanismus)

**Architektur ist jetzt vollständig autonom offline-fähig.** 🚀
