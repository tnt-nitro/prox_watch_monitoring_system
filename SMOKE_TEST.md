# Smoke Test - MVP v0.1

## Voraussetzungen

- Go 1.21+ installiert
- Linux-System mit systemd (für echte Tests)
- Root-Rechte (für Installation)

## Test-Schritte

### 1. Build

```bash
go build -o prox-watch ./cmd/prox-watch
```

**Erwartung:** Binary wird erstellt ohne Fehler

### 2. Unit-Tests

```bash
go test ./...
```

**Erwartung:** Alle Tests bestehen

**Optional - Coverage:**
```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 3. Config-Initialisierung

```bash
sudo mkdir -p /var/lib/prox-watch
sudo chown $USER:$USER /var/lib/prox-watch
./prox-watch init --config-path /var/lib/prox-watch/config.yaml
```

**Erwartung:**
- Wizard startet
- Config wird erstellt
- Secrets werden erstellt
- Rechte: 0600 für secrets.yaml

### 4. Config-Validierung

```bash
# Prüfe, dass Config erstellt wurde
cat /var/lib/prox-watch/config.yaml

# Prüfe, dass keine IPs enthalten sind
grep -E '\b([0-9]{1,3}\.){3}[0-9]{1,3}\b' /var/lib/prox-watch/config.yaml
# Erwartung: Keine Ausgabe
```

**Erwartung:** Config ist valide, keine IPs

### 5. Status-Kommando

```bash
./prox-watch status --config-path /var/lib/prox-watch/config.yaml
```

**Erwartung:**
- Status wird angezeigt
- "No events found" (bei leerer DB)

### 6. Daemon-Modus (Fake-Reader)

```bash
# Starte in Foreground (mit Fake-Reader)
./prox-watch run --config-path /var/lib/prox-watch/config.yaml
# Stoppe mit Ctrl+C
```

**Erwartung:**
- Service startet ohne Fehler
- Graceful Shutdown bei Ctrl+C

### 7. State-DB-Persistenz

```bash
# Erstelle Test-Event (manuell oder über Code)
# Prüfe State-DB
sqlite3 /var/lib/prox-watch/state.db "SELECT * FROM events;"
```

**Erwartung:** State-DB existiert und ist lesbar

### 8. systemd Service (optional)

```bash
# Installiere Service
sudo cp installer/prox-watch.service /etc/systemd/system/
sudo systemctl daemon-reload

# Prüfe Service-Status
sudo systemctl status prox-watch.service

# Starte Service (wenn konfiguriert)
sudo systemctl start prox-watch.service

# Prüfe Logs
sudo journalctl -u prox-watch.service -f
```

**Erwartung:** Service startet ohne Fehler

## Erwartete Ergebnisse

### ✅ Erfolgreich

- Alle Tests bestehen
- Config wird erstellt
- Status-Kommando funktioniert
- Daemon startet und stoppt graceful
- State-DB ist persistiert

### ❌ Fehler

- Build-Fehler → Go-Version prüfen
- Test-Fehler → Dependencies prüfen
- Config-Fehler → Validierung prüfen
- Service-Fehler → Permissions prüfen

## Bekannte Einschränkungen

1. **Journal Reader** - Nur Fake-Implementierung, keine echten Logs
2. **Pattern Registry** - Pattern-Loading noch nicht vollständig
3. **Hard-Error-Erkennung** - `IsHardError()` gibt immer `false` zurück

## Nächste Schritte

Nach erfolgreichem Smoke-Test:

1. **Release v0.1.0** - Tag setzen
2. **Phase 1** - Journal Reader (systemd) implementieren
3. **Pattern Registry** - YAML-Loading vollständig implementieren
