package uds

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func UntilInterrupted(ctx context.Context) (context.Context, context.CancelFunc) {
	return signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
}
