// Package files is the file-system edge of the device module: it reads the
// capability catalog a device is configured with and hands the bytes to the
// pure layer, which parses and validates them.
package files

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/device/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/kernel"
)

func LoadCapabilities(path string) (*domain.Capabilities, error) {
	data, err := kernel.Read(path)
	if err != nil {
		return nil, fmt.Errorf("read capabilities: %w", err)
	}
	caps, err := domain.LoadCapabilities(data)
	if err != nil {
		return nil, fmt.Errorf("load capabilities: %w", err)
	}
	return caps, nil
}
