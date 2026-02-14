//go:build raspberry
// +build raspberry

package gpio

import (
	"fmt"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
)

// raspberryPin ist die Raspberry Pi Hardware-Implementierung
type raspberryPin struct {
	pin gpio.PinIO
}

// NewPin erstellt einen neuen GPIO-Pin (Raspberry Pi Hardware)
func NewPin(pinNum int, mode PinMode) (Pin, error) {
	pinName := fmt.Sprintf("GPIO%d", pinNum)
	pin := gpioreg.ByName(pinName)
	if pin == nil {
		return nil, fmt.Errorf("pin %s not found", pinName)
	}

	switch mode {
	case PinModeOutput:
		if err := pin.Out(gpio.Low); err != nil {
			return nil, fmt.Errorf("failed to set pin %s to output: %w", pinName, err)
		}
	case PinModeInput:
		if err := pin.In(gpio.Float, gpio.NoEdge); err != nil {
			return nil, fmt.Errorf("failed to set pin %s to input: %w", pinName, err)
		}
	case PinModeInputPullUp:
		if err := pin.In(gpio.PullUp, gpio.NoEdge); err != nil {
			return nil, fmt.Errorf("failed to set pin %s to input (pull-up): %w", pinName, err)
		}
	case PinModeInputPullDown:
		if err := pin.In(gpio.PullDown, gpio.NoEdge); err != nil {
			return nil, fmt.Errorf("failed to set pin %s to input (pull-down): %w", pinName, err)
		}
	default:
		return nil, fmt.Errorf("unsupported pin mode: %d", mode)
	}

	return &raspberryPin{pin: pin}, nil
}

func (p *raspberryPin) High() error {
	return p.pin.Out(gpio.High)
}

func (p *raspberryPin) Low() error {
	return p.pin.Out(gpio.Low)
}

func (p *raspberryPin) Read() (PinState, error) {
	level := p.pin.Read()
	if level == gpio.High {
		return PinStateHigh, nil
	}
	return PinStateLow, nil
}

func (p *raspberryPin) Close() error {
	// periph.io Pins müssen nicht explizit geschlossen werden
	// Aber wir setzen den Pin auf LOW für Sicherheit
	return p.pin.Out(gpio.Low)
}
