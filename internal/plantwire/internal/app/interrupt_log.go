package app

import (
	"fmt"
	"log/slog"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

type InterruptLog interface {
	Add(record map[string]any) error
	Flush() error
}

type interruptFile struct {
	file   *run.InterruptLog
	failed bool
}

func openInterruptFile(dir string) (*interruptFile, error) {
	file, err := run.NewInterruptLog(dir)
	if err != nil {
		return nil, fmt.Errorf("plantwire: %w", err)
	}
	return &interruptFile{file: file}, nil
}

func (f *interruptFile) Add(record map[string]any) error {
	if f == nil || f.failed {
		return nil
	}
	if err := f.file.Add(record); err != nil {
		slog.Warn("plantwire: the interrupt log could not take a record, it is left out of the run", "error", err)
		f.failed = true
	}
	return nil
}

func (f *interruptFile) Flush() error {
	if f == nil || f.failed {
		return nil
	}
	if err := f.file.Flush(); err != nil {
		return fmt.Errorf("plantwire: flush the interrupt log: %w", err)
	}
	return nil
}

func (f *interruptFile) finish(dir string) error {
	if f == nil {
		return nil
	}
	if err := f.file.Close(); err != nil {
		f.failed = true
		return fmt.Errorf("plantwire: close the interrupt log: %w", err)
	}
	f.check(dir)
	return nil
}

func (f *interruptFile) check(dir string) {
	if _, err := VerifyInterruptLog(dir); err != nil && !f.failed {
		slog.Warn("plantwire: the interrupt log does not replay, it is left out of the run", "error", err)
		f.failed = true
	}
}

func (f *interruptFile) digest() string {
	if f == nil || f.failed {
		return ""
	}
	return f.file.Digest()
}

func recordInterrupts(entries []domain.InterruptEntry, log InterruptLog, journal func(map[string]any)) {
	for _, entry := range entries {
		if log != nil {
			_ = log.Add(entry)
		}
		if journal != nil {
			journal(entry)
		}
	}
	if log != nil && len(entries) > 0 {
		_ = log.Flush()
	}
}

type InterruptReplay struct {
	Records int    `json:"records"`
	Digest  string `json:"digest"`
}

func VerifyInterruptLog(dir string) (InterruptReplay, error) {
	records, digest, err := run.ReadInterruptLog(dir)
	if err == nil {
		err = checkedAndReplayed(records)
	}
	if err != nil {
		return InterruptReplay{}, fmt.Errorf("plantwire interrupts: %w", err)
	}
	return InterruptReplay{Records: len(records), Digest: digest}, nil
}

func checkedAndReplayed(records []map[string]any) error {
	if err := checkInterruptRecords(records); err != nil {
		return err
	}
	return domain.ReplayInterrupts(records)
}

func checkInterruptRecords(records []map[string]any) error {
	for i, record := range records {
		raw, err := canonical.Marshal(record)
		if err == nil {
			err = model.ValidateInterruptRecord(raw)
		}
		if err != nil {
			return fmt.Errorf("record %d: %w", i, err)
		}
	}
	return nil
}
