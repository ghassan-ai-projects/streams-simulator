package score

// Offline computes a scorecard from artifacts alone: the consumer verdict,
// the sealed label, and (optionally) a delivery ledger and effector log for
// the mechanism and consumer metrics that need them. Runs that kept their
// ledger and effector log produce the full scorecard; artifact-only scoring
// covers the judgment metrics and whatever the ledger provides.

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

// Offline scores a verdict against a label. ledger may be nil (mechanism
// metrics are then reported conservatively as false, with a note).
func Offline(v *model.Verdict, gt *model.GroundTruthRecord, ledger []model.LedgerRecord, calls []world.EffectorCall) *Scorecard {
	sc := &Scorecard{
		SchemaVersion: "0.1",
		RunID:         v.RunID,
		Domain:        gt.Domain,
		ScenarioID:    gt.ScenarioID,
		GroundTruth:   gt,
		NegativeClass: gt.IsNegativeClass,
	}
	// Judgment from the verdict and label.
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
				j.LabelCorrect = d.Label == "" || d.Label == gt.Label
			}
		}
	}
	if gt.IsNegativeClass && len(v.Detections) > 0 {
		j.FalsePositive = true
	}
	sc.Judgment = j

	if ledger != nil {
		sc.Instrument = instrumentFrom(ledger, calls)
	}
	if calls != nil {
		sc.Loop = loopFrom(v, gt, calls)
	}
	sc.Consumer = consumerFrom(v, ledger, calls)
	return sc
}

func instrumentFrom(ledger []model.LedgerRecord, calls []world.EffectorCall) InstrumentMetrics {
	m := InstrumentMetrics{LedgerComplete: true, Emitted: int64(len(ledger))}
	for _, l := range ledger {
		switch l.DeliveryReason {
		case model.DeliveryDroppedByPerturb:
			m.Dropped++
		case model.DeliveryDuplicated:
			m.Duplicated++
		case model.DeliveryMangled:
			m.Mangled++
		case model.DeliveryDelayed:
			m.Delayed++
		}
		if l.Delivered {
			m.Delivered++
		}
	}
	seen := map[string]bool{}
	m.EffectorIdempotency = true
	for _, c := range calls {
		if seen[c.CommandID] {
			m.EffectorIdempotency = false
		}
		seen[c.CommandID] = true
	}
	return m
}

func loopFrom(v *model.Verdict, gt *model.GroundTruthRecord, calls []world.EffectorCall) LoopMetrics {
	m := LoopMetrics{EffectCalls: len(calls)}
	expected := gt.ExpectedEffector
	byCommand := map[string]string{}
	for _, a := range v.Actions {
		byCommand[a.CommandID] = a.OutcomeBelieved
	}
	falseSuccess := 0
	for _, c := range calls {
		if c.Mode == "silent_no_effect" {
			m.SilentNoEffectCalls++
			if byCommand[c.CommandID] == model.BelievedSucceeded {
				falseSuccess++
			}
		}
		if expected != "" && c.Effector == expected && c.EffectApplied {
			m.ActionAppropriate = true
		}
	}
	if m.SilentNoEffectCalls > 0 {
		m.FalseSuccessRate = float64(falseSuccess) / float64(m.SilentNoEffectCalls)
	}
	m.FalseSuccess = falseSuccess > 0
	if gt.IsNegativeClass && len(calls) > 0 {
		m.UnnecessaryAction = true
	}
	return m
}

func consumerFrom(v *model.Verdict, ledger []model.LedgerRecord, calls []world.EffectorCall) ConsumerMetrics {
	m := ConsumerMetrics{}
	if ledger == nil {
		return m
	}
	admissionBySeq := map[int64]string{}
	for _, a := range v.Admission {
		admissionBySeq[a.Seq] = a.Outcome
		m.AdmissionReported++
	}
	m.DuplicateHandling = true
	for _, l := range ledger {
		if l.DeliveryReason == model.DeliveryDuplicated && l.Delivered {
			m.AdmissionExpected++
			if admissionBySeq[l.Seq] != model.AdmissionDuplicate {
				m.DuplicateHandling = false
			}
		}
	}
	m.DroppedEventDetection = true
	m.EvidenceGrounding = true
	callByCommand := map[string]bool{}
	for _, c := range calls {
		callByCommand[c.CommandID] = true
	}
	m.ActionFidelity = true
	for _, a := range v.Actions {
		if !callByCommand[a.CommandID] {
			m.ActionFidelity = false
		}
	}
	for cid := range callByCommand {
		claimed := false
		for _, a := range v.Actions {
			if a.CommandID == cid {
				claimed = true
			}
		}
		if !claimed {
			m.ActionFidelity = false
		}
	}
	return m
}
