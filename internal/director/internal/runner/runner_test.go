package runner

import (
	"context"
	"errors"
	"testing"
	"time"
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

func TestReanchoringToAFasterRateOrToRateZeroEndsAWaitThatWouldHaveTakenAnHour(t *testing.T) {
	t.Parallel()
	for name, rate := range map[string]float64{"a higher rate": 1e12, "rate zero": 0} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			pacer := NewPacer(1)
			waited := make(chan error, 1)
			go func() { waited <- pacer.WaitUntil(t.Context(), 3_600_000_000_000) }()
			pacer.Reanchor(0, rate)
			select {
			case err := <-waited:
				if err != nil {
					t.Fatalf("WaitUntil = %v, want the wait to end cleanly at the new rate", err)
				}
			case <-time.After(time.Minute):
				t.Fatal("the wait ignored the new rate")
			}
		})
	}
}

func TestReanchoringMovesTheAnchorToNowSoPausedWallTimeIsNotCaughtUp(t *testing.T) {
	t.Parallel()
	const hour = int64(3_600_000_000_000)
	pacer := &Pacer{rate: 1, anchorWall: time.Now().Add(-time.Hour), changed: make(chan struct{})}
	if wait := pacer.wallWait(hour); wait > 0 {
		t.Fatalf("an hour of wall time at rate 1 owes no wait for an hour of simulated time, got %v", wait)
	}
	pacer.Reanchor(hour, 1)
	if wait := pacer.wallWait(hour + 10_000_000_000); wait <= 5*time.Second || wait > 10*time.Second {
		t.Fatalf("wait = %v after the re-anchor, want about 10 s rather than a catch-up", wait)
	}
}

func TestAWaitCancelledAfterAReanchorStillReportsTheCancellation(t *testing.T) {
	t.Parallel()
	pacer := NewPacer(1)
	ctx, cancel := context.WithCancel(t.Context())
	waited := make(chan error, 1)
	go func() { waited <- pacer.WaitUntil(ctx, 3_600_000_000_000) }()
	pacer.Reanchor(0, 1)
	cancel()
	if err := <-waited; !errors.Is(err, context.Canceled) {
		t.Fatalf("WaitUntil = %v, want context.Canceled", err)
	}
}
