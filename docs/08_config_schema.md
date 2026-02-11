# Konfigurationsschema (Keys & Defaults)

## Ziel
Struktur klar. Defaults fest. Validierung möglich.

---

## 1️⃣ Konfigurationsstruktur

### YAML-Schema

```yaml
system:
  timezone: <string>

proxmox:
  host: <string>
  api_port: <integer>

alerts:
  channel: <string>
  ntfy:
    server: <string>
    topics:
      info: <string>
      warn: <string>
      crit: <string>

paths:
  state_db: <string>
  config: <string>
  secrets: <string>

rules:
  thresholds:
    warn: <integer>
    crit: <integer>
  windows:
    warn: <duration>
    crit: <duration>
  cooldown: <duration>
```

---

## 2️⃣ System-Konfiguration

### `system`

**Keys:**
- `timezone`: Zeitzone (IANA-Format)

**Defaults:**
- `timezone`: `"UTC"`

**Validierung:**
- Muss gültige IANA-Zeitzone sein
- Keine IPs
- Keine Pfade

**Beispiel (anonym):**
```yaml
system:
  timezone: "Europe/Berlin"
```

---

## 3️⃣ Proxmox-Konfiguration

### `proxmox`

**Keys:**
- `host`: Hostname (Platzhalter)
- `api_port`: API-Port

**Defaults:**
- `host`: `"PROXMOX_HOSTNAME"`
- `api_port`: `8006`

**Validierung:**
- `host`: Keine IP-Adressen
- `host`: Keine realen Hostnames erzwingen
- `api_port`: 1-65535

**Beispiel (anonym):**
```yaml
proxmox:
  host: "PROXMOX_HOSTNAME"
  api_port: 8006
```

---

## 4️⃣ Alerts-Konfiguration

### `alerts`

**Keys:**
- `channel`: Alert-Kanal
- `ntfy.server`: ntfy-Server-URL (optional)
- `ntfy.topics.info`: Topic für INFO
- `ntfy.topics.warn`: Topic für WARN
- `ntfy.topics.crit`: Topic für CRIT

**Defaults:**
- `channel`: `"local-only"`
- `ntfy.server`: `""` (leer, wenn nicht gesetzt)
- `ntfy.topics.info`: `"prox-watch-info"`
- `ntfy.topics.warn`: `"prox-watch-warn"`
- `ntfy.topics.crit`: `"prox-watch-crit"`

**Validierung:**
- `channel`: `"local-only"` | `"ntfy"`
- `ntfy.server`: URL-Format (wenn gesetzt)
- `ntfy.topics.*`: Keine Leerzeichen, keine Sonderzeichen

**Beispiel (anonym):**
```yaml
alerts:
  channel: "ntfy"
  ntfy:
    server: "https://ntfy.sh"
    topics:
      info: "prox-watch-info"
      warn: "prox-watch-warn"
      crit: "prox-watch-crit"
```

---

## 5️⃣ Paths-Konfiguration

### `paths`

**Keys:**
- `state_db`: Pfad zur State-Datenbank
- `config`: Pfad zur Konfiguration
- `secrets`: Pfad zu Secrets

**Defaults:**
- `state_db`: `"/var/lib/prox_watch/state.db"`
- `config`: `"/var/lib/prox_watch/config.yaml"`
- `secrets`: `"/var/lib/prox_watch/secrets.yaml"`

**Validierung:**
- Keine Pfade im Repo
- Schreibrechte erforderlich
- Absolute Pfade

**Beispiel (anonym):**
```yaml
paths:
  state_db: "/var/lib/prox_watch/state.db"
  config: "/var/lib/prox_watch/config.yaml"
  secrets: "/var/lib/prox_watch/secrets.yaml"
```

---

## 6️⃣ Rules-Konfiguration

### `rules`

**Keys:**
- `thresholds.warn`: Schwellwert für WARN
- `thresholds.crit`: Schwellwert für CRIT
- `windows.warn`: Zeitfenster für WARN
- `windows.crit`: Zeitfenster für CRIT
- `cooldown`: Cooldown-Dauer

**Defaults:**
- `thresholds.warn`: `3`
- `thresholds.crit`: `10`
- `windows.warn`: `"10m"`
- `windows.crit`: `"15m"`
- `cooldown`: `"30m"`

**Validierung:**
- `thresholds.*`: > 0
- `windows.*`: Gültige Duration
- `cooldown`: Gültige Duration

