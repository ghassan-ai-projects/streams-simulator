package domain

import (
	"fmt"
	"os"
)

// Load reads and parses a domain file for the layer's tests; production
// loading is the files edge's job.
func Load(path string) (*Compiled, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	return Parse(raw, path)
}
