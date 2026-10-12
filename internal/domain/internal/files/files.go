// Package files is the file-system edge of the domain module: it reads domain
// spec files and directories and hands the bytes to the pure layer, which
// parses, validates and compiles them.
package files

import (
	"bytes"
	"cmp"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/kernel"
)

// Load reads, schema-validates and structurally validates a domain spec from
// a path. The returned Compiled carries the spec, its canonical digest, and
// resolved defaults.
func Load(path string) (*domain.Compiled, error) {
	raw, err := kernel.Read(path)
	if err != nil {
		return nil, fmt.Errorf("domain: read %s: %w", path, err)
	}
	return domain.Parse(raw, path)
}

// LoadAll loads every domain spec in a directory (non-recursive), sorted by
// path. A single unloadable domain fails the whole load: the catalog is a
// fixed, reviewed set, and a silent skip would make coverage reports lie.
func LoadAll(dir string) ([]*domain.Compiled, error) {
	names, err := kernel.Names(dir)
	if err != nil {
		return nil, fmt.Errorf("domain: list %s: %w", dir, err)
	}
	return loadPaths(domainPaths(dir, names))
}

func domainPaths(dir string, names []string) []string {
	var paths []string
	for _, name := range names {
		if strings.HasSuffix(name, ".json") {
			paths = append(paths, filepath.Join(dir, name))
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

// LoadCatalog loads every domain in a directory into a catalog; an empty dir
// selects "domains", the layout of the repository. A directory that holds no
// domain is refused, because an empty catalog would report coverage of nothing.
func LoadCatalog(dir string) (*domain.Catalog, error) {
	dir = cmp.Or(dir, "domains")
	specs, err := LoadAll(dir)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	if len(specs) == 0 {
		return nil, fmt.Errorf("no domains found in %s", dir)
	}
	return domain.NewCatalog(specs), nil
}

type FormatResult struct {
	Path    string `json:"formatted"`
	Changed bool   `json:"changed"`
	Digest  string `json:"digest"`
}

func Format(path string) (*FormatResult, error) {
	compiled, err := Load(path)
	if err != nil {
		return nil, err
	}
	formatted, err := domain.Format(compiled)
	if err != nil {
		return nil, fmt.Errorf("domain: %w", err)
	}
	changed := !bytes.Equal(compiled.Raw, formatted)
	if changed {
		err = writeFormatted(path, formatted)
	}
	return &FormatResult{Path: path, Changed: changed, Digest: compiled.Digest}, err
}

func writeFormatted(path string, content []byte) error {
	if err := kernel.Replace(path, content); err != nil {
		return fmt.Errorf("domain: %w", err)
	}
	return nil
}
