package grading

import (
	"context"

	"github.com/ghassan-ai-projects/streams-simulator/internal/grading/internal/app"
	"github.com/ghassan-ai-projects/streams-simulator/internal/score"
)

func Grade(ctx context.Context, options Options) (*score.Scorecard, error) {
	return app.Grade(ctx, options)
}
