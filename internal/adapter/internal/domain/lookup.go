package domain

import (
	"errors"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// ErrUnknownAdapter is returned by Lookup for an id the set does not hold.
var ErrUnknownAdapter = errors.New("unknown adapter")

// Lookup returns the adapter with the given id from a loaded set.
func Lookup(adapters map[string]*model.Adapter, id string) (*model.Adapter, error) {
	if found, ok := adapters[id]; ok {
		return found, nil
	}
	return nil, fmt.Errorf("%w %q", ErrUnknownAdapter, id)
}
