package director

import (
	"context"

	layer "github.com/ghassan-ai-projects/streams-simulator/internal/director/internal/app"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

type Director struct {
	impl *layer.Director
}

func New(ctx context.Context, cat *domain.Catalog, adapters map[string]*model.Adapter, outDir string) *Director {
	return &Director{impl: layer.NewDirector(ctx, cat, adapters, outDir)}
}

func NewFromDirs(ctx context.Context, domainsDir, adaptersDir, outDir string, componentDirs []string, check func(*model.DomainSpec) error) (*Director, error) {
	impl, err := layer.LoadDirector(ctx, domainsDir, adaptersDir, outDir, componentDirs, check)
	if err != nil {
		return nil, err
	}
	return &Director{impl: impl}, nil
}
