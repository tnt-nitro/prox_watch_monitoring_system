# GPIO Test Tool - Ein Skript für alles

## Auf dem Raspberry Pi ausführen:

**Schritt 1: Skript erstellen**

```bash
cd ~/prox_watch_monitoring_system
cat > create_ledtest.sh << 'SCRIPTEOF'
#!/bin/bash
# Erstellt alle Dateien für LED-Test-Tool

mkdir -p cmd/ledtest internal/gpio internal/neopixel internal/ui

# cmd/ledtest/main.go
cat > cmd/ledtest/main.go << 'EOF'
package main

import (
	"fmt"
	"os"

	"prox-watch/internal/ui"
)

func main() {
	menu := ui.NewMenu()
	if err := menu.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Fehler: %v\n", err)
		os.Exit(1)
	}
}
EOF

# internal/gpio/gpio.go
cat > internal/gpio/gpio.go << 'EOF'
package gpio

const (
	NeoPixelPin = 18
	BeeperPin = 24
	ButtonPin = 23
)
EOF

# internal/gpio/pin.go
cat > internal/gpio/pin.go << 'EOF'
package gpio

type PinMode int
const (
	PinModeInput PinMode = iota
	PinModeInputPullUp
	PinModeInputPullDown
	PinModeOutput
)

type PinState int
const (
	PinStateLow PinState = iota
	PinStateHigh
)

type Pin interface {
	High() error
	Low() error
	Read() (PinState, error)
	Close() error
}
EOF

# Weitere Dateien... (siehe vollständiges Skript)
SCRIPTEOF

chmod +x create_ledtest.sh
```

**Aber:** Das wird zu lang. Besser: Kopiere das vollständige Skript von `installer/create_all_ledtest_files.sh` direkt auf den Raspberry Pi.

## Einfachste Lösung:

**Option A: Skript von GitHub herunterladen (wenn verfügbar)**

```bash
cd ~/prox_watch_monitoring_system
curl -fsSL https://raw.githubusercontent.com/tnt-nitro/prox_watch_monitoring_system/main/installer/create_all_ledtest_files.sh -o create_ledtest.sh
chmod +x create_ledtest.sh
bash create_ledtest.sh
```

**Option B: Dateien manuell kopieren (wenn du Zugriff auf das lokale System hast)**

Von deinem Windows-System aus:
```bash
scp installer/create_all_ledtest_files.sh pi@<raspberry-pi-ip>:~/prox_watch_monitoring_system/
```

Dann auf dem Raspberry Pi:
```bash
cd ~/prox_watch_monitoring_system
bash create_ledtest.sh
```
