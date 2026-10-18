package run

import (
	"context"

	layer "github.com/ghassan-ai-projects/streams-simulator/internal/run/internal/app"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

type Run struct {
	ID     string
	Config Config
	World  *world.World

	run *layer.Run
}

func New(ctx context.Context, cfg Config) (*Run, error) {
	inner, err := layer.New(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return &Run{ID: inner.ID, Config: inner.Config, World: inner.World, run: inner}, nil
}

type FieldTrace struct {
	impl *layer.FieldTrace
}

func NewFieldTrace(dir string) (*FieldTrace, error) {
	inner, err := layer.OpenFieldTrace(dir)
	if err != nil {
		return nil, err
	}
	return &FieldTrace{impl: inner}, nil
}

type WireTrace struct {
	impl *layer.WireTrace
}

func NewWireTrace(dir string) (*WireTrace, error) {
	inner, err := layer.OpenWireTrace(dir)
	if err != nil {
		return nil, err
	}
	return &WireTrace{impl: inner}, nil
}
