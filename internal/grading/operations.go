package grading

import (
	"context"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"

	"github.com/ghassan-ai-projects/streams-simulator/internal/grading/internal/app"
	"github.com/ghassan-ai-projects/streams-simulator/internal/score"
)

func GradeRun(ctx context.Context, options Options) (*score.RunScore, error) {
	return app.GradeRun(ctx, options)
}

func Document(graded *score.RunScore) any {
	return app.Document(graded)
}

func GradeOutcome(ctx context.Context, options OutcomeOptions) (*model.OutcomeReport, error) {
	return app.GradeOutcome(ctx, options)
}
