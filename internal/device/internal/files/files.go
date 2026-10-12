// Package files is the file-system edge of the device module: it reads the
// capability catalog a device is configured with and hands the bytes to the
// pure layer, which parses and validates them.
package files

import (
	"fmt"
	"os"

	"github.com/ghassan-ai-projects/streams-simulator/internal/device/internal/domain"
)

// LoadCapabilities reads and validates the capability catalog at path.
func LoadCapabilities(path string) (*domain.Capabilities, error) {
	// #nosec G304 G703 -- a CLI flag naming a file is user intent.
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read capabilities: %w", err)
	}
	caps, err := domain.LoadCapabilities(data)
	if err != nil {
		return nil, fmt.Errorf("load capabilities: %w", err)
	}
	return caps, nil
}
