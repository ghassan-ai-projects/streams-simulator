package domain

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// This file is the package's only file-system edge: everything else parses,
// validates and compiles in-memory documents.

// Load reads, schema-validates and structurally validates a domain spec from
// a path. The returned Compiled carries the spec, its canonical digest, and
// resolved defaults.
func Load(path string) (*Compiled, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("domain: read %s: %w", path, err)
	}
	return Parse(raw, path)
}

// LoadAll loads every domain spec in a directory (non-recursive), sorted by
// id. A single unloadable domain fails the whole load: the catalog is a
// fixed, reviewed set, and a silent skip would make coverage reports lie.
func LoadAll(dir string) ([]*Compiled, error) {
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

func loadPaths(paths []string) ([]*Compiled, error) {
	var out []*Compiled
	for _, p := range paths {
		c, err := Load(p)
		if err != nil {
			return nil, fmt.Errorf("streamsim: %w", err)
		}
		out = append(out, c)
	}
	return out, nil
}
