package app

import (
	"context"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
)

func LoadDirector(ctx context.Context, domainsDir, adaptersDir, outDir string) (*Director, error) {
	cat, err := domain.NewCatalogFromDir(domainsDir)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	adapters, err := adapter.LoadDir(adaptersDir)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return NewDirector(ctx, cat, adapters, outDir), nil
}
