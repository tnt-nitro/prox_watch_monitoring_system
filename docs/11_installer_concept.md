# Installer-Konzept (One-Shot, offline)

## Ziel
Einmalige Installation. Offline. Deterministisch.

---

## 1️⃣ Installer-Phasen

### Phase 1: Voraussetzungen prüfen

**Prüfungen:**
- Betriebssystem (Linux, systemd)
- Root-Rechte vorhanden
- systemd verfügbar
- journald verfügbar
- Go-Binary vorhanden (oder Build)
- Keine Konflikte (bestehende Installation)

**Fehlerbehandlung:**
- Fehlende Voraussetzungen → Exit 1
- Konflikte → Warnung + Abfrage

---

### Phase 2: User & Verzeichnisse

**Aktionen:**
1. User `proxwatch` erstellen (falls nicht vorhanden)
   - `useradd -r -s /usr/sbin/nologin proxwatch`
2. Group `proxwatch` erstellen (falls nicht vorhanden)
   - `groupadd -r proxwatch`
3. Verzeichnisse erstellen:
   - `/var/lib/prox-watch` (State, Config)
   - `/usr/local/bin` (Binary)
   - `/etc/systemd/system` (Service-Unit)
4. Rechte setzen:
   - `/var/lib/prox-watch`: `700` (proxwatch:proxwatch)
   - `/usr/local/bin`: Standard

**Fehlerbehandlung:**
- Fehlende Rechte → Exit 1
- Verzeichnis-Erstellung fehlgeschlagen → Exit 1

---

### Phase 3: Binary installieren

**Aktionen:**
1. Binary kopieren:
   - Quelle: `./prox-watch` (oder Build)
   - Ziel: `/usr/local/bin/prox-watch`
2. Rechte setzen:
   - `755` (ausführbar)
   - Owner: `root:root`
3. Verifizierung:
   - Binary ausführbar?
   - Version prüfen?

**Fehlerbehandlung:**
- Kopieren fehlgeschlagen → Exit 1
- Binary nicht ausführbar → Exit 1

---

### Phase 4: systemd Service installieren

**Aktionen:**
1. Service-Unit kopieren:
   - Quelle: `./installer/prox-watch.service`
   - Ziel: `/etc/systemd/system/prox-watch.service`
2. systemd neu laden:
   - `systemctl daemon-reload`
3. Service aktivieren:
   - `systemctl enable prox-watch.service`
4. Service nicht starten (manuell)

**Fehlerbehandlung:**
- Kopieren fehlgeschlagen → Exit 1
- systemd-Fehler → Exit 1

---

### Phase 5: Konfiguration initialisieren

**Aktionen:**
1. CLI-Wizard starten:
   - `prox-watch init --config-path /var/lib/prox-watch`
2. Wizard fragt:
   - Zeitzone
   - Proxmox Host (Platzhalter)
   - API-Port
   - Alert-Kanal
   - Config-Pfad bestätigen
3. Validierung:
   - Keine IPs
   - Keine Pfade im Repo
   - Schreibrechte
4. Dateien erstellen:
   - `/var/lib/prox-watch/config.yaml`
   - `/var/lib/prox-watch/secrets.yaml` (leer)
5. Rechte setzen:
   - `600` (proxwatch:proxwatch)

**Fehlerbehandlung:**
- Wizard-Abbruch → Exit 1
- Validierungsfehler → Exit 1
- Schreibfehler → Exit 1

---

### Phase 6: Pattern-Registry installieren

**Aktionen:**
1. Pattern-Registry kopieren:
   - Quelle: `./config/patterns.yaml.example`
   - Ziel: `/var/lib/prox-watch/patterns.yaml` (lokal)
2. Lokale Regex-Zuordnung:
   - Benutzer konfiguriert lokal (nicht versioniert)
3. Rechte setzen:
   - `600` (proxwatch:proxwatch)

**Fehlerbehandlung:**
- Kopieren fehlgeschlagen → Warnung (optional)
- Pattern-Registry optional (kann später nachgeholt werden)

---

### Phase 7: State-Datenbank initialisieren

**Aktionen:**
1. SQLite-Datenbank erstellen:
   - Pfad: `/var/lib/prox-watch/state.db`
2. Schema erstellen:
   - Tabellen: `events`, `cooldowns`, `acknowledges`
   - Indizes
3. Rechte setzen:
   - `600` (proxwatch:proxwatch)

**Fehlerbehandlung:**
- Datenbank-Erstellung fehlgeschlagen → Exit 1
- Schema-Fehler → Exit 1

---

### Phase 8: Verifizierung

**Prüfungen:**
1. Binary vorhanden und ausführbar
2. Service-Unit vorhanden
3. Config-Datei vorhanden
4. State-DB vorhanden
5. Rechte korrekt
6. User vorhanden
7. systemd-Status OK

