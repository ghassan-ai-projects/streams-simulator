package score

// Offline computes a scorecard from artifacts alone: the consumer verdict,
// the sealed label, and (optionally) a delivery ledger and effector log for
// the mechanism and consumer metrics that need them. Runs that kept their
// ledger and effector log produce the full scorecard; artifact-only scoring
// covers the judgment metrics and whatever the ledger provides.

import (
	"fmt"

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
	m := InstrumentMetrics{LedgerComplete: true}
	seenIDs := map[uint64]bool{}
	seenSeq := map[int64]bool{}
	for _, l := range ledger {
		if l.DeliveryID == 0 || seenIDs[l.DeliveryID] {
			m.LedgerComplete = false
		}
		seenIDs[l.DeliveryID] = true
		seenSeq[l.Seq] = true
		if l.Seq >= 0 {
			m.Emitted = max64(m.Emitted, l.Seq+1)
		}
		switch l.DeliveryReason {
		case model.DeliveryDroppedByPerturb:
			m.Dropped++
		case model.DeliveryDuplicated:
			m.Duplicated++
		case model.DeliveryMangled:
			m.Mangled++
		case model.DeliveryDelayed:
			m.Delayed++
		case model.DeliveryRewritten, model.DeliveryReordered, model.DeliveryOmitted:
			// Perturbed is counted below; these are still valid terminal rows.
		}
		if l.Delivered {
			m.Delivered++
		}
	}
	for seq := int64(0); seq < m.Emitted; seq++ {
		if !seenSeq[seq] {
			m.LedgerComplete = false
			break
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
	admissionBySeq := map[int64][]string{}
	for _, a := range v.Admission {
		admissionBySeq[a.Seq] = append(admissionBySeq[a.Seq], a.Outcome)
		m.AdmissionReported++
	}
	m.DuplicateHandling = true
	duplicateSeqs := map[int64]bool{}
	for _, l := range ledger {
		if l.DeliveryReason == model.DeliveryDuplicated && l.Delivered {
			duplicateSeqs[l.Seq] = true
			if !hasOutcome(admissionBySeq[l.Seq], model.AdmissionDuplicate) {
				m.DuplicateHandling = false
			}
		}
	}
	m.AdmissionExpected = len(duplicateSeqs)
	m.IdentityConflict = true
	for _, outcomes := range admissionBySeq {
		if hasOutcome(outcomes, model.AdmissionConflict) {
			m.IdentityConflict = true
			break
		}
	}
	m.LatenessClassification = true
	for _, l := range ledger {
		if l.DeliveryReason == model.DeliveryDelayed && l.Delivered && !hasOutcome(admissionBySeq[l.Seq], model.AdmissionLate) {
			m.LatenessClassification = false
		}
	}
	drops := 0
	detectionTimes := map[string][]int64{}
	for _, d := range v.Detections {
		if t, err := model.ParseTime(d.DetectedAt); err == nil {
			detectionTimes[d.EntityID] = append(detectionTimes[d.EntityID], t)
		}
	}
	for _, l := range ledger {
		if l.DeliveryReason == model.DeliveryDroppedByPerturb && !l.Delivered {
			drops++
		}
	}
	m.DroppedEventDetection = drops == 0
	usedDetections := map[string]map[int]bool{}
	if drops > 0 {
		m.DroppedEventDetection = true
		for _, l := range ledger {
			if l.DeliveryReason != model.DeliveryDroppedByPerturb || l.Delivered {
				continue
			}
			if usedDetections[l.EntityID] == nil {
				usedDetections[l.EntityID] = map[int]bool{}
			}
			matched := false
			for i, t := range detectionTimes[l.EntityID] {
				if !usedDetections[l.EntityID][i] && t >= l.ObservedTimeNS-30*60*1e9 && t <= l.ObservedTimeNS+30*60*1e9 {
					usedDetections[l.EntityID][i] = true
					matched = true
					break
				}
			}
			if !matched {
				m.DroppedEventDetection = false
			}
		}
	}
	m.EvidenceGrounding = true
	deliveredSeqs := map[int64]bool{}
	for _, l := range ledger {
		if l.Delivered {
			deliveredSeqs[l.Seq] = true
		}
	}
	for _, d := range v.Detections {
		for _, ref := range d.EvidenceRefs {
			var seq int64
			if _, err := fmt.Sscanf(ref, "seq:%d", &seq); err != nil || !deliveredSeqs[seq] {
				m.EvidenceGrounding = false
			}
		}
	}
	m.ActionFidelity = true
	used := map[int]bool{}
	for _, c := range calls {
		matched := false
		for i, a := range v.Actions {
			if used[i] || a.CommandID != c.CommandID || a.Effector != c.Effector || a.EntityID != c.EntityID {
				continue
			}
			issued, err := model.ParseTime(a.IssuedAt)
			if err != nil || issued < c.AtNS || (c.Accepted && a.OutcomeBelieved == model.BelievedFailed) || (!c.Accepted && a.OutcomeBelieved == model.BelievedSucceeded) {
				continue
			}
			used[i] = true
			matched = true
			break
		}
		if !matched {
			m.ActionFidelity = false
		}
	}
	if len(used) != len(v.Actions) {
		m.ActionFidelity = false
	}
	return m
}

func max64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
