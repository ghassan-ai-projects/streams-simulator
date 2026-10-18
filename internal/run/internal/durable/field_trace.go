package durable

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/ghassan-ai-projects/streams-simulator/internal/kernel"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

type FieldTraceFile struct {
	log *kernel.Log
}

func OpenFieldTrace(dir string) (*FieldTraceFile, error) {
	log, err := kernel.NewLog(filepath.Join(dir, model.FileFieldTrace), false)
	if err != nil {
		return nil, fmt.Errorf("run: open field trace in %s: %w", dir, err)
	}
	return &FieldTraceFile{log: log}, nil
}

func (f *FieldTraceFile) Append(line []byte) error {
	if err := f.log.Write(line); err != nil {
		return fmt.Errorf("run: write field trace: %w", err)
	}
	return nil
}

func (f *FieldTraceFile) Flush() error {
	if err := f.log.Flush(); err != nil {
		return fmt.Errorf("run: flush field trace: %w", err)
	}
	return nil
}

func (f *FieldTraceFile) Close() error {
	if err := f.log.Close(); err != nil {
		return fmt.Errorf("run: close field trace: %w", err)
	}
	return nil
}

func ReadFieldTrace(dir string, limit int64) (raw []byte, present bool, err error) {
	raw, err = kernel.ReadRegular(filepath.Join(dir, model.FileFieldTrace), limit)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("run: read field trace: %w", err)
	}
	return raw, true, nil
}
