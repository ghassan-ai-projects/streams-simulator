package app

import (
	"context"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/domain"
)

func TestAFullReadWindowDoesNotClaimTheReaderIsCaughtUp(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.write(t, "runs/w-1/journal.jsonl", strings.Repeat("x", followChunk))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var emitted []domain.Record
	err := h.bridge.Follow(ctx, "w-1", -1, func(record domain.Record) error {
		emitted = append(emitted, record)
		return nil
	})
	if err != nil || len(emitted) != 0 {
		t.Fatalf("Follow over a full read window = %v with %d records, want nothing and no caught-up signal", err, len(emitted))
	}
}
