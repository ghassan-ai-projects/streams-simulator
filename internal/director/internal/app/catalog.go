package app

import (
	"sort"

	rules "github.com/ghassan-ai-projects/streams-simulator/internal/director/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

func (d *Director) CatalogList(group string) map[string]any {
	return map[string]any{"domains": d.catalog().List(group)}
}

func (d *Director) CatalogDescribe(domainID string) (map[string]any, error) {
	c, err := d.catalog().Describe(domainID)
	if err != nil {
		return nil, rules.ToolErrorf(rules.CodeDomainInvalid, "%v", err)
	}
	return map[string]any{"spec": c.Spec, "digest": c.Digest, "fault_categories": rules.FaultCategories(c.Spec)}, nil
}

func (d *Director) PerturbationCatalog() map[string]any {
	return map[string]any{"perturbations": run.Perturbations()}
}

func (d *Director) CatalogCoverage() any {
	return d.catalog().Coverage()
}

func (d *Director) DomainSpec(domainID string) (any, error) {
	compiled, err := d.catalog().Describe(domainID)
	if err != nil {
		return nil, rules.ToolErrorf(rules.CodeDomainInvalid, "%v", err)
	}
	return compiled.Spec, nil
}

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
