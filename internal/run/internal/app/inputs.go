package app

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func LoadInputs(domainsDir, adaptersDir, domainID, adapterID string) (*domain.Compiled, *model.Adapter, error) {
	cat, err := domain.NewCatalogFromDir(domainsDir)
	if err != nil {
		return nil, nil, fmt.Errorf("streamsim: %w", err)
	}
	adapters, err := adapter.LoadDir(adaptersDir)
	if err != nil {
		return nil, nil, fmt.Errorf("streamsim: %w", err)
	}
	return selectInputs(cat, adapters, domainID, adapterID)
}

func selectInputs(cat *domain.Catalog, adapters map[string]*model.Adapter, domainID, adapterID string) (*domain.Compiled, *model.Adapter, error) {
	spec, err := cat.Describe(domainID)
	if err != nil {
		return nil, nil, fmt.Errorf("streamsim: %w", err)
	}
	adap, err := adapter.Lookup(adapters, adapterID)
	if err != nil {
		return nil, nil, fmt.Errorf("streamsim: %w", err)
	}
	return spec, adap, nil
}
