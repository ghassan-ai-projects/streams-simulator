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
func (d directorTools) handleCatalogPerturbations(_ context.Context, _ map[string]any) (any, error) {
	return d.PerturbationCatalog(), nil
}
func (d directorTools) handleComponentList(_ context.Context, _ map[string]any) (any, error) {
	return d.ComponentList()
}

func (d directorTools) handleComponentDescribe(_ context.Context, args map[string]any) (any, error) {
	return d.ComponentDescribe(stringArg(args, "ref"))
}

func (d directorTools) handleDomainCompose(_ context.Context, args map[string]any) (any, error) {
	composition, _ := args["composition"].(map[string]any)
	return d.ComposeDomain(composition, boolArg(args, "install"))
}

func (d directorTools) handleAdapterList(_ context.Context, args map[string]any) (any, error) {
	return d.ListAdapters(), nil
}
