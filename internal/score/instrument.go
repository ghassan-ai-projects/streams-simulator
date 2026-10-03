package score

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

func instrument(r *run.Run) InstrumentMetrics {
	return instrumentEvidence(r.Ledger(), r.World.EffectorCalls(), r.AppliedPerturbations(), r.World.EmittedCount())
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
	}
	return timingDeliveryReason(perturbName)
}

func timingDeliveryReason(perturbName string) string {
	switch perturbName {
	case "clock_skew", "non_monotonic", "time_encoding", "precision_edge", "injection_probe":
		return model.DeliveryRewritten
	case "reorder":
		return model.DeliveryReordered
	}
	return ""
}
