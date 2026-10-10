package domain

import (
	layer "github.com/ghassan-ai-projects/streams-simulator/internal/domain/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain/internal/files"
)

// Load reads, schema-validates and structurally validates a domain spec from
// a path. The returned Compiled carries the spec, its canonical digest, and
// resolved defaults.
func Load(path string) (*Compiled, error) {
	return files.Load(path)
}

// LoadAll loads every domain spec in a directory (non-recursive), sorted by
// path. A single unloadable domain fails the whole load: the catalog is a
// fixed, reviewed set, and a silent skip would make coverage reports lie.
func LoadAll(dir string) ([]*Compiled, error) {
	return files.LoadAll(dir)
}

// Parse reads, schema-validates and structurally validates a domain spec from
// an in-memory document. src names the source for error messages.
func Parse(raw []byte, src string) (*Compiled, error) {
	return layer.Parse(raw, src)
}

// List returns the catalog rows whose id starts with group ("" lists all).
func (c *Catalog) List(group string) []Entry {
	if c == nil || c.catalog == nil {
		return nil
	}
	return c.catalog.List(group)
}

// Describe returns the compiled domain with the given id.
func (c *Catalog) Describe(id string) (*Compiled, error) {
	if c == nil || c.catalog == nil {
		return nil, ErrNoCatalog
	}
	return c.catalog.Describe(id)
}

// Coverage reports which axis values the installed domains cover.
func (c *Catalog) Coverage() CoverageReport {
	if c == nil || c.catalog == nil {
		return CoverageReport{}
	}
	return c.catalog.Coverage()
}
