//go:build raspberry
// +build raspberry

package gpio

import (
	"sync"

	"periph.io/x/host/v3"
)

var (
	initOnce sync.Once
	initErr  error
)

// Init initialisiert periph.io (nur einmal, thread-safe)
func Init() error {
	initOnce.Do(func() {
		_, initErr = host.Init()
	})
	return initErr
}
