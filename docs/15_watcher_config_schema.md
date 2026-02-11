# Watcher-Konfigurationsschema (Phase 1)

## Status

**Phase 1 - Spezifikation** 🟡

## Ziel des Watchers

- ✅ Erreichbarkeit prüfen
- ✅ Fehlzähler führen
- ✅ Severity bestimmen
- ✅ Push senden
- ✅ GPIO steuern

## Nicht enthalten

- ❌ Logs lesen
- ❌ Daten speichern außer minimalem State
- ❌ Automatische Reparatur
- ❌ Power-Cycle (nur Design, keine Umsetzung in Phase 1)

## Konfigurationsschema

### Top-Level-Struktur

```yaml
watcher:
  interval_seconds: 30
  cooldown_seconds: 600  # Phase 2: Cooldown in Sekunden (Default: 600 = 10 Minuten)

target:
  mode: "ping+https"        # ping | https | ping+https
  host: "PLACEHOLDER"
  port: 8006
  timeout_seconds: 5

thresholds:
  warn: 3
  crit: 10

push:
  enabled: true
  adapter: "ntfy"
  topics:
    warn: "prox-watch-warn"
    crit: "prox-watch-crit"

gpio:
  enabled: false
  led_pin: 17
  beeper_pin: 27
  beeper_day_only: true

security:
  block_ip_literals: true
  require_manual_powercycle: true
```

### Detaillierte Beschreibung

#### `watcher.interval_seconds`

- **Typ:** Integer
- **Default:** 30
- **Beschreibung:** Intervall zwischen Health-Checks in Sekunden
- **Min:** 10
- **Max:** 300

#### `watcher.cooldown_seconds`

- **Typ:** Integer
- **Default:** 600 (10 Minuten)
- **Beschreibung:** Cooldown-Dauer für Push-Benachrichtigungen in Sekunden. Verhindert Push-Spam bei wiederholten Eskalationen.
- **Min:** 0 (0 = Cooldown deaktiviert)
- **Max:** 86400 (24 Stunden)
- **Phase 2:** Neu hinzugefügt

#### `target.mode`

- **Typ:** String (Enum)
- **Optionen:**
  - `ping` - Nur Ping-Check
  - `https` - Nur HTTPS-Check (Port 8006)
  - `ping+https` - Beide Checks (beide müssen erfolgreich sein)
- **Default:** `ping+https`

#### `target.host`

- **Typ:** String
- **Beschreibung:** Hostname oder IP-Adresse (wird validiert)
- **Validierung:** Keine IP-Literale im Repo (nur Platzhalter)
- **Beispiel (Repo):** `"PLACEHOLDER"`
- **Beispiel (lokal):** `"proxmox.local"` oder `"192.168.1.100"`

#### `target.port`

- **Typ:** Integer
- **Default:** 8006
- **Beschreibung:** Port für HTTPS-Check
- **Min:** 1
- **Max:** 65535

#### `target.timeout_seconds`

- **Typ:** Integer
- **Default:** 5
- **Beschreibung:** Timeout für Health-Checks in Sekunden
- **Min:** 1
- **Max:** 30

#### `thresholds.warn`

- **Typ:** Integer
- **Default:** 3
- **Beschreibung:** Anzahl Fehlversuche für WARN-Severity
- **Min:** 1

#### `thresholds.crit`

- **Typ:** Integer
- **Default:** 10
- **Beschreibung:** Anzahl Fehlversuche für CRIT-Severity
- **Min:** 1

#### `push.enabled`

- **Typ:** Boolean
- **Default:** true
- **Beschreibung:** Push-Benachrichtigungen aktivieren/deaktivieren

#### `push.adapter`

- **Typ:** String
- **Default:** `"ntfy"`
- **Beschreibung:** Push-Adapter (aktuell nur `ntfy`)

#### `push.topics.warn`

- **Typ:** String
- **Default:** `"prox-watch-warn"`
- **Beschreibung:** ntfy-Topic für WARN-Events

#### `push.topics.crit`

