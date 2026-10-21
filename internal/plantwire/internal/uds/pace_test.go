package uds

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
)

func TestTheWallPacerWaitsForEverySliceOnItsAbsoluteScheduleAndNeverForAPastOne(t *testing.T) {
	t.Parallel()
	const slice = int64(time.Second)
	before := time.Now()
	pacer := StartWallPacer()
	if start := pacer.StartNS(); start > before.UnixNano() || start < before.UnixNano()-int64(time.Millisecond) {
		t.Fatalf("start %d, want the wall time at start floored to the millisecond (read at %d)", start, before.UnixNano())
	}
	for k := int64(1); k <= 3; k++ {
		if wait := pacer.WaitNS(pacer.StartNS() + k*slice); wait > k*slice || wait < k*slice-int64(time.Second) {
			t.Fatalf("slice %d waits %v, want its wall time minus the time since the start", k, time.Duration(wait))
		}
	}
	if wait := pacer.WaitNS(pacer.StartNS() - int64(time.Hour)); wait != 0 {
		t.Fatalf("a past instant waits %v, want none", time.Duration(wait))
	}
}

func TestALinkDrainedForAWaitReadsEveryRecordTheHostSendsMeanwhile(t *testing.T) {
	t.Parallel()
	path, listener := privateListener(t)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()
		line, _ := json.Marshal(map[string]any{"type": "STATE"})
		for range 200 {
			if _, err := conn.Write(append(line, '\n')); err != nil {
				return
			}
		}
		<-t.Context().Done()
	}()
	link, err := Dial(t.Context(), path, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = link.Close() }()
	started := time.Now()
	held, err := link.Drain(t.Context(), 200*time.Millisecond)
	if err != nil || len(held) != 200 || time.Since(started) < 200*time.Millisecond {
		t.Fatalf("Drain = %d records, %v after %v, want all 200 read during the whole wait", len(held), err, time.Since(started))
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := link.Drain(ctx, time.Hour); !errors.Is(err, context.Canceled) {
		t.Fatalf("an interrupted drain: %v, want context.Canceled", err)
	}
}

func TestTheWallPacersWorldNowIsTheWallClockOnItsSchedule(t *testing.T) {
	t.Parallel()
	pacer := NewWallPacer(domain.WallSchedule{WorldStartNS: 1_000, WallStartNS: time.Now().UnixNano()})
	if now := pacer.WorldNowNS(); now < 1_000 || now > 1_000+int64(5*time.Second) {
		t.Fatalf("world now %d, want the world start plus the wall time since the schedule began", now)
	}
}
