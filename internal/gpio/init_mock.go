//go:build !raspberry
// +build !raspberry

package gpio

// Init ist eine No-Op für Mock-Builds
func Init() error {
	return nil
}
