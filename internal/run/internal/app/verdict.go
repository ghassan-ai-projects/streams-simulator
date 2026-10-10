package app

import (
	"fmt"
	rules "github.com/ghassan-ai-projects/streams-simulator/internal/run/internal/domain"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run/internal/clock"
)

// SubmitVerdict stores and validates a consumer verdict.
func (r *Run) SubmitVerdict(v *model.Verdict) error {
	r.commandMu.Lock()
	defer r.commandMu.Unlock()
	if err := r.admitVerdict(v); err != nil {
		return err
	}
	if err := rules.ValidateSubmittedVerdict(v); err != nil {
		return err
	}
	r.verdict = rules.CloneVerdict(v)
	return nil
}

// Verdict returns the submitted verdict, or nil.
func (r *Run) Verdict() *model.Verdict { return rules.CloneVerdict(r.verdict) }

// Ledger returns the delivery ledger in delivery order.
func (r *Run) Ledger() []model.LedgerRecord {
	return append([]model.LedgerRecord(nil), r.ledger...)
}

// TraceDigest is the sha256 of the delivered trace bytes.
func (r *Run) TraceDigest() string { return r.traceDigest }

// Unblind permanently stamps the run and excludes it from scorecards.
func (r *Run) Unblind() {
	r.unblinded = true
	r.unblindedAt = clock.Stamp()
}

// Unblinded reports the stamp state.
func (r *Run) Unblinded() (bool, string) { return r.unblinded, r.unblindedAt }

// Reproducible reports whether the run is hash-reproducible.
func (r *Run) Reproducible() bool { return r.reproducible }

func (r *Run) admitVerdict(v *model.Verdict) error {
	// The operator endpoint outlives End; late verdicts cannot alter published artifacts.
	if r.finished {
		return fmt.Errorf("SubmitVerdict: run is finished")
	}
	if v == nil {
		return fmt.Errorf("SubmitVerdict: verdict is required")
	}
	if v.RunID != r.ID {
		return fmt.Errorf("run: verdict run_id %q does not match run %q", v.RunID, r.ID)
	}
	return nil
}
