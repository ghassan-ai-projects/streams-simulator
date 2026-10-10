package domain

import (
	"errors"

	layer "github.com/ghassan-ai-projects/streams-simulator/internal/domain/internal/domain"
)

// Compiled is the typed, validated, digest-carrying form of a domain spec:
// the spec, its canonical digest, the validated source document, and
// read-only lookups over its declared names.
type Compiled = layer.Compiled

// Entry is one catalog row.
type Entry = layer.Entry

// CoverageReport is the catalog's axis coverage over the installed domains.
type CoverageReport = layer.CoverageReport

// ErrNoCatalog is returned by Describe on a nil or zero Catalog.
var ErrNoCatalog = errors.New("domain: a catalog built by NewCatalog is required")
