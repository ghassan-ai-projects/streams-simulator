package app

import (
	"context"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
)

// LoadDirector builds the director over every domain in domainsDir and every
// adapter in adaptersDir; its runs publish evidence under outDir.
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
