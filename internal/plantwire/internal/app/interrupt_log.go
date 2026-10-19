package app

import (
	"errors"
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

var ErrInterruptLogRejected = errors.New("plantwire: the interrupt log is not evidence and was set aside")

type interruptFile struct {
	file   *run.InterruptLog
	failed bool
	cause  error
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
		f.failed, f.cause = true, err
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
		f.failed, f.cause = true, err
	} else if !f.failed {
		f.check(dir)
	}
	if !f.failed {
		return nil
	}
	slog.Warn("plantwire: the interrupt log is set aside and left out of the run", "error", f.cause)
	return errors.Join(fmt.Errorf("%w: %w", ErrInterruptLogRejected, f.cause), run.RejectInterruptLog(dir))
}

func (f *interruptFile) check(dir string) {
	if _, err := VerifyInterruptLog(dir); err != nil {
		f.failed, f.cause = true, err
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
	records, digest, err := verifiedInterruptLog(dir)
	if err != nil {
		return InterruptReplay{}, err
	}
	return InterruptReplay{Records: len(records), Digest: digest}, nil
}

func verifiedInterruptLog(dir string) ([]map[string]any, string, error) {
	records, digest, err := run.ReadInterruptLog(dir)
	if err == nil {
		err = checkedAndReplayed(records)
	}
	if err != nil {
		return nil, "", fmt.Errorf("plantwire interrupts: %w", err)
	}
	return records, digest, nil
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
