// Package files is the file-system edge of the domain module: it reads domain
// spec files and directories and hands the bytes to the pure layer, which
// parses, validates and compiles them.
package files

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain/internal/domain"
)

// Load reads, schema-validates and structurally validates a domain spec from
// a path. The returned Compiled carries the spec, its canonical digest, and
// resolved defaults.
func Load(path string) (*domain.Compiled, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("domain: read %s: %w", path, err)
	}
	return domain.Parse(raw, path)
}

// LoadAll loads every domain spec in a directory (non-recursive), sorted by
// path. A single unloadable domain fails the whole load: the catalog is a
// fixed, reviewed set, and a silent skip would make coverage reports lie.
func LoadAll(dir string) ([]*domain.Compiled, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("domain: list %s: %w", dir, err)
	}
	return loadPaths(domainPaths(dir, entries))
}

func domainPaths(dir string, entries []os.DirEntry) []string {
	var paths []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
			paths = append(paths, filepath.Join(dir, e.Name()))
		}
	}
	sort.Strings(paths)
	return paths
}

func loadPaths(paths []string) ([]*domain.Compiled, error) {
	var out []*domain.Compiled
	for _, p := range paths {
		c, err := Load(p)
		if err != nil {
			return nil, fmt.Errorf("streamsim: %w", err)
		}
		out = append(out, c)
	}
	return out, nil
}
