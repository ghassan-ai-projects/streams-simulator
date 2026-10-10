package domain

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func instrument(ev Evidence) InstrumentMetrics {
	return instrumentEvidence(ev.Ledger, ev.Calls, ev.Perturbations, ev.Emitted)
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
