package app

import (
	"errors"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/component"
)

func listComponents(library *component.Library, options componentOptions) (any, error) {
	low, high := options.reading.bounds()
	listed, err := library.ListFor(options.unit, low, high)
	if errors.Is(err, component.ErrReadingRange) {
		return nil, &usageError{message: err.Error(), cause: err}
	}
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return map[string]any{"components": listed}, nil
}
