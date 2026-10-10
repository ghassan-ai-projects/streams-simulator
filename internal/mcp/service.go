package mcp

import (
	"context"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	layer "github.com/ghassan-ai-projects/streams-simulator/internal/mcp/internal/app"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// Director is the director role's state: the catalog, adapters, the world
// registry and the sealed truth store.
type Director struct {
	director *layer.Director
}

// NewDirector builds the director with its registries. outDir is where world
// runs publish their evidence.
func NewDirector(ctx context.Context, cat *domain.Catalog, adapters map[string]*model.Adapter, outDir string) *Director {
	return &Director{director: layer.NewDirector(ctx, cat, adapters, outDir)}
}
