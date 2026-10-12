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
