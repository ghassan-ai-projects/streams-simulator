// Package durable is the file-system edge of the run module: the append-only
// delivery ledger file, the published evidence files and the artifact reader.
//
//nolint:staticcheck // ST1005: "End" names the public Run.End operation these messages report on; the text is part of the error contract
package durable

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// Ledger is an append-only JSONL delivery ledger on disk behind a buffered
// writer. It is flushed at command boundaries and synced when the run ends.
type Ledger struct {
	file   *os.File
	writer *bufio.Writer
}

// OpenLedger creates (or appends to) the ledger file at path, creating its
// directory.
func OpenLedger(path string) (*Ledger, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("run: ledger dir: %w", err)
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return nil, fmt.Errorf("run: open ledger %s: %w", path, err)
	}
	return &Ledger{file: file, writer: bufio.NewWriter(file)}, nil
}

// Append writes one delivery row as a JSON line. A row that cannot be encoded
// or buffered is dropped from the file (it stays in the in-memory ledger).
func (l *Ledger) Append(rec model.LedgerRecord) {
	if raw, err := json.Marshal(rec); err == nil {
		_, _ = l.writer.Write(raw)
		_ = l.writer.WriteByte('\n')
	}
}

// Flush pushes buffered rows to the file descriptor without syncing.
func (l *Ledger) Flush() error {
	if err := l.writer.Flush(); err != nil {
		return fmt.Errorf("run: flush ledger: %w", err)
	}
	return nil
}

// Finish flushes, syncs and closes the file at the end of the run.
func (l *Ledger) Finish() error {
	if err := l.writer.Flush(); err != nil {
		_ = l.file.Close()
		return fmt.Errorf("End: flush ledger: %w", err)
	}
	if err := l.file.Sync(); err != nil {
		_ = l.file.Close()
		return fmt.Errorf("End: sync ledger: %w", err)
	}
	if err := l.file.Close(); err != nil {
		return fmt.Errorf("End: close ledger: %w", err)
	}
	return nil
}
