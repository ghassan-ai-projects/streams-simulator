package score

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func groundsEvidence(detections []model.Detection, ledger []model.LedgerRecord) bool {
	grounded := true
	deliveredSeqs := map[int64]bool{}
	for _, l := range ledger {
		if l.Delivered {
			deliveredSeqs[l.Seq] = true
		}
	}
	for _, d := range detections {
		for _, ref := range d.EvidenceRefs {
			var seq int64
			if _, err := fmt.Sscanf(ref, "seq:%d", &seq); err != nil || !deliveredSeqs[seq] {
				grounded = false
			}
		}
	}
	return grounded
}
