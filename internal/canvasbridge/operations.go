package canvasbridge

import (
	"context"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/app"
)

func Serve(ctx context.Context, cfg Config) error {
	return app.Serve(ctx, cfg)
}
