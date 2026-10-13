package process

import (
	"context"
	"errors"
	"testing"
)

func TestAnInterruptContextEndsWithItsParent(t *testing.T) {
	t.Parallel()
	parent, cancel := context.WithCancel(t.Context())
	ctx, stop := UntilInterrupted(parent)
	defer stop()
	cancel()
	<-ctx.Done()
	if !errors.Is(ctx.Err(), context.Canceled) {
		t.Fatalf("ctx = %v", ctx.Err())
	}
}
