package plantwire

import (
	"context"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/app"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

func Serve(ctx context.Context, r *run.Run, socketPath string, barrierCap time.Duration, options Options) error {
	return app.Serve(ctx, r, socketPath, barrierCap, options)
}

func ContractDigest() string {
	return domain.ContractDigest()
}
