# Build-Tags für GPIO-Implementierung

## Übersicht

Die GPIO-Implementierung verwendet Build-Tags, um zwischen Mock- und Hardware-Implementierung zu wählen.

## Build-Tags

### `!raspberry` (Default)

- **Datei:** `gpio_raspberry.go`
- **Implementierung:** MockPins (für Tests und Entwicklung ohne Hardware)
- **Verwendung:** Standard-Build ohne Tags

```bash
go build -o prox-watch-watcher ./cmd/watcher
```

### `raspberry`

- **Datei:** `gpio_raspberry_hw.go`
- **Implementierung:** Echte periph.io Hardware-Anbindung
- **Verwendung:** Build mit `-tags raspberry`

```bash
go build -tags raspberry -o prox-watch-watcher ./cmd/watcher
```

## Dependencies

### Mock-Implementierung (Default)

Keine zusätzlichen Dependencies erforderlich.

### Hardware-Implementierung

Erfordert periph.io Dependencies:

```bash
go get periph.io/x/conn/v3/gpio
go get periph.io/x/conn/v3/gpio/gpioreg
go get periph.io/x/host/v3
```

## Tests

Tests laufen standardmäßig mit Mock-Implementierung (ohne Hardware):

```bash
go test ./internal/watcher
```

Um Tests mit Hardware-Implementierung auszuführen (nur auf Raspberry Pi):

```bash
go test -tags raspberry ./internal/watcher
```

## Hinweise

- **Default:** Mock-Implementierung (für Entwicklung und CI/CD)
- **Hardware:** Nur auf Raspberry Pi mit periph.io Dependencies
- **Kein CGO:** periph.io ist Pure Go, keine CGO-Abhängigkeit
- **Kein Root:** periph.io sollte ohne Root-Rechte laufen (je nach GPIO-Konfiguration)
