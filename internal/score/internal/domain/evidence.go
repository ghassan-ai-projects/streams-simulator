package domain

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func groundsEvidence(detections []model.Detection, ledger []model.LedgerRecord) bool {
	delivered := deliveredSequences(ledger)
	for _, detection := range detections {
		for _, ref := range detection.EvidenceRefs {
			if !referencesDeliveredSequence(ref, delivered) {
				return false
			}
		}
	}
	return true
}

func deliveredSequences(ledger []model.LedgerRecord) map[int64]bool {
	delivered := map[int64]bool{}
	for _, row := range ledger {
		if row.Delivered {
			delivered[row.Seq] = true
		}
	}
	return delivered
}

func referencesDeliveredSequence(ref string, delivered map[int64]bool) bool {
	var seq int64
	_, err := fmt.Sscanf(ref, "seq:%d", &seq)
	return err == nil && delivered[seq]
}
