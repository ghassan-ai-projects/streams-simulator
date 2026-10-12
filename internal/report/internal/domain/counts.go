package domain

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/journal"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func countsOf(folder Folder, in streams) Counts {
	var counts Counts
	if folder.Artifact != nil {
		emitted := folder.Artifact.Counts.Emitted
		counts.Emitted = &emitted
	}
	if folder.Observer.Present {
		observed := countType(in.observer, journal.TypeObservation)
		counts.Observed = &observed
	}
	if folder.Ledger.Present {
		countLedger(&counts, in.ledger)
	}
	return counts
}

func countType(records []journal.Record, recordType string) int64 {
	var n int64
	for _, record := range records {
		if record.Type == recordType {
			n++
		}
	}
	return n
}

func countLedger(counts *Counts, rows []model.LedgerRecord) {
	total := int64(len(rows))
	counts.LedgerRows = &total
	counts.Delivered, counts.Dropped = new(int64), new(int64)
	counts.DeliveredByReason, counts.DroppedByReason = map[string]int64{}, map[string]int64{}
	for _, row := range rows {
		tally(counts, row)
	}
}

func tally(counts *Counts, row model.LedgerRecord) {
	byReason, total := counts.DroppedByReason, counts.Dropped
	if row.Delivered {
		byReason, total = counts.DeliveredByReason, counts.Delivered
	}
	byReason[row.DeliveryReason]++
	*total++
}

func readLedger(doc Document) ([]model.LedgerRecord, error) {
	var rows []model.LedgerRecord
	for index, line := range bytes.Split(doc.Raw, []byte{'\n'}) {
		if len(line) == 0 {
			continue
		}
		var row model.LedgerRecord
		if err := json.Unmarshal(line, &row); err != nil {
			return nil, fmt.Errorf("report: ledger.jsonl line %d: %w", index+1, err)
		}
		rows = append(rows, row)
	}
	return rows, nil
}
