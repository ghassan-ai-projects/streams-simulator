// Package process is the kernel's process edge: a context that ends on interrupt or SIGTERM.
package process

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func UntilInterrupted(ctx context.Context) (context.Context, context.CancelFunc) {
	return signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
}
