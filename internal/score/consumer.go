package score

import (
	"fmt"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

func consumer(r *run.Run, gt *model.GroundTruthRecord) ConsumerMetrics {
	v := r.Verdict()
	ledger := r.Ledger()
	m := ConsumerMetrics{}

	admissionBySeq := map[int64][]string{}
	for _, a := range v.Admission {
		admissionBySeq[a.Seq] = append(admissionBySeq[a.Seq], a.Outcome)
		m.AdmissionReported++
	}
	deliveredSeqs := map[int64]bool{}
	for _, l := range ledger {
		if l.Delivered {
			deliveredSeqs[l.Seq] = true
		}
	}

	// Duplicate handling: every injected duplicate reported as duplicate.
	m.DuplicateHandling = true
	m.AdmissionExpected = 0
	for _, l := range ledger {
		if l.DeliveryReason == model.DeliveryDuplicated && l.Delivered {
			if !hasOutcome(admissionBySeq[l.Seq], model.AdmissionDuplicate) {
				m.DuplicateHandling = false
			}
		}
	}
	duplicateSeqs := map[int64]bool{}
	for _, l := range ledger {
		if l.DeliveryReason == model.DeliveryDuplicated && l.Delivered {
			duplicateSeqs[l.Seq] = true
		}
	}
	m.AdmissionExpected = len(duplicateSeqs)
	// Identity conflict is applicable only when the identity-reuse
	// perturbation was actually applied; otherwise it is not a failing gate.
	m.IdentityConflict = true
	if hasPerturb(r, "id_reuse") {
		m.IdentityConflict = false
		for _, outcomes := range admissionBySeq {
			if hasOutcome(outcomes, model.AdmissionConflict) {
				m.IdentityConflict = true
				break
			}
		}
	}
	// Lateness: delayed records reported late.
	m.LatenessClassification = true
	for _, l := range ledger {
		if l.DeliveryReason == model.DeliveryDelayed && l.Delivered {
			if o, ok := admissionBySeq[l.Seq]; !ok || !hasOutcome(o, model.AdmissionLate) {
				m.LatenessClassification = false
			}
		}
	}
	// Dropped-event detection: every drop yields an absence detection or an
	// explicit admission gap.
	drops := 0
	for _, l := range ledger {
		if l.DeliveryReason == model.DeliveryDroppedByPerturb && !l.Delivered {
			drops++
		}
	}
	if drops == 0 {
		m.DroppedEventDetection = true
	} else {
		detectionTimes := map[string][]int64{}
		for _, d := range v.Detections {
			t, err := model.ParseTime(d.DetectedAt)
			if err == nil {
				detectionTimes[d.EntityID] = append(detectionTimes[d.EntityID], t)
			}
		}
		m.DroppedEventDetection = false
		used := map[string]map[int]bool{}
		for _, l := range ledger {
			if l.DeliveryReason != model.DeliveryDroppedByPerturb || l.Delivered {
				continue
			}
			if used[l.EntityID] == nil {
				used[l.EntityID] = map[int]bool{}
			}
			for i, t := range detectionTimes[l.EntityID] {
				if used[l.EntityID][i] {
					continue
				}
				// A detection within 30 minutes of the drop window counts.
				if t >= l.ObservedTimeNS-30*60*1e9 && t <= l.ObservedTimeNS+30*60*1e9 {
					used[l.EntityID][i] = true
					break
				}
			}
		}
		matched := 0
		for _, indices := range used {
			matched += len(indices)
		}
		m.DroppedEventDetection = matched == drops
	}
	// Clock skew: skewed records must be explicitly rejected or classified
	// malformed/out-of-contract when the perturbation is active.
	m.ClockSkewRejection = true
	if hasPerturb(r, "clock_skew") {
		m.ClockSkewRejection = false
		for _, outcomes := range admissionBySeq {
			if hasOutcome(outcomes, model.AdmissionRejected) || hasOutcome(outcomes, model.AdmissionMalformed) || hasOutcome(outcomes, model.AdmissionOutOfContract) {
				m.ClockSkewRejection = true
				break
			}
		}
	}
	// Evidence grounding: every cited seq was actually delivered.
	m.EvidenceGrounding = true
	for _, d := range v.Detections {
		for _, ref := range d.EvidenceRefs {
			var seq int64
			if _, err := fmt.Sscanf(ref, "seq:%d", &seq); err == nil {
				if !deliveredSeqs[seq] {
					m.EvidenceGrounding = false
				}
			}
		}
	}
	// Action fidelity: claimed actions match the effector log, and vice
	// versa.
	m.ActionFidelity = true
	usedActions := map[int]bool{}
	for _, c := range r.World.EffectorCalls() {
		matched := false
		for i, a := range v.Actions {
			if usedActions[i] || a.CommandID != c.CommandID || a.Effector != c.Effector || a.EntityID != c.EntityID {
				continue
			}
			issued, err := model.ParseTime(a.IssuedAt)
			if err != nil || issued < c.AtNS {
				continue
			}
			if c.Accepted && a.OutcomeBelieved == model.BelievedFailed {
				continue
			}
			if !c.Accepted && a.OutcomeBelieved == model.BelievedSucceeded {
				continue
			}
			usedActions[i] = true
			matched = true
			break
		}
		if !matched {
			m.ActionFidelity = false
		}
	}
	if len(usedActions) != len(v.Actions) {
		m.ActionFidelity = false
	}
	// Interlock handling: a refusal is never retried with a new command_id
	// for the same effector/entity.
	m.InterlockHandling = true
	refused := map[string]bool{}
	for _, c := range r.World.EffectorCalls() {
		if c.InterlockRefused {
			refused[c.Effector+"/"+c.EntityID] = true
		}
	}
	for _, c := range r.World.EffectorCalls() {
		if refused[c.Effector+"/"+c.EntityID] && !c.InterlockRefused {
			m.InterlockHandling = false
		}
	}
	return m
}
