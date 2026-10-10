package protocol

import (
	"context"
	"sort"

	"github.com/ghassan-ai-projects/streams-simulator/internal/mcp/internal/app"
)

func (d directorTools) handleCatalogList(_ context.Context, args map[string]any) (any, error) {
	return map[string]any{"domains": d.Catalog.List(app.Str(args, "group"))}, nil
}
func (d directorTools) handleCatalogDescribe(_ context.Context, args map[string]any) (any, error) {
	c, err := d.Catalog.Describe(app.Str(args, "domain"))
	if err != nil {
		return nil, app.ToolErrorf(app.CodeDomainInvalid, "%v", err)
	}
	return map[string]any{"spec": c.Spec, "digest": c.Digest}, nil
}
func (d directorTools) handleCatalogCoverage(_ context.Context, args map[string]any) (any, error) {
	return d.Catalog.Coverage(), nil
}
func (d directorTools) handleAdapterList(_ context.Context, args map[string]any) (any, error) {
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