**Beispiel:**
```yaml
rules:
  thresholds:
    warn: 3
    crit: 10
  windows:
    warn: "10m"
    crit: "15m"
  cooldown: "30m"
```

---

## 7️⃣ Secrets-Schema (separate Datei)

### `secrets.yaml`

**Keys:**
- `proxmox.api_token`: Proxmox API-Token (optional)
- `ntfy.token`: ntfy-Token (optional)

**Defaults:**
- Alle leer (optional)

**Validierung:**
- Nicht im Repo
- Rechte: 600
- Keine Defaults

**Beispiel (anonym):**
```yaml
proxmox:
  api_token: ""

ntfy:
  token: ""
```

---

## 8️⃣ Vollständiges Beispiel (anonym)

### `config.yaml.example`

```yaml
system:
  timezone: "Europe/Berlin"

proxmox:
  host: "PROXMOX_HOSTNAME"
  api_port: 8006

alerts:
  channel: "ntfy"
  ntfy:
    server: "https://ntfy.sh"
    topics:
      info: "prox-watch-info"
      warn: "prox-watch-warn"
      crit: "prox-watch-crit"

paths:
  state_db: "/var/lib/prox_watch/state.db"
  config: "/var/lib/prox_watch/config.yaml"
  secrets: "/var/lib/prox_watch/secrets.yaml"

rules:
  thresholds:
    warn: 3
    crit: 10
  windows:
    warn: "10m"
    crit: "15m"
  cooldown: "30m"
```

---

## 9️⃣ Validierungsregeln (hart)

### Datenschutz-Guards

1. **Keine IPs:**
   - Prüfung auf IPv4/IPv6-Pattern
   - Fehler bei Treffer

2. **Keine Pfade im Repo:**
   - Prüfung, ob Pfad im Repo-Verzeichnis
   - Fehler bei Treffer

3. **Keine realen Hostnames:**
   - Warnung bei Punkten in Hostnames
   - Erlaubt, aber Warnung

4. **Schreibrechte:**
   - Prüfung auf Schreibrechte
   - Fehler bei fehlenden Rechten

---

## 🔟 Konfigurationsladung

### Reihenfolge

1. Defaults setzen
2. Datei laden (wenn vorhanden)
3. CLI-Overrides (wenn vorhanden)
4. Validierung
5. Guards prüfen

### Fehlerbehandlung

- Validierungsfehler → Exit 1
- Guards-Fehler → Exit 1
- Fehlende Datei → Defaults verwenden (warnen)

---

## 1️⃣1️⃣ Type-Definitionen (Go)

```go
type Config struct {
    System  SystemConfig  `yaml:"system"`
    Proxmox ProxmoxConfig `yaml:"proxmox"`
    Alerts  AlertsConfig  `yaml:"alerts"`
    Paths   PathsConfig   `yaml:"paths"`
    Rules   RulesConfig   `yaml:"rules"`
}

type SystemConfig struct {
    Timezone string `yaml:"timezone"`
}

type ProxmoxConfig struct {
    Host    string `yaml:"host"`
    APIPort int    `yaml:"api_port"`
}

type AlertsConfig struct {
    Channel string     `yaml:"channel"`
    Ntfy    NtfyConfig `yaml:"ntfy,omitempty"`
}

type NtfyConfig struct {
    Server string      `yaml:"server"`
    Topics TopicConfig `yaml:"topics"`
}

type TopicConfig struct {
    Info string `yaml:"info"`
    Warn string `yaml:"warn"`
    Crit string `yaml:"crit"`
}

type PathsConfig struct {
    StateDB string `yaml:"state_db"`
    Config  string `yaml:"config"`
    Secrets string `yaml:"secrets"`
}

type RulesConfig struct {
    Thresholds ThresholdConfig `yaml:"thresholds"`
    Windows    WindowConfig    `yaml:"windows"`
    Cooldown   string          `yaml:"cooldown"`
}

type ThresholdConfig struct {
    Warn int `yaml:"warn"`
    Crit int `yaml:"crit"`
}

type WindowConfig struct {
    Warn string `yaml:"warn"`
    Crit string `yaml:"crit"`
}
```

---

## Status

- Konfigurationsschema definiert
- Defaults festgelegt
- Validierung spezifiziert
- Datenschutz-Guards definiert
