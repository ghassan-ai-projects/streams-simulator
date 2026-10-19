package durable

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/ghassan-ai-projects/streams-simulator/internal/kernel"
)

const (
	NounFieldTrace   = "field trace"
	NounInterruptLog = "interrupt log"
)

type FieldTraceFile struct {
	log  *kernel.Log
	noun string
}

func OpenEvidenceLog(dir, file, noun string) (*FieldTraceFile, error) {
	log, err := kernel.NewLog(filepath.Join(dir, file), false)
	if err != nil {
		return nil, fmt.Errorf("run: open %s in %s: %w", noun, dir, err)
	}
	return &FieldTraceFile{log: log, noun: noun}, nil
}

func (f *FieldTraceFile) Append(line []byte) error {
	if err := f.log.Write(line); err != nil {
		return fmt.Errorf("run: write %s: %w", f.noun, err)
	}
	return nil
}

func (f *FieldTraceFile) Flush() error {
	if err := f.log.Flush(); err != nil {
		return fmt.Errorf("run: flush %s: %w", f.noun, err)
	}
	return nil
}

func (f *FieldTraceFile) Close() error {
	if err := f.log.Close(); err != nil {
		return fmt.Errorf("run: close %s: %w", f.noun, err)
	}
	return nil
}

func EvidenceLogExists(dir, file string) (bool, error) {
	reader, present, err := kernel.OpenIfPresent(filepath.Join(dir, file))
	if err != nil {
		return false, fmt.Errorf("run: look for %s: %w", file, err)
	}
	if present {
		_ = reader.Close()
	}
	return present, nil
}

func RenameEvidenceLog(dir, from, to string) error {
	if err := kernel.Rename(filepath.Join(dir, from), filepath.Join(dir, to)); err != nil {
		return fmt.Errorf("run: set aside %s: %w", from, err)
	}
	return nil
}

func ReadEvidenceLog(dir, file, noun string, limit int64) (raw []byte, present bool, err error) {
	raw, err = kernel.ReadRegular(filepath.Join(dir, file), limit)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("run: read %s: %w", noun, err)
	}
	return raw, true, nil
}
