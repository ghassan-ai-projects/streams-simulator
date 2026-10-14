package domain

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/journal"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func (truth *Truth) readFaults(director []journal.Record) error {
	truth.Faults = []model.FaultInterval{}
	for _, record := range director {
		if record.Type != journal.TypeFaultTruth {
			continue
		}
		truth.Faults = []model.FaultInterval{}
		if err := convert(record.Fields["faults"], &truth.Faults); err != nil {
			return fmt.Errorf("report: sealed/journal.jsonl record %d: %w", record.Seq, err)
		}
	}
	return nil
}
