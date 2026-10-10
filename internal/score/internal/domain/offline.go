package domain

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

// Offline scores a verdict against a label. ledger may be nil (mechanism
// metrics are then reported conservatively as false, with a note).
// perturbations is the run's applied-perturbation list (from the artifact),
// needed for the clock-skew metric exactly as the online path uses it.
func Offline(v *model.Verdict, gt *model.GroundTruthRecord, ledger []model.LedgerRecord, calls []world.EffectorCall, perturbations []string) *Scorecard {
	sc := newScorecard(v.RunID, gt)
	sc.Judgment = judgmentFrom(v, gt)
	if ledger != nil {
		sc.Instrument = instrumentFrom(ledger, calls, perturbations)
	}
	if calls != nil {
		sc.Loop = loopFrom(v, gt, calls)
	}
	sc.Consumer = consumerFrom(v, ledger, calls, perturbations)
	return sc
}
