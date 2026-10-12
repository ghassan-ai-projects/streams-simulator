// Package files is the file-system edge of the deviceworld module: it reads the
// binding catalog a device is composed with and hands the bytes to the pure
// layer, which parses and validates them.
package files

import (
	"fmt"
	"os"
)

// ReadBindings returns the contents of the binding catalog at path.
func ReadBindings(path string) ([]byte, error) {
	// #nosec G304 G703 -- a CLI flag naming a file is user intent.
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read world bindings: %w", err)
	}
	return data, nil
}
