package app

import (
	"sort"

	rules "github.com/ghassan-ai-projects/streams-simulator/internal/director/internal/domain"
)

// CatalogList lists the installed domains, optionally one group
// (sim.catalog.list).
func (d *Director) CatalogList(group string) map[string]any {
	return map[string]any{"domains": d.Catalog.List(group)}
}

// CatalogDescribe describes one domain: its spec and digest.
func (d *Director) CatalogDescribe(domainID string) (map[string]any, error) {
	c, err := d.Catalog.Describe(domainID)
	if err != nil {
		return nil, rules.ToolErrorf(rules.CodeDomainInvalid, "%v", err)
	}
	return map[string]any{"spec": c.Spec, "digest": c.Digest}, nil
}

// CatalogCoverage reports the axis-coverage matrix.
func (d *Director) CatalogCoverage() any {
	return d.Catalog.Coverage()
}

// DomainSpec returns the declared spec of one domain.
func (d *Director) DomainSpec(domainID string) (any, error) {
	compiled, err := d.Catalog.Describe(domainID)
	if err != nil {
		return nil, rules.ToolErrorf(rules.CodeDomainInvalid, "%v", err)
	}
	return compiled.Spec, nil
}

// ListAdapters lists the installed output adapters in id order.
func (d *Director) ListAdapters() map[string]any {
	ids := make([]string, 0, len(d.Adapters))
	// determinism-safe: collected here, sorted below before output.
	for id := range d.Adapters {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	var out []map[string]any
	for _, id := range ids {
		a := d.Adapters[id]
		out = append(out, map[string]any{"id": id, "version": a.Version, "encoding": a.Encoding, "title": a.Title})
	}
	return map[string]any{"adapters": out}
}
