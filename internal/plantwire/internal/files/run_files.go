package files

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/ghassan-ai-projects/streams-simulator/internal/kernel"
)

const MaxRunFileBytes = 64 << 20

func ReadRunFile(dir, name string) (raw []byte, present bool, err error) {
	return ReadRunFileUpTo(dir, name, MaxRunFileBytes)
}

func RunFilePresent(dir, name string) (bool, error) {
	reader, present, err := kernel.OpenIfPresent(filepath.Join(dir, name))
	if err != nil {
		return false, fmt.Errorf("plantwire: look for %s: %w", name, err)
	}
	if present {
		_ = reader.Close()
	}
	return present, nil
}

func ReadRunFileUpTo(dir, name string, limit int64) (raw []byte, present bool, err error) {
	raw, err = kernel.ReadRegular(filepath.Join(dir, name), limit)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("plantwire: read %s: %w", name, err)
	}
	return raw, true, nil
}

func WriteRunFile(dir, name string, data []byte) error {
	if err := kernel.WriteAtomic(filepath.Join(dir, name), data); err != nil {
		return fmt.Errorf("plantwire: write %s: %w", name, err)
	}
	return nil
}
