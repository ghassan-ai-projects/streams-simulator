package domain

import (
	"errors"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

var ErrUnknownAdapter = errors.New("unknown adapter")

func Lookup(adapters map[string]*model.Adapter, id string) (*model.Adapter, error) {
	if found, ok := adapters[id]; ok {
		return found, nil
	}
	return nil, fmt.Errorf("%w %q", ErrUnknownAdapter, id)
}
