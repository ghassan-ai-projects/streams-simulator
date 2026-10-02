package score

import (
	"fmt"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

func consumerFrom(v *model.Verdict, ledger []model.LedgerRecord, calls []world.EffectorCall, perturbations []string) ConsumerMetrics {
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
	// Clock skew: skewed records must be explicitly rejected or classified
	// malformed/out-of-contract when the perturbation is active — the same
	// rule as the online path, fed by the artifact's perturbation list.
	m.ClockSkewRejection = true
	if hasPerturbation(perturbations, "clock_skew") {
		m.ClockSkewRejection = false
		for _, outcomes := range admissionBySeq {
			if hasOutcome(outcomes, model.AdmissionRejected) || hasOutcome(outcomes, model.AdmissionMalformed) || hasOutcome(outcomes, model.AdmissionOutOfContract) {
				m.ClockSkewRejection = true
				break
			}
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
	// Interlock handling: a refusal is never retried with a new command_id
	// for the same effector/entity — the same rule as the online path.
	m.InterlockHandling = true
	refused := map[string]bool{}
	for _, c := range calls {
		if c.InterlockRefused {
			refused[c.Effector+"/"+c.EntityID] = true
		}
	}
	for _, c := range calls {
		if refused[c.Effector+"/"+c.EntityID] && !c.InterlockRefused {
			m.InterlockHandling = false
		}
	}
	return m
}

func max64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

// hasPerturbation reports whether the run applied the named perturbation.
func hasPerturbation(perturbations []string, wanted string) bool {
	for _, p := range perturbations {
		if p == wanted {
			return true
		}
	}
	return false
}