**Ausgabe:**
- Erfolgreich → Zusammenfassung
- Fehler → Fehlerliste

---

## 2️⃣ Installer-Aufruf

### Kommando

```bash
./install.sh [--skip-wizard] [--config-path PATH]
```

**Optionen:**
- `--skip-wizard`: Konfiguration überspringen (nur für Tests)
- `--config-path`: Alternativer Config-Pfad

**Voraussetzungen:**
- Root-Rechte
- Offline (keine Netzwerk-Zugriffe)

---

## 3️⃣ Rollback-Strategie

### Bei Fehler

**Aktionen:**
1. Installierte Dateien entfernen:
   - Binary
   - Service-Unit
   - Config (wenn erstellt)
   - State-DB (wenn erstellt)
2. User/Group entfernen (optional, nur wenn erstellt)
3. Verzeichnisse entfernen (optional)
4. systemd neu laden

**Fehlerbehandlung:**
- Rollback kann fehlschlagen → Manuelle Bereinigung nötig

---

## 4️⃣ Update-Strategie

### Bestehende Installation

**Prüfungen:**
1. Bestehende Installation erkannt?
2. Config vorhanden?
3. State-DB vorhanden?

**Verhalten:**
- Config behalten (nicht überschreiben)
- State-DB behalten (nicht löschen)
- Binary aktualisieren
- Service-Unit aktualisieren
- systemd neu laden

**Fehlerbehandlung:**
- Backup vor Update (optional)
- Rollback möglich

---

## 5️⃣ Deinstallation

### Kommando

```bash
./uninstall.sh [--keep-config] [--keep-state]
```

**Optionen:**
- `--keep-config`: Config behalten
- `--keep-state`: State-DB behalten

**Aktionen:**
1. Service stoppen
2. Service deaktivieren
3. Service-Unit entfernen
4. Binary entfernen
5. Config entfernen (wenn nicht --keep-config)
6. State-DB entfernen (wenn nicht --keep-state)
7. User/Group entfernen (optional)
8. Verzeichnisse entfernen (optional)
9. systemd neu laden

---

## 6️⃣ Offline-Anforderungen

### Keine Netzwerk-Zugriffe

**Während Installation:**
- Keine Downloads
- Keine API-Calls
- Keine Telemetrie
- Keine Cloud-Services

**Alles lokal:**
- Binary lokal
- Config lokal
- Pattern-Registry lokal
- Service-Unit lokal

---

## 7️⃣ Installer-Struktur

### Verzeichnis

```
installer/
├─ install.sh          # Haupt-Installer
├─ uninstall.sh        # Deinstaller
├─ prox-watch.service  # systemd Unit
└─ README.md           # Installations-Anleitung
```

---

## 8️⃣ Ablauf-Diagramm

```
Start
  ↓
Voraussetzungen prüfen
  ↓ (Fehler → Exit 1)
User & Verzeichnisse
  ↓ (Fehler → Exit 1)
Binary installieren
  ↓ (Fehler → Exit 1)
Service installieren
  ↓ (Fehler → Exit 1)
Konfiguration (Wizard)
  ↓ (Fehler → Exit 1)
Pattern-Registry
  ↓ (Warnung, optional)
State-DB initialisieren
  ↓ (Fehler → Exit 1)
Verifizierung
  ↓
Erfolg / Fehler
```

---

## 9️⃣ Fehlerbehandlung

### Kritische Fehler

- Fehlende Root-Rechte → Exit 1
- Fehlende Voraussetzungen → Exit 1
- Binary nicht ausführbar → Exit 1
- Service-Installation fehlgeschlagen → Exit 1
- Config-Validierung fehlgeschlagen → Exit 1
- State-DB-Erstellung fehlgeschlagen → Exit 1

### Warnungen

- Bestehende Installation → Warnung + Abfrage
- Pattern-Registry fehlt → Warnung (optional)
- User bereits vorhanden → Info

---

## 🔟 Installations-Log

### Ausgabe

```
[INFO] Checking prerequisites...
[OK]   systemd available
[OK]   journald available
[INFO] Creating user and directories...
[OK]   User 'proxwatch' created
[OK]   Directory '/var/lib/prox-watch' created
[INFO] Installing binary...
[OK]   Binary installed to '/usr/local/bin/prox-watch'
[INFO] Installing systemd service...
[OK]   Service installed
[INFO] Initializing configuration...
[OK]   Configuration created
[INFO] Installation complete!
```

---

## 1️⃣1️⃣ Post-Installation

### Manuelle Schritte

1. Service starten:
   ```bash
   systemctl start prox-watch.service
   ```

2. Status prüfen:
   ```bash
   systemctl status prox-watch.service
   ```

3. Logs prüfen:
   ```bash
   journalctl -u prox-watch.service -f
   ```

---

## Status

- Installer-Konzept definiert
- Ablauf spezifiziert
- Offline-First
- Rollback möglich
