package domain

import (
	"errors"
	"fmt"
	"slices"
	"sort"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

var (
	ErrUnknownDomain = errors.New("unknown domain")
	ErrDomainIDTaken = errors.New("domain id is already installed with another digest")
)

type Catalog struct {
	domains []*Compiled
}

func NewCatalog(domains []*Compiled) *Catalog {
	sorted := make([]*Compiled, len(domains))
	copy(sorted, domains)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Spec.ID < sorted[j].Spec.ID })
	return &Catalog{domains: sorted}
}

func CatalogWith(base *Catalog, compiled *Compiled) (*Catalog, error) {
	existing, err := base.Describe(compiled.Spec.ID)
	switch {
	case err != nil:
		return NewCatalog(append(slices.Clone(base.domains), compiled)), nil
	case existing.Digest == compiled.Digest:
		return base, nil
	}
	return nil, fmt.Errorf("catalog: %w: %q", ErrDomainIDTaken, compiled.Spec.ID)
}

type Entry struct {
	ID       string   `json:"id"`
	Title    string   `json:"title"`
	Version  string   `json:"version"`
	Axes     []string `json:"axes"`
	Stresses string   `json:"stresses"`
}

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

func (c *Catalog) Describe(id string) (*Compiled, error) {
	for _, d := range c.domains {
		if d.Spec.ID == id {
			return d, nil
		}
	}
	return nil, fmt.Errorf("catalog: %w %q", ErrUnknownDomain, id)
}

type CoverageReport struct {
	ByAxis map[string]map[string]int `json:"by_axis"`

	Thin []string `json:"thin"`
}

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
