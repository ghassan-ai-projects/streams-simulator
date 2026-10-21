package uds

import (
	"context"

	"github.com/ghassan-ai-projects/streams-simulator/internal/kernel"
)

func UntilInterrupted(ctx context.Context) (context.Context, context.CancelFunc) {
	return kernel.UntilInterrupted(ctx)
}

func PeerGone(err error) bool {
	return kernel.PeerGone(err)
}
