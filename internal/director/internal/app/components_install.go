package app

import (
	"errors"
	"fmt"

	rules "github.com/ghassan-ai-projects/streams-simulator/internal/director/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
)

func (d *Director) installDomain(compiled *domain.Compiled) error {
	d.catalogMu.Lock()
	defer d.catalogMu.Unlock()
	grown, err := domain.NewCatalogWith(d.domains, compiled)
	switch {
	case errors.Is(err, domain.ErrDomainIDTaken):
		return rules.ToolErrorf(rules.CodeDomainInvalid, "%v", err)
	case err != nil:
		return fmt.Errorf("director: install domain: %w", err)
	}
	d.domains = grown
	return nil
}
