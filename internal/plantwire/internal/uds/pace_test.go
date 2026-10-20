package uds

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
)

func TestTheWallPacerReleasesEverySliceOnItsAbsoluteScheduleAndNeverEarly(t *testing.T) {
	t.Parallel()
	const slice = 25 * time.Millisecond
	before := time.Now()
	pacer := StartWallPacer()
	if start := pacer.StartNS(); start > before.UnixNano() || start < before.UnixNano()-int64(time.Millisecond) {
		t.Fatalf("start %d, want the wall time at start floored to the millisecond (read at %d)", start, before.UnixNano())
	}
	for k := int64(1); k <= 4; k++ {
		if err := pacer.Release(t.Context(), pacer.StartNS()+k*int64(slice)); err != nil {
			t.Fatal(err)
		}
		if early := time.Until(time.Unix(0, pacer.StartNS()+k*int64(slice))); early > 0 {
			t.Fatalf("slice %d released %v before its wall time", k, early)
		}
	}
	if elapsed := time.Since(before); elapsed < 4*slice-time.Millisecond || elapsed > 4*slice+5*time.Second {
		t.Fatalf("four slices of %v took %v", slice, elapsed)
	}
}

func TestTheWallPacerDoesNotWaitForAPastInstantAndStopsWhenInterrupted(t *testing.T) {
	t.Parallel()
	pacer := NewWallPacer(domain.WallScheduleFrom(time.Now().UnixNano()))
	before := time.Now()
	if err := pacer.Release(t.Context(), pacer.StartNS()-int64(time.Hour)); err != nil || time.Since(before) > time.Second {
		t.Fatalf("a past instant: %v after %v, want an immediate release", err, time.Since(before))
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if err := pacer.Release(ctx, pacer.StartNS()+int64(time.Hour)); !errors.Is(err, context.Canceled) || !Interrupted(err) {
		t.Fatalf("an interrupted wait: %v, want context.Canceled", err)
	}
}
