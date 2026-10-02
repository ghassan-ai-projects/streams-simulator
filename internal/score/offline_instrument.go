package score

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

func instrumentFrom(ledger []model.LedgerRecord, calls []world.EffectorCall, perturbations []string) InstrumentMetrics {
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
	// Perturbation fidelity: every applied perturbation left a mark in the
	// ledger — the same rule as the online path.
	m.PerturbationFidelity = true
	for _, name := range perturbations {
		found := false
		for _, l := range ledger {
			if reasonOf(name) == l.DeliveryReason && l.Delivered != (name == "drop") {
				found = true
				break
			}
		}
		if !found {
			m.PerturbationFidelity = false
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
