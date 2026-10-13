package domain

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/journal"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func (truth *Truth) readKPIs(director []journal.Record) error {
	for _, record := range director {
		if record.Type != journal.TypeKPITruth {
			continue
		}
		truth.KPIs = &model.KPITruth{}
		if err := convert(record.Fields["kpis"], truth.KPIs); err != nil {
			return fmt.Errorf("report: sealed/journal.jsonl record %d: %w", record.Seq, err)
		}
	}
	return nil
}
