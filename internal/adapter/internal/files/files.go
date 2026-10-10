// Package files is the file-system edge of the adapter module: it reads an
// adapter file and hands the bytes to the pure layer, which validates and
// decodes them.
package files

import (
	"fmt"
	"os"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// Load reads and validates an adapter file. The validation covers the
// output-adapter schema plus adapter-specific cross-checks (transform
// arity, source names, identity preservation).
func Load(path string) (*model.Adapter, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("adapter: read %s: %w", path, err)
	}
	return domain.DecodeFile(raw, path)
}
