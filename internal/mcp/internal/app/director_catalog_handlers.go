package app

import (
	"context"
	"sort"
)

func (d *Director) handleCatalogList(_ context.Context, args map[string]any) (any, error) {
	return map[string]any{"domains": d.Catalog.List(str(args, "group"))}, nil
}
func (d *Director) handleCatalogDescribe(_ context.Context, args map[string]any) (any, error) {
	c, err := d.Catalog.Describe(str(args, "domain"))
	if err != nil {
		return nil, errTool(CodeDomainInvalid, "%v", err)
	}
	return map[string]any{"spec": c.Spec, "digest": c.Digest}, nil
}
func (d *Director) handleCatalogCoverage(_ context.Context, args map[string]any) (any, error) {
	return d.Catalog.Coverage(), nil
}
func (d *Director) handleAdapterList(_ context.Context, args map[string]any) (any, error) {
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
	return map[string]any{"adapters": out}, nil
}
