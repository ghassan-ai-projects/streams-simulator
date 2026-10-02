package score

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

func instrument(r *run.Run) InstrumentMetrics {
	ledger := r.Ledger()
	m := InstrumentMetrics{LedgerComplete: true, Emitted: r.World.EmittedCount()}
	seenIDs := map[uint64]bool{}
	seenSeq := map[int64]bool{}
	for _, l := range ledger {
		if l.DeliveryID == 0 || seenIDs[l.DeliveryID] {
			m.LedgerComplete = false
		}
		seenIDs[l.DeliveryID] = true
		seenSeq[l.Seq] = true
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
	for seq := int64(0); seq < m.Emitted; seq++ {
		if !seenSeq[seq] {
			m.LedgerComplete = false
			break
		}
	}
	// Perturbation fidelity: every perturbation that was applied left a mark
	// in the ledger (and never in the delivered event, which is enforced by
	// construction).
	m.PerturbationFidelity = true
	for _, name := range r.AppliedPerturbations() {
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
	// Effector idempotency: no command_id applied twice.
	seen := map[string]bool{}
	m.EffectorIdempotency = true
	for _, c := range r.World.EffectorCalls() {
		if seen[c.CommandID] {
			m.EffectorIdempotency = false
		}
		seen[c.CommandID] = true
	}
	return m
}

func reasonOf(perturbName string) string {
	switch perturbName {
	case "drop":
		return model.DeliveryDroppedByPerturb
	case "duplicate_burst", "storm", "id_reuse":
		return model.DeliveryDuplicated
	case "out_of_enum", "out_of_range", "unit_mismatch", "oversize", "malformed", "nan_inf":
		return model.DeliveryMangled
	case "delay_tail", "gross_backfill", "producer_flap":
		return model.DeliveryDelayed
	case "clock_skew", "non_monotonic", "time_encoding", "precision_edge", "injection_probe":
		return model.DeliveryRewritten
	case "reorder":
		return model.DeliveryReordered
	}
	return ""
}
