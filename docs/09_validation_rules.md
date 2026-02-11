# Validierungsregeln (Config Guards)

## Ziel
Harte Stops. Datenschutz erzwingen. Keine Kompromisse.

---

## 1️⃣ Start-Blockaden (hart)

### ❌ Blockiert Start sofort

1. **Config-Pfad im Repo**
   - Prüfung: Ist Config-Pfad innerhalb Repo-Verzeichnis?
   - Fehler: `"config path must be outside repository"`
   - Exit-Code: 1

2. **IP-Adressen in Config**
   - Prüfung: IPv4/IPv6-Pattern in Config
   - Fehler: `"IP addresses not allowed in configuration"`
   - Exit-Code: 1

3. **State-DB-Pfad im Repo**
   - Prüfung: Ist State-DB-Pfad innerhalb Repo-Verzeichnis?
   - Fehler: `"state database path must be outside repository"`
   - Exit-Code: 1

4. **Schreibrechte fehlen**
   - Prüfung: Schreibrechte für State-DB-Pfad
   - Fehler: `"no write permission for state database path"`
   - Exit-Code: 1

5. **Secrets-Pfad im Repo**
   - Prüfung: Ist Secrets-Pfad innerhalb Repo-Verzeichnis?
   - Fehler: `"secrets path must be outside repository"`
   - Exit-Code: 1

6. **Ungültige Zeitzone**
   - Prüfung: IANA-Zeitzone gültig
   - Fehler: `"invalid timezone"`
   - Exit-Code: 1

7. **Ungültiger API-Port**
   - Prüfung: Port 1-65535
   - Fehler: `"invalid API port"`
   - Exit-Code: 1

8. **Ungültige Thresholds**
   - Prüfung: Thresholds > 0
   - Fehler: `"thresholds must be greater than 0"`
   - Exit-Code: 1

9. **Ungültige Zeitfenster**
   - Prüfung: Gültige Duration-Strings
   - Fehler: `"invalid time window"`
   - Exit-Code: 1

10. **Ungültiger Alert-Kanal**
    - Prüfung: `"local-only"` | `"ntfy"`
    - Fehler: `"invalid alert channel"`
    - Exit-Code: 1

---

## 2️⃣ Warnungen (nicht blockierend)

### ⚠️ Warnung, aber Start erlaubt

1. **Hostname mit Punkten**
   - Prüfung: Punkte in Hostname
   - Warnung: `"hostname contains dots, ensure it's a placeholder"`
   - Start: Erlaubt

2. **Config-Datei nicht gefunden**
   - Prüfung: Config-Datei existiert
   - Warnung: `"config file not found, using defaults"`
   - Start: Erlaubt (mit Defaults)

3. **Secrets-Datei nicht gefunden**
   - Prüfung: Secrets-Datei existiert
   - Warnung: `"secrets file not found, some features may be disabled"`
   - Start: Erlaubt

4. **State-DB existiert bereits**
   - Prüfung: State-DB existiert
   - Warnung: `"state database exists, will append"`
   - Start: Erlaubt

5. **Pattern-Registry nicht gefunden**
   - Prüfung: Pattern-Registry existiert
   - Warnung: `"pattern registry not found, no patterns will be matched"`
   - Start: Erlaubt (aber funktionslos)

---

## 3️⃣ Validierungsreihenfolge

### Phase 1: Pfad-Validierung

1. Repo-Pfad ermitteln
2. Config-Pfad prüfen
3. State-DB-Pfad prüfen
4. Secrets-Pfad prüfen
5. **Bei Fehler → Exit 1**

### Phase 2: Datei-Validierung

1. Config-Datei laden
2. Secrets-Datei laden (optional)
3. Pattern-Registry laden (optional)
4. **Bei kritischem Fehler → Exit 1**

### Phase 3: Inhalts-Validierung

1. IP-Adressen prüfen
2. Hostname prüfen (Warnung)
3. Zeitzone prüfen
4. Port prüfen
5. Thresholds prüfen
6. Zeitfenster prüfen
7. **Bei Fehler → Exit 1**

### Phase 4: Rechte-Validierung

