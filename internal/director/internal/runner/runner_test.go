package runner

import (
	"context"
	"errors"
	"testing"
)

func TestAPacerAtAHighRateLetsSimulatedTimeThroughAtOnce(t *testing.T) {
	t.Parallel()
	pacer := NewPacer(1e12)
	if err := pacer.WaitUntil(t.Context(), 3_600_000_000_000); err != nil {
		t.Fatalf("WaitUntil = %v, want the wait to end at once", err)
	}
}

func TestAPacerWaitEndsWhenItsContextIsCancelled(t *testing.T) {
	t.Parallel()
	pacer := NewPacer(1)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if err := pacer.WaitUntil(ctx, 3_600_000_000_000); !errors.Is(err, context.Canceled) {
		t.Fatalf("WaitUntil = %v, want context.Canceled", err)
	}
}

func TestAPacerWithoutARateNeverWaits(t *testing.T) {
	t.Parallel()
	if err := NewPacer(0).WaitUntil(t.Context(), 3_600_000_000_000); err != nil {
		t.Fatalf("WaitUntil = %v, want no wait", err)
	}
}

func TestGoRunsTheWorkOnAnotherGoroutine(t *testing.T) {
	t.Parallel()
	done := make(chan struct{})
	Go(func() { close(done) })
	<-done
}
