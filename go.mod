module prox-watch

go 1.21

require (
	github.com/mattn/go-sqlite3 v1.14.17
	gopkg.in/yaml.v3 v3.0.1
)

// Note: Run 'go mod tidy' after adding dependencies
// Required dependencies:
// - gopkg.in/yaml.v3 for YAML parsing
// - github.com/mattn/go-sqlite3 for SQLite database
//
// Optional dependencies (for Raspberry Pi hardware GPIO):
// - periph.io/x/conn/v3/gpio for GPIO pin control
// - periph.io/x/conn/v3/gpio/gpioreg for GPIO pin registry
// - periph.io/x/host/v3 for periph.io host initialization
//
// To build with hardware GPIO support:
//   go build -tags raspberry -o prox-watch-watcher ./cmd/watcher
//
// To install periph.io dependencies:
//   go get periph.io/x/conn/v3/gpio
//   go get periph.io/x/conn/v3/gpio/gpioreg
//   go get periph.io/x/host/v3