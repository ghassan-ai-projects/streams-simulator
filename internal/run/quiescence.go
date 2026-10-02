package run

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
	timer := r.Config.QuiescenceClock.NewTimer(DefaultQuiescenceTimeout)
	defer timer.Stop()
	for {
		r.quiesceMu.Lock()
		if r.quiescedThroughNS >= toNS {
			r.quiesceMu.Unlock()
			return nil
		}
		ch := r.quiesceNotify
		through := r.quiescedThroughNS
		r.quiesceMu.Unlock()
		if r.quiesceParked != nil {
			r.quiesceParked()
		}
		select {
		case <-ch:
			continue
		case <-ctx.Done():
			return fmt.Errorf("run: quiescence wait canceled: %w", ctx.Err())
		case <-timer.C():
			return fmt.Errorf("run: %w: quiesced through %d, asked for %d", ErrConsumerNotQuiesced, through, toNS)
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
func (r *Run) UnblindedStamp() bool { return r.unblinded }

// AppliedPerturbations returns the perturbations applied during the run, in
// application order (the scorer's perturbation-fidelity input).
func (r *Run) AppliedPerturbations() []string {
	out := make([]string, len(r.perturbHistory))
	copy(out, r.perturbHistory)
	return out
}
