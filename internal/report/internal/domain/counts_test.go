package domain

import (
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestCountsSplitLedgerRowsByDeliveryAndReason(t *testing.T) {
	t.Parallel()
	ledger := doc(
		`{"delivery_id":1,"delivered":true,"delivery_reason":"ok"}`,
		`{"delivery_id":2,"delivered":true,"delivery_reason":"duplicated"}`,
		`{"delivery_id":3,"delivered":false,"delivery_reason":"dropped_by_perturbation"}`,
		`{"delivery_id":4,"delivered":false,"delivery_reason":"dropped_by_perturbation"}`,
	)
	art := &model.RunArtifact{Counts: model.Counts{Emitted: 3}}
	in := streams{}
	rows, err := readLedger(ledger)
	if err != nil {
		t.Fatal(err)
	}
	in.ledger = rows
	counts := countsOf(Folder{Artifact: art, Ledger: ledger}, in)
	if *counts.Emitted != 3 || *counts.LedgerRows != 4 || *counts.Delivered != 2 || *counts.Dropped != 2 ||
		counts.DeliveredByReason["duplicated"] != 1 || counts.DroppedByReason["dropped_by_perturbation"] != 2 {
		t.Fatalf("counts = %+v, want 3 emitted, 4 rows, 2 delivered, 2 dropped by reason", counts)
	}
}

func TestCountsOmitWhatTheFolderDoesNotHold(t *testing.T) {
	t.Parallel()
	if counts := countsOf(Folder{}, streams{}); counts.Emitted != nil || counts.Observed != nil || counts.Delivered != nil {
		t.Fatalf("counts = %+v, want none", counts)
	}
}
