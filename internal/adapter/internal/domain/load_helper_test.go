package domain

import (
	"fmt"
	"os"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// Load reads and validates an adapter file for the layer's tests;
// production loading is the files edge's job.
func Load(path string) (*model.Adapter, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	return DecodeFile(raw, path)
}
