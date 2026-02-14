# GPIO Test Tool - Setup & Build

## Schritt 2: Go Binding integrieren

### Voraussetzung: Projekt-Verzeichnis finden

Das Projekt muss zuerst geklont werden (falls noch nicht vorhanden):

```bash
# Repository klonen
git clone https://github.com/tnt-nitro/prox_watch_monitoring_system.git
cd prox_watch_monitoring_system
```

**Oder** falls das Projekt bereits existiert (z.B. nach Installation):

```bash
# Prüfe, ob Repository existiert
ls -la ~/prox_watch_monitoring_system
# Oder
ls -la /tmp/prox-watch-install

# Falls vorhanden, dorthin wechseln
cd ~/prox_watch_monitoring_system
# Oder
cd /tmp/prox-watch-install
```

### Auf dem Raspberry Pi ausführen:

```bash
# 1. In das Projekt-Verzeichnis wechseln (siehe oben)
cd /path/to/prox_watch_monitoring_system

# 2. Go Binding installieren
go get github.com/rpi-ws281x/rpi-ws281x-go

# 3. Abhängigkeiten aktualisieren
go mod tidy

# 4. CGO prüfen (muss 1 sein)
go env CGO_ENABLED

# Falls nicht 1, aktivieren:
export CGO_ENABLED=1

# 5. Build-Test
go build ./cmd/ledtest
```

### Erwartetes Ergebnis:

✅ **Build erfolgreich** - Keine Linker-Fehler bezüglich ws2811

❌ **Linker-Fehler** - Falls Fehler auftreten, prüfe:
- Ist `libws2811.a` vorhanden? (Schritt 1 muss erfolgreich sein)
- Ist CGO_ENABLED=1?
- Sind alle Build-Tools installiert?
