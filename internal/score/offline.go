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
	// Judgment from the verdict and label. Label matching is strict, exactly
	// like the online path: an unlabelled detection is not a correct label.
	j := JudgmentMetrics{DetectionCount: len(v.Detections)}
	if gt.FirstObservableTimeNS > 0 {
		for _, d := range v.Detections {
			t, err := model.ParseTime(d.DetectedAt)
			if err != nil {
				continue
			}
			if d.EntityID != gt.EntityID {
				continue
			}
			if t < gt.FirstObservableTimeNS {
				j.Suspicious = true
				continue
			}
			if !j.Detected || t < j.DetectionLatencyNS {
				j.Detected = true
				j.DetectionLatencyNS = t - gt.FirstObservableTimeNS
				j.LabelCorrect = d.Label == gt.Label
			}
		}
	}
	if gt.IsNegativeClass && len(v.Detections) > 0 {
		j.FalsePositive = true
	}
	sc.Judgment = j

	if ledger != nil {
		sc.Instrument = instrumentFrom(ledger, calls, perturbations)
	}
	if calls != nil {
		sc.Loop = loopFrom(v, gt, calls)
	}
	sc.Consumer = consumerFrom(v, ledger, calls, perturbations)
	return sc
}
