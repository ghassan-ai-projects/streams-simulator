package score

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

func instrumentFrom(ledger []model.LedgerRecord, calls []world.EffectorCall, perturbations []string) InstrumentMetrics {
	var emitted int64
	for _, row := range ledger {
		if row.Seq >= 0 {
			emitted = max(emitted, row.Seq+1)
		}
	}
	return instrumentEvidence(ledger, calls, perturbations, emitted)
}

func instrumentEvidence(ledger []model.LedgerRecord, calls []world.EffectorCall, perturbations []string, emitted int64) InstrumentMetrics {
	m := ledgerMetrics(ledger, emitted)
	m.PerturbationFidelity = perturbationsObserved(ledger, perturbations)
	m.EffectorIdempotency = commandsAppliedOnce(calls)
	return m
}

func perturbationsObserved(ledger []model.LedgerRecord, perturbations []string) bool {
	// Each applied perturbation must leave its expected delivery mark.
	for _, name := range perturbations {
		if !ledgerShowsPerturbation(ledger, name) {
			return false
		}
	}
	return true
}

func commandsAppliedOnce(calls []world.EffectorCall) bool {
	m := InstrumentMetrics{}
	seen := map[string]bool{}
	m.EffectorIdempotency = true
	for _, c := range calls {
		if seen[c.CommandID] {
			m.EffectorIdempotency = false
		}
		seen[c.CommandID] = true
	}
	return m.EffectorIdempotency
}

func ledgerShowsPerturbation(ledger []model.LedgerRecord, name string) bool {
	for _, row := range ledger {
		if reasonOf(name) == row.DeliveryReason && row.Delivered != (name == "drop") {
			return true
		}
	}
	return false
}
