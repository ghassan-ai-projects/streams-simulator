package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/domain"
)

func TestFollowingARunEmitsEachLineOnceByIndexAndStopsAtTheClose(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	journal := `{"type":"run.opened"}` + "\n" + `{"type":"observation"}` + "\n"
	h.write(t, "runs/w-1/journal.jsonl", journal+`{"type":"obs`)
	var got []domain.Record
	done := make(chan error, 1)
	go func() {
		done <- h.bridge.Follow(context.Background(), "w-1", 0, func(record domain.Record) error {
			got = append(got, record)
			return nil
		})
	}()
	h.ticks <- time.Time{}
	h.write(t, "runs/w-1/journal.jsonl", journal+`{"type":"observation"}`+"\n"+`{"type":"run.closed"}`+"\n")
	if err := tickUntil(h, done); err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 || got[0].Index != 1 || got[2].Type != "run.closed" || !got[2].Closes {
		t.Fatalf("records after index 0 = %+v", got)
	}
}

func TestFollowRefusesABadRunAndStopsOnABrokenLineOrTheContext(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	if err := h.bridge.Follow(context.Background(), "..", -1, nil); !errors.Is(err, domain.ErrBadPath) {
		t.Fatalf("a run id that climbs: %v", err)
	}
	h.write(t, "runs/w-1/journal.jsonl", "not json\n")
	if err := h.bridge.Follow(context.Background(), "w-1", -1, nil); err == nil {
		t.Fatal("a broken journal line was streamed")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := h.bridge.Follow(ctx, "w-2", -1, nil); err != nil {
		t.Fatalf("a journal that never appears, cancelled: %v", err)
	}
}

func TestWatchingTellsEachChangeOfTheWorkspace(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	ctx, cancel := context.WithCancel(context.Background())
	changes := make(chan struct{}, 4)
	done := make(chan error, 1)
	go func() {
		done <- h.bridge.Watch(ctx, func() error {
			changes <- struct{}{}
			return nil
		})
	}()
	h.ticks <- time.Time{}
	h.write(t, plant, "{}")
	h.ticks <- time.Time{}
	<-changes
	cancel()
	if err := <-done; err != nil || len(changes) != 0 {
		t.Fatalf("Watch: %v, %d extra changes", err, len(changes))
	}
}

func tickUntil(h *harness, done <-chan error) error {
	for {
		select {
		case err := <-done:
			return err
		case h.ticks <- time.Time{}:
		}
	}
}
