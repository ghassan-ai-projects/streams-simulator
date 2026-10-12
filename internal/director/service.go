package director

import (
	"context"

	layer "github.com/ghassan-ai-projects/streams-simulator/internal/director/internal/app"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// Director holds the director-role state: the catalog, the installed
// adapters, the world registry and the sealed truth store.
type Director struct {
	impl *layer.Director
}

// New builds the director with its registries. outDir is where world runs
// publish their evidence.
func New(ctx context.Context, cat *domain.Catalog, adapters map[string]*model.Adapter, outDir string) *Director {
	return &Director{impl: layer.NewDirector(ctx, cat, adapters, outDir)}
}

// NewFromDirs builds the director over every domain in domainsDir and every
// adapter in adaptersDir; its runs publish evidence under outDir.
func NewFromDirs(ctx context.Context, domainsDir, adaptersDir, outDir string) (*Director, error) {
	impl, err := layer.LoadDirector(ctx, domainsDir, adaptersDir, outDir)
	if err != nil {
		return nil, err
	}
	return &Director{impl: impl}, nil
}
