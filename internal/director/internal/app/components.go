package app

import (
	"encoding/json"

	"github.com/ghassan-ai-projects/streams-simulator/internal/component"
	rules "github.com/ghassan-ai-projects/streams-simulator/internal/director/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
)

func (d *Director) catalog() *domain.Catalog {
	d.catalogMu.RLock()
	defer d.catalogMu.RUnlock()
	return d.domains
}

func (d *Director) componentLibrary() (*component.Library, error) {
	if d.library == nil {
		return nil, rules.ToolErrorf(rules.CodeNoLibrary, "this director has no component library")
	}
	return d.library, nil
}

func (d *Director) ComponentList() (map[string]any, error) {
	library, err := d.componentLibrary()
	if err != nil {
		return nil, err
	}
	return map[string]any{"components": library.List()}, nil
}

func (d *Director) ComponentDescribe(ref string) (any, error) {
	library, err := d.componentLibrary()
	if err != nil {
		return nil, err
	}
	detail, err := library.Describe(ref)
	if err != nil {
		return nil, rules.ValidationToolError(err, rules.CodeComponentInvalid)
	}
	return detail, nil
}

type composedDomain struct {
	domain.Summary
	Domain    any  `json:"domain"`
	Installed bool `json:"installed"`
}

func (d *Director) ComposeDomain(composition map[string]any, install bool) (any, error) {
	composed, err := d.composed(composition)
	if err != nil {
		return nil, err
	}
	if install {
		if err := d.installDomain(composed.Domain); err != nil {
			return nil, err
		}
	}
	return composedDomain{Summary: composed.Summary, Domain: composed.Domain.Spec, Installed: install}, nil
}

func (d *Director) composed(composition map[string]any) (*component.Composed, error) {
	library, err := d.componentLibrary()
	if err != nil {
		return nil, err
	}
	raw, err := json.Marshal(composition)
	if err != nil {
		return nil, rules.ToolErrorf(rules.CodeInvalidArgs, "composition: %v", err)
	}
	composed, err := component.Compose(library, raw, "composition")
	if err != nil {
		return nil, rules.ValidationToolError(err, rules.CodeComponentInvalid)
	}
	return composed, nil
}
