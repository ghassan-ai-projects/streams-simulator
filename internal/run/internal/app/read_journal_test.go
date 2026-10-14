package app

import (
	"bytes"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/journal"
)

func decodeJournalBytes(raw []byte) ([]journal.Record, error) {
	lines := bytes.Split(raw, []byte{'\n'})
	var records []journal.Record
	for index, line := range lines[:len(lines)-1] {
		record, err := journal.DecodeLine(line, index+1)
		if err != nil {
			return records, fmt.Errorf("%w", err)
		}
		records = append(records, record)
	}
	return records, nil
}
