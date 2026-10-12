// Package files is the file-system edge of the adapter module: it reads an
// adapter file and hands the bytes to the pure layer, which validates and
// decodes them.
package files

import (
	"cmp"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

// LoadDir loads every adapter in a directory, keyed by adapter id; an empty
// dir selects "adapters", the layout of the repository. A directory that holds
// no adapter is refused.
func LoadDir(dir string) (map[string]*model.Adapter, error) {
	dir = cmp.Or(dir, "adapters")
	names, err := fileNames(dir)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return decodeDirectory(dir, names)
}

// fileNames lists the names of the non-directory entries of dir, sorted by
// name.
func fileNames(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	return names, nil
}

func decodeDirectory(dir string, names []string) (map[string]*model.Adapter, error) {
	out, err := loadEntries(dir, names)
	if err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no adapters found in %s", dir)
	}
	return out, nil
}

func loadEntries(dir string, names []string) (map[string]*model.Adapter, error) {
	out := map[string]*model.Adapter{}
	for _, name := range names {
		if !strings.HasSuffix(name, ".json") {
			continue
		}
		a, err := Load(filepath.Join(dir, name))
		if err != nil {
			return nil, fmt.Errorf("streamsim: %w", err)
		}
		out[a.ID] = a
	}
	return out, nil
}
