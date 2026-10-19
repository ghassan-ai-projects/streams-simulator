package app

import (
	"errors"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/files"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/host"
)

var ErrHostInput = errors.New("plantwire: a host script or inventory does not parse")

func LoadHost(options host.Options, scriptPath, inventoryPath string) (host.Options, error) {
	var err error
	if scriptPath != "" {
		options.Script, err = readHostInput(scriptPath, host.ParseScript)
	}
	if err == nil && inventoryPath != "" {
		options.Inventory, err = readHostInput(inventoryPath, host.ParseInventory)
	}
	return options, err
}

func readHostInput[T any](path string, parse func([]byte) (T, error)) (T, error) {
	var zero T
	raw, err := files.ReadHostInput(path)
	if err != nil {
		return zero, err
	}
	parsed, err := parse(raw)
	if err != nil {
		return zero, fmt.Errorf("%w: %w", ErrHostInput, err)
	}
	return parsed, nil
}
