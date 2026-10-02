package run

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
	if r.runErr != nil {
		r.commandMu.Unlock()
		return 0, r.runErr
	}
	emitted, effects, err := r.World.Advance(toNS)
	if err != nil {
		r.commandMu.Unlock()
		return 0, fmt.Errorf("run: advance: %w", err)
	}
	r.worldEndTimeNS = toNS
	// Flush windowing perturbations (reorder, flaps, backfill) at the
	// boundary.
	for _, d := range r.Perturb.Flush(toNS) {
		r.deliver(d)
	}
	if r.runErr != nil {
		r.commandMu.Unlock()
		return emitted, r.runErr
	}
	// Log the advance before awaiting quiescence: the world has already
	// moved, and replay must reproduce exactly this state even when the
	// consumer never reports quiescence.
	r.commandLog = append(r.commandLog, model.Command{
		Seq: int64(len(r.commandLog)), AtNS: r.World.Clock(),
		Op: model.OpClockAdvance, Args: map[string]any{"to_ns": toNS, "await_consumer": awaitConsumer},
	})
	// Command boundary: the ledger rows and trace bytes for this advance are
	// now on file descriptors, so a crash here loses nothing acknowledged.
	if err := r.flushDurable(); err != nil {
		r.fail(err)
		r.commandMu.Unlock()
		return emitted, err
	}
	if !awaitConsumer {
		r.commandMu.Unlock()
		_ = effects
		return emitted, nil
	}
	// The quiescence wait is a consumer-sync barrier, not a world mutation:
	// release the command mutex so the consumer's effector call can land
	// while the advance waits. Without this the closed loop deadlocks.
	r.commandMu.Unlock()
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
		return emitted, err
	}
	_ = effects
	return emitted, nil
}
