package main

import (
	"fmt"
	"os"
)

// main ist der Entry-Point für den externen Wächter (Raspberry Pi).
// Phase 1: MVP - Erreichbarkeitsprüfung, Push, LED/Beeper
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: prox-watch-watcher <command>")
		fmt.Println("Commands:")
		fmt.Println("  run     - Start watcher daemon")
		fmt.Println("  status  - Show current status")
		fmt.Println("  version - Show version")
		os.Exit(1)
	}

	command := os.Args[1]
	switch command {
	case "run":
		// TODO: Implement watcher daemon
		fmt.Println("Watcher daemon (not yet implemented)")
	case "status":
		// TODO: Implement status command
		fmt.Println("Status (not yet implemented)")
	case "version":
		fmt.Println("prox-watch-watcher v0.1.0 (Phase 1)")
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}
