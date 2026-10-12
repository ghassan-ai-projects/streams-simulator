package protocol

import (
	"context"
)

func (d directorTools) handleCatalogList(_ context.Context, args map[string]any) (any, error) {
	return d.CatalogList(stringArg(args, "group")), nil
}
func (d directorTools) handleCatalogDescribe(_ context.Context, args map[string]any) (any, error) {
	return d.CatalogDescribe(stringArg(args, "domain"))
}
func (d directorTools) handleCatalogCoverage(_ context.Context, args map[string]any) (any, error) {
	return d.CatalogCoverage(), nil
}
func (d directorTools) handleAdapterList(_ context.Context, args map[string]any) (any, error) {
	return d.ListAdapters(), nil
}
