package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// Advance moves the world to toNS, delivering everything along the way.
// awaitConsumer blocks until quiescence is reported through toNS; the wait
// is bounded by DefaultQuiescenceTimeout and canceled by ctx. On a
// quiescence timeout the world has already moved, so the advance is logged
// (replay reproduces the same state), the run is marked incomplete, and the
// emitted count is still returned alongside the error.
func (r *Run) Advance(ctx context.Context, toNS int64, awaitConsumer bool) (int, error) {
	r.commandMu.Lock()
	emitted, err := r.advanceThroughBoundary(toNS, awaitConsumer)
	r.commandMu.Unlock()
	if err != nil || !awaitConsumer {
		return emitted, err
	}
	if err := r.waitForConsumer(ctx, toNS); err != nil {
		return emitted, err
	}
	return emitted, nil
}

// advanceThroughBoundary holds commandMu through delivery and durable logging.
func (r *Run) advanceThroughBoundary(toNS int64, awaitConsumer bool) (int, error) {
	emitted, err := r.advanceWorld(toNS)
	if err != nil {
		return 0, err
	}
	r.flushPendingDeliveries(toNS)
	if r.runErr != nil {
		return emitted, r.runErr
	}
	// Log the moved world before quiescence, even when the consumer never reports.
	r.recordWorldCommand(model.OpClockAdvance, map[string]any{"to_ns": toNS, "await_consumer": awaitConsumer})
	return emitted, r.persistAdvanceBoundary()
}

// Consumer synchronization releases commandMu so closed-loop effects can land.
func (r *Run) waitForConsumer(ctx context.Context, toNS int64) error {
	if err := r.awaitQuiescence(ctx, toNS); err != nil {
		// Only a timeout is a simulator failure: a canceled wait is a
		// caller-side abandonment and leaves the run open-loop. The failure
		// state is written under the command mutex so a concurrent End or
		// Score never reads it half-written.
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
	// Windowing perturbations release their records at the advance boundary.
	for _, delivery := range r.Perturb.Flush(toNS) {
		r.deliver(delivery)
	}
}

func (r *Run) persistAdvanceBoundary() error {
	// Trace and ledger bytes must reach file descriptors before acknowledgement.
	if err := r.flushDurable(); err != nil {
		r.fail(err)
		return err
	}
	return nil
}
