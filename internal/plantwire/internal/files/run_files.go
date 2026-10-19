package files

import (
	"fmt"
	"path/filepath"

	"github.com/ghassan-ai-projects/streams-simulator/internal/kernel"
)

func ReadRunFile(dir, name string) (raw []byte, present bool, err error) {
	raw, present, err = kernel.ReadIfPresent(filepath.Join(dir, name))
	if err != nil {
		return nil, false, fmt.Errorf("plantwire: read %s: %w", name, err)
	}
	return raw, present, nil
}
