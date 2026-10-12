package grading

import (
	"context"

	"github.com/ghassan-ai-projects/streams-simulator/internal/grading/internal/app"
	"github.com/ghassan-ai-projects/streams-simulator/internal/score"
)

// Grade scores a finished run from its artifacts with the scorer the director
// uses online, so the two scorecards agree. Delivery and state evidence come
// from the files the run published beside its artifact; the effector calls
// come from replaying the artifact. A replay that does not reproduce the
// recorded trace is refused with a *ReplayMismatchError.
func Grade(ctx context.Context, options Options) (*score.Scorecard, error) {
	return app.Grade(ctx, options)
}
