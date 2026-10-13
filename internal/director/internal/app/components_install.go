package app

import (
	rules "github.com/ghassan-ai-projects/streams-simulator/internal/director/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
)

func (d *Director) installDomain(compiled *domain.Compiled) error {
	d.catalogMu.Lock()
	defer d.catalogMu.Unlock()
	if existing, err := d.Catalog.Describe(compiled.Spec.ID); err == nil {
		if existing.Digest == compiled.Digest {
			return nil
		}
		return rules.ToolErrorf(rules.CodeDomainInvalid, "domain %q is already installed with another digest", compiled.Spec.ID)
	}
	d.Catalog = domain.NewCatalog(append(installedDomains(d.Catalog), compiled))
	return nil
}

func installedDomains(catalog *domain.Catalog) []*domain.Compiled {
	var out []*domain.Compiled
	for _, entry := range catalog.List("") {
		if compiled, err := catalog.Describe(entry.ID); err == nil {
			out = append(out, compiled)
		}
	}
	return out
}
