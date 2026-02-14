package gpio

// PinMode definiert den Modus eines GPIO-Pins
type PinMode int

const (
	PinModeInput PinMode = iota
	PinModeInputPullUp
	PinModeInputPullDown
	PinModeOutput
)

// PinState definiert den Zustand eines GPIO-Pins
type PinState int

const (
	PinStateLow PinState = iota
	PinStateHigh
)

// Pin ist die Schnittstelle für GPIO-Pin-Zugriff
type Pin interface {
	// High setzt den Pin auf HIGH
	High() error

	// Low setzt den Pin auf LOW
	Low() error

	// Read liest den Pin-Zustand
	Read() (PinState, error)

	// Close schließt den Pin
	Close() error
}
