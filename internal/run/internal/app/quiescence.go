package app

import (
	"context"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
)

// ReportQuiesced records the consumer's quiescence assertion. The watermark
// is monotonic, so a report can only move it forward: a stale report for an
// earlier instant can never satisfy a later await, and a report received
// between advances is honored on the next wait (the fast path).
func (r *Run) ReportQuiesced(throughNS int64) {
	r.quiesceMu.Lock()
	defer r.quiesceMu.Unlock()
	if throughNS > r.quiescedThroughNS {
		r.quiescedThroughNS = throughNS
		close(r.quiesceNotify)
		r.quiesceNotify = make(chan struct{})
	}
}

// SetQuiesceParkedHook installs a callback fired each time a quiescence wait
// is about to block (test harness; nil by default).
func (r *Run) SetQuiesceParkedHook(h func()) {
	r.quiesceParked = h
}

func (r *Run) awaitQuiescence(ctx context.Context, toNS int64) error {
	timer := r.Config.quiescenceClock.NewTimer(DefaultQuiescenceTimeout)
	defer timer.Stop()
	for {
		ch, through, done := r.quiescenceStatus(toNS)
		if done {
			return nil
		}
		r.notifyQuiescenceParked()
		if err := waitForQuiescence(ctx, timer, ch, through, toNS); err != nil {
			return err
		}
	}
}

// Domain exposes the compiled domain spec.
func (r *Run) Domain() *domain.Compiled { return r.Config.Domain }

// Digest is the world digest: over (sim_version, domain digest, seed,
// world config). Two creates with the same arguments must agree.
func (r *Run) Digest() string { return worldDigest(r) }

// SetFailureMode overrides the effector failure-mode distribution for
// subsequent invocations (test-only knob).
func (r *Run) SetFailureMode(mode string) {
	r.World.SetFailureMode(mode)
}

// UnblindedStamp reports whether the run was permanently stamped.
func (r *Run) UnblindedStamp() bool {
	r.commandMu.Lock()
	defer r.commandMu.Unlock()
	return r.unblinded
}

// AppliedPerturbations returns the perturbations applied during the run, in
// application order (the scorer's perturbation-fidelity input).
func (r *Run) AppliedPerturbations() []string {
	r.commandMu.Lock()
	defer r.commandMu.Unlock()
	return r.appliedPerturbations()
}

// appliedPerturbations copies the perturbation history; the caller holds the
// command lock.
func (r *Run) appliedPerturbations() []string {
	out := make([]string, len(r.perturbHistory))
	copy(out, r.perturbHistory)
	return out
}

func (r *Run) quiescenceStatus(toNS int64) (chan struct{}, int64, bool) {
	r.quiesceMu.Lock()
	defer r.quiesceMu.Unlock()
	if r.quiescedThroughNS >= toNS {
		return nil, r.quiescedThroughNS, true
	}
	return r.quiesceNotify, r.quiescedThroughNS, false
}

func (r *Run) notifyQuiescenceParked() {
	if r.quiesceParked != nil {
		r.quiesceParked()
	}
}

func waitForQuiescence(ctx context.Context, timer QuiescenceTimer, ch chan struct{}, through, toNS int64) error {
	select {
	case <-ch:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("run: quiescence wait canceled: %w", ctx.Err())
	case <-timer.C():
		return fmt.Errorf("run: %w: quiesced through %d, asked for %d", ErrConsumerNotQuiesced, through, toNS)
	}
}
