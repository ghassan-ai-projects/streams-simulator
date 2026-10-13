package adapter

import (
	layer "github.com/ghassan-ai-projects/streams-simulator/internal/adapter/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

type Engine struct {
	engine *layer.Engine
}

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
