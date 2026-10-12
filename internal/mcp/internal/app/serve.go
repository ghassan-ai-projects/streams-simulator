package app

import (
	"context"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
)

// ServeOptions say which director to serve: where its domains and adapters
// are loaded from, where its runs publish evidence, and the address that
// binds the operator role over streamable HTTP ("" binds none).
type ServeOptions struct {
	DomainsDir, AdaptersDir, OutDir, OperatorAddr string
}

// LoadDirector builds the director over every domain in options.DomainsDir
// and every adapter in options.AdaptersDir.
func LoadDirector(ctx context.Context, options ServeOptions) (*Director, error) {
	cat, err := domain.NewCatalogFromDir(options.DomainsDir)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	adapters, err := adapter.LoadDir(options.AdaptersDir)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return NewDirector(ctx, cat, adapters, options.OutDir), nil
}