1. Schreibrechte für State-DB
2. Schreibrechte für Config-Verzeichnis
3. **Bei Fehler → Exit 1**

### Phase 5: Funktions-Validierung

1. Journal-Zugriff prüfen
2. SQLite-Verbindung prüfen
3. Push-Adapter prüfen (optional)
4. **Bei Fehler → Exit 1**

---

## 4️⃣ Validierungsfunktionen

### `ValidateConfig(c *Config) error`

**Prüfungen:**
- Alle Pfade außerhalb Repo
- Keine IP-Adressen
- Gültige Werte (Port, Zeitzone, etc.)
- Thresholds > 0
- Gültige Zeitfenster

**Rückgabe:**
- `nil` bei Erfolg
- `error` bei Fehler

---

### `ValidatePaths(c *Config, repoPath string) error`

**Prüfungen:**
- Config-Pfad außerhalb Repo
- State-DB-Pfad außerhalb Repo
- Secrets-Pfad außerhalb Repo
- Absolute Pfade

**Rückgabe:**
- `nil` bei Erfolg
- `error` bei Fehler

---

### `ValidateContent(c *Config) error`

**Prüfungen:**
- Keine IP-Adressen (IPv4/IPv6)
- Hostname-Warnung (Punkte)
- Gültige Zeitzone
- Gültiger Port
- Gültige Thresholds
- Gültige Zeitfenster

**Rückgabe:**
- `nil` bei Erfolg
- `error` bei Fehler

---

### `ValidatePermissions(c *Config) error`

**Prüfungen:**
- Schreibrechte für State-DB-Verzeichnis
- Schreibrechte für Config-Verzeichnis
- Leserechte für Config-Datei

**Rückgabe:**
- `nil` bei Erfolg
- `error` bei Fehler

---

## 5️⃣ IP-Erkennung

### IPv4-Pattern

```go
var ipv4Pattern = regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`)
```

**Prüfung:**
- In allen String-Feldern
- Case-insensitive
- Keine Ausnahmen

---

### IPv6-Pattern

```go
var ipv6Pattern = regexp.MustCompile(`\b(?:[0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}\b`)
```

**Prüfung:**
- In allen String-Feldern
- Case-insensitive
- Keine Ausnahmen

---

## 6️⃣ Repo-Pfad-Erkennung

### Prüfung

```go
func IsInRepo(path string, repoPath string) bool {
    absPath, _ := filepath.Abs(path)
    absRepo, _ := filepath.Abs(repoPath)
    return strings.HasPrefix(absPath, absRepo)
}
```

**Verwendung:**
- Config-Pfad
- State-DB-Pfad
- Secrets-Pfad
- Pattern-Registry-Pfad

---

## 7️⃣ Fehlerbehandlung

### Fehler-Typen

```go
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation error in %s: %s", e.Field, e.Message)
}
```

**Verwendung:**
- Spezifische Fehlermeldungen
- Feld-Identifikation
- Exit-Code 1

---

## 8️⃣ Validierungslogik (Pseudocode)

```
1. Lade Config
2. Prüfe Repo-Pfade → Exit 1 bei Fehler
3. Prüfe IP-Adressen → Exit 1 bei Fehler
4. Prüfe Werte → Exit 1 bei Fehler
5. Prüfe Rechte → Exit 1 bei Fehler
6. Warnungen ausgeben
7. Start erlauben
```

---

## 9️⃣ Beispiel-Validierung

### Erfolgreich

```
✓ Config loaded
✓ Paths validated (outside repo)
✓ No IP addresses found
✓ Permissions OK
⚠ Hostname contains dots (ensure placeholder)
→ Starting...
```

### Fehlgeschlagen

```
✗ Config loaded
✗ Paths validated: config path in repository
→ Exit 1: "config path must be outside repository"
```

---

## 🔟 Guards-Aktivierung

### Standard

- Alle Guards aktiv
- Keine Ausnahmen
- Harte Stops

### Override (nur für Tests)

- `--skip-validation` (nur in Tests)
- Produktion: niemals deaktivieren

---

## Status

- Validierungsregeln definiert
- Guards spezifiziert
- Fehlerbehandlung klar
- Datenschutz erzwingbar
