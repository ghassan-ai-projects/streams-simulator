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
		out = append(out, Entry{
			ID:       d.Spec.ID,
			Title:    d.Spec.Title,
			Version:  d.Spec.Version,
			Axes:     axisVector(d.Spec),
			Stresses: d.Spec.Stresses,
		})
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
	axes := map[string][]string{
		"rate": nil, "cardinality": nil, "value_shape": nil,
		"cadence": nil, "lateness": nil, "absence": nil,
		"time_reference": nil, "correlation": nil, "seasonality": nil,
		"actuation": nil, "consequence": nil, "fidelity": nil,
	}
	by := map[string]map[string]int{}
	for a := range axes {
		by[a] = map[string]int{}
	}
	for _, d := range c.domains {
		add := func(axis string, values []string) {
			for _, v := range values {
				if v != "" {
					by[axis][v]++
				}
			}
		}
		s := d.Spec
		add("rate", []string{s.Axes.Rate})
		add("cardinality", []string{s.Axes.Cardinality})
		add("value_shape", s.Axes.ValueShape)
		add("cadence", s.Axes.Cadence)
		add("lateness", []string{s.Axes.Lateness})
		add("absence", []string{s.Axes.Absence})
		add("time_reference", []string{s.Axes.TimeRef})
		add("correlation", s.Axes.Correlation)
		add("seasonality", s.Axes.Seasonality)
		add("actuation", []string{s.Axes.Actuation})
		add("consequence", s.Axes.Consequence)
		add("fidelity", s.Axes.Fidelity)
	}
	var thin []string
	for axis, counts := range by {
		for value, n := range counts {
			if n < 2 {
				thin = append(thin, fmt.Sprintf("%s=%s(%d)", axis, value, n))
			}
		}
	}
	sort.Strings(thin)
	return CoverageReport{ByAxis: by, Thin: thin}
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
