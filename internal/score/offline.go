package score

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

// Offline scores a verdict against a label. ledger may be nil (mechanism
// metrics are then reported conservatively as false, with a note).
// perturbations is the run's applied-perturbation list (from the artifact),
// needed for the clock-skew metric exactly as the online path uses it.
func Offline(v *model.Verdict, gt *model.GroundTruthRecord, ledger []model.LedgerRecord, calls []world.EffectorCall, perturbations []string) *Scorecard {
	sc := &Scorecard{
		SchemaVersion: "0.1",
		Bundle:        scoringBundleVersion,
		RunID:         v.RunID,
		Domain:        gt.Domain,
		ScenarioID:    gt.ScenarioID,
		GroundTruth:   gt,
		NegativeClass: gt.IsNegativeClass,
	}
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
