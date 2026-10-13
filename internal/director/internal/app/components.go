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
	return d.Catalog
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
		return nil, rules.ToolErrorf(rules.CodeComponentInvalid, "%v", err)
	}
	return detail, nil
}

func (d *Director) ComposeDomain(composition map[string]any, install bool) (map[string]any, error) {
	compiled, err := d.composed(composition)
	if err != nil {
		return nil, err
	}
	if install {
		if err := d.installDomain(compiled); err != nil {
			return nil, err
		}
	}
	return map[string]any{"domain": compiled.Spec, "digest": compiled.Digest, "installed": install}, nil
}

func (d *Director) composed(composition map[string]any) (*domain.Compiled, error) {
	out, err := d.composedBytes(composition)
	if err != nil {
		return nil, err
	}
	compiled, err := domain.Parse(out, "composition")
	if err != nil {
		return nil, rules.ToolErrorf(rules.CodeDomainInvalid, "%v", err)
	}
	return compiled, nil
}

func (d *Director) composedBytes(composition map[string]any) ([]byte, error) {
	library, err := d.componentLibrary()
	if err != nil {
		return nil, err
	}
	raw, err := json.Marshal(composition)
	if err != nil {
		return nil, rules.ToolErrorf(rules.CodeInvalidArgs, "composition: %v", err)
	}
	out, err := component.Compose(library, raw, "composition")
	if err != nil {
		return nil, rules.ToolErrorf(rules.CodeComponentInvalid, "%v", err)
	}
	return out, nil
}
