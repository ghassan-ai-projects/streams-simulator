package app

import (
	"context"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/component"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func LoadDirector(ctx context.Context, domainsDir, adaptersDir, outDir string, componentDirs []string, check func(*model.DomainSpec) error) (*Director, error) {
	cat, adapters, err := loadInputs(domainsDir, adaptersDir)
	if err != nil {
		return nil, err
	}
	library, err := component.NewLibrary(componentDirs)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	d := NewDirector(ctx, cat, adapters, outDir)
	d.library, d.check = library, check
	return d, nil
}

func loadInputs(domainsDir, adaptersDir string) (*domain.Catalog, map[string]*model.Adapter, error) {
	cat, err := domain.NewCatalogFromDir(domainsDir)
	if err != nil {
		return nil, nil, fmt.Errorf("streamsim: %w", err)
	}
	adapters, err := adapter.LoadDir(adaptersDir)
	if err != nil {
		return nil, nil, fmt.Errorf("streamsim: %w", err)
	}
	return cat, adapters, nil
}
