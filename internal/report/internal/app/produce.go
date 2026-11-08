package app

import (
	"context"

	"github.com/ghassan-ai-projects/streams-simulator/internal/report/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

type Request struct {
	Dir         string
	A, B        string
	Replay      bool
	Outcome     bool
	Directories run.Directories
}

func Produce(ctx context.Context, request Request) (any, error) {
	if request.A != "" || request.B != "" {
		return compared(ctx, request)
	}
	built, err := buildRequested(ctx, request)
	if err != nil || !request.Outcome {
		return built, err
	}
	return withOutcome(ctx, request.Dir, built)
}

func buildRequested(ctx context.Context, request Request) (*domain.Report, error) {
	if request.Replay {
		return BuildVerified(ctx, request.Dir, request.Directories)
	}
	return Build(ctx, request.Dir)
}

func compared(ctx context.Context, request Request) (any, error) {
	comparison, err := Compare(ctx, request.A, request.B)
	if err != nil {
		return nil, err
	}
	return comparison, nil
}
