// Package durable is the file-system edge of the run module: the append-only
// delivery ledger file, the published evidence files and the artifact reader.
//
//nolint:staticcheck // ST1005: "End" names the public Run.End operation these messages report on; the text is part of the error contract
package durable

import (
	"encoding/json"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/kernel"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

type Ledger struct {
	log *kernel.Log
}

func OpenLedger(path string) (*Ledger, error) {
	log, err := kernel.NewLog(path, false)
	if err != nil {
		return nil, fmt.Errorf("run: open ledger %s: %w", path, err)
	}
	return &Ledger{log: log}, nil
}

func (l *Ledger) Append(rec model.LedgerRecord) {
	if raw, err := json.Marshal(rec); err == nil {
		_ = l.log.Write(append(raw, '\n'))
	}
}

func (l *Ledger) Flush() error {
	if err := l.log.Flush(); err != nil {
		return fmt.Errorf("run: flush ledger: %w", err)
	}
	return nil
}

func (l *Ledger) Finish() error {
	if err := l.log.Close(); err != nil {
		return fmt.Errorf("End: close ledger: %w", err)
	}
	return nil
}
