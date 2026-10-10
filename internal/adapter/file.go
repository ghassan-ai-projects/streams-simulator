package adapter

import (
	"fmt"
	"os"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// This file is the package's only file-system edge: everything else decodes
// and renders in-memory documents.

// Load reads and validates an adapter file. The validation covers the
// output-adapter schema plus adapter-specific cross-checks (transform
// arity, source names, identity preservation).
func Load(path string) (*model.Adapter, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("adapter: read %s: %w", path, err)
	}
	return decodeAdapter(raw, path, "\n  ")
}
