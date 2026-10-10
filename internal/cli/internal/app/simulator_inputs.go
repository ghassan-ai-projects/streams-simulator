package app

import (
	"fmt"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func loadSimulatorInputs(domainsDir, adaptersDir, domainID, adapterID string) (*domain.Compiled, *model.Adapter, error) {
	cat, err := loadCatalog(domainsDir)
	if err != nil {
		return nil, nil, fmt.Errorf("streamsim: %w", err)
	}
	adapters, err := loadAdapters(adaptersDir)
	if err != nil {
		return nil, nil, fmt.Errorf("streamsim: %w", err)
	}
	return describeSimulatorInputs(cat, adapters, domainID, adapterID)
}

func describeSimulatorInputs(cat *domain.Catalog, adapters map[string]*model.Adapter, domainID, adapterID string) (*domain.Compiled, *model.Adapter, error) {
	spec, err := cat.Describe(domainID)
	if err != nil {
		return nil, nil, fmt.Errorf("streamsim: %w", err)
	}
	adap, ok := adapters[adapterID]
	if !ok {
		return nil, nil, fmt.Errorf("unknown adapter %q", adapterID)
	}
	return spec, adap, nil
}

func loadDomainSpec(dir, id string) (*domain.Compiled, error) {
	cat, err := loadCatalog(dir)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	spec, err := cat.Describe(id)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return spec, nil
}
