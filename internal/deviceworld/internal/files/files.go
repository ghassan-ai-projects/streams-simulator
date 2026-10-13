// Package files is the file-system edge of the deviceworld module: it reads the
// binding catalog a device is composed with and hands the bytes to the pure
// layer, which parses and validates them.
package files

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/kernel"
)

func ReadBindings(path string) ([]byte, error) {
	data, err := kernel.Read(path)
	if err != nil {
		return nil, fmt.Errorf("read world bindings: %w", err)
	}
	return data, nil
}
