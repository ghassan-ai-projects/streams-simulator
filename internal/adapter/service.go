package adapter

import (
	layer "github.com/ghassan-ai-projects/streams-simulator/internal/adapter/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// Engine renders native events for one run through a validated adapter.
type Engine struct {
	engine *layer.Engine
}

// NewEngine builds a render engine for an adapter and the run metadata its
// preamble and postamble read. A nil adapter is refused with ErrNoAdapter.
func NewEngine(a *model.Adapter, meta map[string]any) (*Engine, error) {
	if a == nil {
		return nil, ErrNoAdapter
	}
	inner, err := layer.NewEngine(a, meta)
	if err != nil {
		return nil, err
	}
	return &Engine{engine: inner}, nil
}
