package app

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// SubmitVerdict stores and validates a consumer verdict.
func (r *Run) SubmitVerdict(v *model.Verdict) error {
	r.commandMu.Lock()
	defer r.commandMu.Unlock()
	if err := r.admitVerdict(v); err != nil {
		return err
	}
	if err := validateSubmittedVerdict(v); err != nil {
		return err
	}
	r.verdict = cloneVerdict(v)
	return nil
}

// Verdict returns the submitted verdict, or nil.
func (r *Run) Verdict() *model.Verdict { return cloneVerdict(r.verdict) }

// Ledger returns the delivery ledger in delivery order.
func (r *Run) Ledger() []model.LedgerRecord {
	return append([]model.LedgerRecord(nil), r.ledger...)
}

// TraceDigest is the sha256 of the delivered trace bytes.
func (r *Run) TraceDigest() string { return r.traceDigest }

// Unblind permanently stamps the run and excludes it from scorecards.
func (r *Run) Unblind() {
	r.unblinded = true
	r.unblindedAt = time.Now().UTC().Format(time.RFC3339Nano)
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

func validateSubmittedVerdict(v *model.Verdict) error {
	raw, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("SubmitVerdict: %w", err)
	}
	if err := model.ValidateVerdict(raw); err != nil {
		return fmt.Errorf("SubmitVerdict: %w", err)
	}
	return nil
}
