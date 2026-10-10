package domain

import "github.com/ghassan-ai-projects/streams-simulator/internal/model"

func admissionMetrics(admission []model.Admission, ledger []model.LedgerRecord, perturbations []string) ConsumerMetrics {
	bySeq := admissionOutcomes(admission)
	duplicates, handled := duplicateAdmissions(ledger, bySeq)
	return ConsumerMetrics{
		AdmissionReported: len(admission), AdmissionExpected: duplicates,
		DuplicateHandling:      handled,
		IdentityConflict:       !hasPerturbation(perturbations, "id_reuse") || reportsOutcome(bySeq, model.AdmissionConflict),
		LatenessClassification: classifiesLateRecords(ledger, bySeq),
		ClockSkewRejection:     rejectsClockSkew(bySeq, perturbations),
	}
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

func admissionOutcomes(admission []model.Admission) map[int64][]string {
	bySeq := map[int64][]string{}
	for _, a := range admission {
		bySeq[a.Seq] = append(bySeq[a.Seq], a.Outcome)
	}
	return bySeq
}

func duplicateAdmissions(ledger []model.LedgerRecord, bySeq map[int64][]string) (int, bool) {
	handled := true
	sequences := map[int64]bool{}
	for _, row := range ledger {
		if row.DeliveryReason == model.DeliveryDuplicated && row.Delivered {
			sequences[row.Seq] = true
			handled = handled && hasOutcome(bySeq[row.Seq], model.AdmissionDuplicate)
		}
	}
	return len(sequences), handled
}

func classifiesLateRecords(ledger []model.LedgerRecord, bySeq map[int64][]string) bool {
	for _, row := range ledger {
		if row.DeliveryReason == model.DeliveryDelayed && row.Delivered && !hasOutcome(bySeq[row.Seq], model.AdmissionLate) {
			return false
		}
	}
	return true
}

func rejectsClockSkew(bySeq map[int64][]string, perturbations []string) bool {
	// Skewed records must be rejected or classified malformed/out-of-contract.
	if !hasPerturbation(perturbations, "clock_skew") {
		return true
	}
	for _, outcomes := range bySeq {
		if hasOutcome(outcomes, model.AdmissionRejected) || hasOutcome(outcomes, model.AdmissionMalformed) || hasOutcome(outcomes, model.AdmissionOutOfContract) {
			return true
		}
	}
	return false
}
