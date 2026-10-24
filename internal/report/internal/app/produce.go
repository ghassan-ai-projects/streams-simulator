package app

import (
	"context"

	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

type Request struct {
	Dir         string
	A, B        string
	Replay      bool
	Directories run.Directories
}

func Produce(ctx context.Context, request Request) (any, error) {
	switch {
	case request.A != "" || request.B != "":
		return compared(ctx, request)
	case request.Replay:
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