- **Typ:** String
- **Default:** `"prox-watch-crit"`
- **Beschreibung:** ntfy-Topic für CRIT-Events

#### `gpio.enabled`

- **Typ:** Boolean
- **Default:** false
- **Beschreibung:** GPIO-Steuerung aktivieren/deaktivieren

#### `gpio.led_pin`

- **Typ:** Integer
- **Default:** 17
- **Beschreibung:** GPIO-Pin für LED (BCM-Nummerierung)
- **Hinweis:** Nur relevant, wenn `gpio.enabled: true`

#### `gpio.beeper_pin`

- **Typ:** Integer
- **Default:** 27
- **Beschreibung:** GPIO-Pin für Beeper (BCM-Nummerierung)
- **Hinweis:** Nur relevant, wenn `gpio.enabled: true`

#### `gpio.beeper_day_only`

- **Typ:** Boolean
- **Default:** true
- **Beschreibung:** Beeper nur tagsüber aktivieren
- **Hinweis:** Nacht = 22:00 - 06:00 (konfigurierbar, später)

#### `security.block_ip_literals`

- **Typ:** Boolean
- **Default:** true
- **Beschreibung:** Blockiere IP-Literale in Konfiguration (Datenschutz-Guard)

#### `security.require_manual_powercycle`

- **Typ:** Boolean
- **Default:** true
- **Beschreibung:** Erfordert manuelle Freigabe für Power-Cycle (Phase 1: nicht implementiert, nur Design)

## Datenschutzregeln (Watcher)

### Verboten

- ❌ **Keine Speicherung von Host/IP im Log**
- ❌ **Keine Persistenz außer Counter (optional)**
- ❌ **Kein Rückkanal**
- ❌ **Kein Auto-Reboot**
- ❌ **Keine IP-Literale im Repository**

### Erlaubt

- ✔ **Minimaler State** (Counter, FirstSeen, LastSeen)
- ✔ **Push-Metadaten** (EventID, Severity, Timestamp)
- ✔ **GPIO-Steuerung** (lokal, keine Datenübertragung)

## Beispiel-Konfiguration (Repo-safe)

```yaml
# config/watcher.yaml.example

watcher:
  interval_seconds: 30
  cooldown_seconds: 600  # Phase 2: Cooldown in Sekunden (Default: 600 = 10 Minuten)

target:
  mode: "ping+https"
  host: "PLACEHOLDER"  # Lokal ersetzen
  port: 8006
  timeout_seconds: 5

thresholds:
  warn: 3
  crit: 10

push:
  enabled: true
  adapter: "ntfy"
  topics:
    warn: "prox-watch-warn"
    crit: "prox-watch-crit"

gpio:
  enabled: false
  led_pin: 17
  beeper_pin: 27
  beeper_day_only: true

security:
  block_ip_literals: true
  require_manual_powercycle: true
```

## Validierungsregeln

### Start-Blockaden

1. **IP-Literal-Erkennung:**
   - Wenn `security.block_ip_literals: true` und IP-Literal in `target.host` → **STOP**

2. **Pfad-Validierung:**
   - Config-Pfad nicht im Repo → **STOP**
   - Keine Schreibrechte → **STOP**

3. **GPIO-Validierung:**
   - Wenn `gpio.enabled: true` und GPIO nicht verfügbar → **WARN** (nicht STOP)

4. **Threshold-Validierung:**
   - `thresholds.warn >= thresholds.crit` → **STOP**

5. **Cooldown-Validierung (Phase 2):**
   - `watcher.cooldown_seconds < 0` → **STOP**
   - `watcher.cooldown_seconds > 86400` → **STOP**

## Integration mit Core-Config

Der Watcher nutzt **separate Konfiguration** (`watcher.yaml`), teilt aber:
- Push-Adapter-Konfiguration (optional, kann überschrieben werden)
- Security-Guards (gleiche Validierungslogik)

## Nächste Schritte

1. Health-Check-Engine Design
2. Counter-Implementierung
3. Severity-Flow
4. GPIO-Abstraktion
5. Runner-Implementierung
