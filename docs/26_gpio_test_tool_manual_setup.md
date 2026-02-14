# GPIO Test Tool - Manuelle Einrichtung (Temporär)

## Problem

Das `cmd/ledtest` Verzeichnis existiert noch nicht auf GitHub, da der Branch `feature/gpio-test-tool` noch nicht gepusht wurde.

## Temporäre Lösung: Dateien manuell erstellen

### Auf dem Raspberry Pi ausführen:

```bash
# 1. In das Projekt-Verzeichnis wechseln
cd ~/prox_watch_monitoring_system

# 2. Verzeichnisse erstellen
mkdir -p cmd/ledtest
mkdir -p internal/gpio
mkdir -p internal/neopixel
mkdir -p internal/ui

# 3. Dateien erstellen (siehe nächste Abschnitte)
```

### Dateien erstellen

Die Dateien müssen manuell erstellt werden. Siehe die entsprechenden Dateien im Repository:
- `cmd/ledtest/main.go`
- `internal/gpio/*.go`
- `internal/neopixel/strip.go`
- `internal/ui/menu.go`

**Oder:** Warte, bis der Branch auf GitHub ist, dann:

```bash
git fetch origin
git checkout feature/gpio-test-tool
```

## Zukünftige Lösung

Das Setup-Skript wird erweitert, um:
1. WS2812-Bibliothek automatisch zu installieren
2. Go-Bindings automatisch zu installieren
3. LED-Test-Tool automatisch zu bauen
4. Alles in einem Schritt
