package domain

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
	specs "github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

type composer struct {
	lib         *Library
	source      string
	composition *Composition
	areas       map[string]bool
	kinds       map[string]bool
}

func Compose(lib *Library, raw []byte, source string) ([]byte, error) {
	composition, err := parseComposition(lib.contracts.composition, raw, source)
	if err != nil {
		return nil, err
	}
	c := &composer{lib: lib, source: source, composition: composition, areas: map[string]bool{}, kinds: map[string]bool{}}
	document, err := c.document()
	if err != nil {
		return nil, err
	}
	return render(document, source)
}

func (c *composer) refuse(at, format string, args ...any) error {
	return refuse(model.ContractComposition, c.source, at, format, args...)
}

func (c *composer) document() (map[string]any, error) {
	if err := c.declareAreas(); err != nil {
		return nil, err
	}
	entries, err := c.buildAll()
	if err != nil {
		return nil, err
	}
	links, err := c.connect(entries)
	if err != nil {
		return nil, err
	}
	return c.assemble(entries, links), nil
}

func (c *composer) buildAll() ([]*built, error) {
	entries := make([]*built, 0, len(c.composition.Equipment))
	for index, entry := range c.composition.Equipment {
		next, err := c.buildEntry(index, entry)
		if err != nil {
			return nil, err
		}
		entries = append(entries, next)
	}
	return entries, nil
}

func render(document map[string]any, source string) ([]byte, error) {
	raw, err := canonical.Marshal(document)
	if err != nil {
		return nil, fmt.Errorf("component: %s: %w", source, err)
	}
	compiled, err := specs.Parse(raw, source)
	if err != nil {
		return nil, fmt.Errorf("component: the domain composed from %s is invalid: %w", source, err)
	}
	formatted, err := specs.Render(compiled)
	if err != nil {
		return nil, fmt.Errorf("component: format the domain composed from %s: %w", source, err)
	}
	return formatted, nil
}

func (c *composer) declareAreas() error {
	if len(c.composition.Areas) > 0 && c.composition.Site == nil {
		return c.refuse("/areas", "areas need a site: declare site.name")
	}
	for index, area := range c.composition.Areas {
		if err := c.declareArea(pointer("areas", indexPart(index)), area); err != nil {
			return err
		}
	}
	return nil
}

func (c *composer) declareArea(at string, area model.Area) error {
	if c.areas[area.Name] {
		return c.refuse(at+"/name", "area %q is declared twice", area.Name)
	}
	if area.Parent != "" && !c.areas[area.Parent] {
		return c.refuse(at+"/parent", "parent %q must be declared before area %q", area.Parent, area.Name)
	}
	c.areas[area.Name] = true
	return nil
}
