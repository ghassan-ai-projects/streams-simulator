package app

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// artifact assembles the run artifact from the run state.
func (r *Run) artifact() *model.RunArtifact {
	commands := r.sortedCommandLog()
	counts := r.artifactCounts()
	artifact := r.artifactIdentity()
	r.attachArtifactInputs(artifact)
	r.attachArtifactState(artifact, commands, counts)
	return artifact
}

func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func countLedger(ledger []model.LedgerRecord, reasons ...string) int {
	n := 0
	for _, l := range ledger {
		for _, r := range reasons {
			if l.DeliveryReason == r {
				n++
				break
			}
		}
	}
	return n
}

func writeJSONL(path string, v any) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	defer func() { _ = file.Close() }()
	switch records := v.(type) {
	case []model.LedgerRecord:
		return writeJSONRecords(file, records)
	case []model.StateSnapshot:
		return writeJSONRecords(file, records)
	}
	return nil
}

func writeJSONRecords[T any](file *os.File, records []T) error {
	for _, record := range records {
		raw, _ := json.Marshal(record)
		if _, err := file.Write(append(raw, '\n')); err != nil {
			return fmt.Errorf("streamsim: %w", err)
		}
	}
	return nil
}
