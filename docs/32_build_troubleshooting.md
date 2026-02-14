# Build Troubleshooting

## Problem: missing go.sum entry

Wenn beim Build folgende Fehler auftreten:
```
missing go.sum entry for module providing package golang.org/x/term
missing go.sum entry for go.mod file
```

### Lösung:

```bash
cd ~/prox_watch_monitoring_system

# Abhängigkeiten aktualisieren
go mod tidy

# Dann neu bauen
go build -tags raspberry ./cmd/ledtest
```

## Problem: Binary nicht gefunden

Wenn `./ledtest` nicht gefunden wird:

```bash
# Prüfe, ob das Binary erstellt wurde
ls -lh ledtest

# Falls nicht im aktuellen Verzeichnis:
find . -name "ledtest" -type f

# Falls gefunden, mit vollem Pfad ausführen:
sudo ./ledtest
# Oder:
sudo /home/pi/prox_watch_monitoring_system/ledtest
```
