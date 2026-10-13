package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	rules "github.com/ghassan-ai-projects/streams-simulator/internal/run/internal/domain"
)

func (r *Run) Advance(ctx context.Context, toNS int64, awaitConsumer bool) (int, error) {
	emitted, err := r.advanceLocked(toNS, awaitConsumer)
	if err != nil || !awaitConsumer {
		return emitted, err
	}
	if err := r.waitForConsumer(ctx, toNS); err != nil {
		return emitted, err
	}
	return emitted, nil
}

func (r *Run) advanceLocked(toNS int64, awaitConsumer bool) (int, error) {
	r.commandMu.Lock()
	defer r.commandMu.Unlock()
	if err := r.refuseWhenFinished("Advance"); err != nil {
		return 0, err
	}
	if rules.ExceedsRunSpan(r.worldStartTimeNS, toNS) {
		return 0, fmt.Errorf("run: advance to %d: %w", toNS, ErrRunSpanExceeded)
	}
	return r.advanceThroughBoundary(toNS, awaitConsumer)
}

func (r *Run) advanceThroughBoundary(toNS int64, awaitConsumer bool) (int, error) {
	emitted, err := r.advanceWorld(toNS)
	if err != nil {
		return 0, err
	}
	r.flushPendingDeliveries(toNS)
	if r.runErr != nil {
		return emitted, r.runErr
	}

	r.journal.ClockAdvanced(r.journalNow(), toNS)
	r.recordWorldCommand(model.OpClockAdvance, map[string]any{"to_ns": toNS, "await_consumer": awaitConsumer})
	return emitted, r.persistAdvanceBoundary()
}

func (r *Run) waitForConsumer(ctx context.Context, toNS int64) error {
	if err := r.awaitQuiescence(ctx, toNS); err != nil {

		if errors.Is(err, ErrConsumerNotQuiesced) {
			r.commandMu.Lock()
			r.fail(err)
			r.commandMu.Unlock()
		}
		return err
	}
	return nil
}

func (r *Run) advanceWorld(toNS int64) (int, error) {
	if r.runErr != nil {
		return 0, r.runErr
	}
	emitted, _, err := r.World.Advance(toNS)
	if err != nil {
		return 0, fmt.Errorf("run: advance: %w", err)
	}
	r.worldEndTimeNS = toNS
	return emitted, nil
}

func (r *Run) flushPendingDeliveries(toNS int64) {

	for _, delivery := range r.Perturb.Flush(toNS) {
		r.deliver(delivery)
	}
}

func (r *Run) persistAdvanceBoundary() error {

	if err := r.flushDurable(); err != nil {
		r.fail(err)
		return err
	}
	return nil
}
