package run

import (
	"encoding/json"
	"fmt"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"time"
)

// SubmitVerdict stores and validates a consumer verdict.
func (r *Run) SubmitVerdict(v *model.Verdict) error {
	r.commandMu.Lock()
	defer r.commandMu.Unlock()
	if r.finished {
		// The operator endpoint outlives run.end; a verdict arriving then
		// must not be silently dropped from an already-written artifact.
		return fmt.Errorf("SubmitVerdict: run is finished")
	}
	if v == nil {
		return fmt.Errorf("SubmitVerdict: verdict is required")
	}
	if v.RunID != r.ID {
		return fmt.Errorf("run: verdict run_id %q does not match run %q", v.RunID, r.ID)
	}
	raw, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("SubmitVerdict: %w", err)
	}
	if err := model.ValidateVerdict(raw); err != nil {
		return fmt.Errorf("SubmitVerdict: %w", err)
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
