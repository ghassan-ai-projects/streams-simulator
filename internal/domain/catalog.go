package domain

import (
	"fmt"
	"sort"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// Catalog is the set of installed domains, with list/describe/coverage
// operations. It is the director role's catalog surface.
type Catalog struct {
	domains []*Compiled
}

// NewCatalog builds a catalog from loaded domain specs, sorted by id.
func NewCatalog(domains []*Compiled) *Catalog {
	sorted := make([]*Compiled, len(domains))
	copy(sorted, domains)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Spec.ID < sorted[j].Spec.ID })
	return &Catalog{domains: sorted}
}

// Entry is one row of the catalog list.
type Entry struct {
	ID       string   `json:"id"`
	Title    string   `json:"title"`
	Version  string   `json:"version"`
	Axes     []string `json:"axes"`
	Stresses string   `json:"stresses"`
}

// List returns catalog entries, optionally filtered by group. The group
// concept is an organizational label in the docs; the binary groups by
// nothing but id, so the filter matches the domain id prefix.
func (c *Catalog) List(group string) []Entry {
	var out []Entry
	for _, d := range c.domains {
		if group != "" && !hasPrefix(d.Spec.ID, group) {
			continue
		}
		out = append(out, catalogEntry(d.Spec))
	}
	return out
}

// Describe returns the full spec for a domain id.
func (c *Catalog) Describe(id string) (*Compiled, error) {
	for _, d := range c.domains {
		if d.Spec.ID == id {
			return d, nil
		}
	}
	return nil, fmt.Errorf("catalog: unknown domain %q", id)
}

// CoverageReport is the axis-coverage matrix: which axis values are
// exercised, and which are thin.
type CoverageReport struct {
	// ByAxis maps axis name -> value -> number of domains carrying it.
	ByAxis map[string]map[string]int `json:"by_axis"`
	// Thin lists axis values covered by fewer than two domains.
	Thin []string `json:"thin"`
}

// Coverage computes the axis-coverage matrix across the catalog.
func (c *Catalog) Coverage() CoverageReport {
	by := coverageCounts()
	for _, d := range c.domains {
		addDomainCoverage(by, d.Spec)
	}
	return CoverageReport{ByAxis: by, Thin: thinCoverage(by)}
}

func axisVector(s *model.DomainSpec) []string {
	return []string{
		"rate=" + s.Axes.Rate,
		"cardinality=" + s.Axes.Cardinality,
		"lateness=" + s.Axes.Lateness,
		"absence=" + s.Axes.Absence,
		"time_reference=" + s.Axes.TimeRef,
		"actuation=" + s.Axes.Actuation,
	}
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
