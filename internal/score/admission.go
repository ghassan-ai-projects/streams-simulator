package score

import "github.com/ghassan-ai-projects/streams-simulator/internal/model"

func admissionMetrics(admission []model.Admission, ledger []model.LedgerRecord, perturbations []string) ConsumerMetrics {
	m := ConsumerMetrics{}
	admissionBySeq := map[int64][]string{}
	for _, a := range admission {
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
	m.IdentityConflict = !hasPerturbation(perturbations, "id_reuse") || reportsOutcome(admissionBySeq, model.AdmissionConflict)
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
	return m
}

func reportsOutcome(admission map[int64][]string, wanted string) bool {
	for _, outcomes := range admission {
		if hasOutcome(outcomes, wanted) {
			return true
		}
	}
	return false
}

func hasPerturbation(perturbations []string, wanted string) bool {
	for _, name := range perturbations {
		if name == wanted {
			return true
		}
	}
	return false
}
