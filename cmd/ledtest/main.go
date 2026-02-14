package main

import (
	"fmt"
	"os"

	"prox-watch/internal/ui"
)

func main() {
	// Terminal UI starten
	menu := ui.NewMenu()
	if err := menu.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Fehler: %v\n", err)
		os.Exit(1)
	}
}
