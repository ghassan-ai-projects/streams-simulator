package deviceworld

import (
	layer "github.com/ghassan-ai-projects/streams-simulator/internal/deviceworld/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

type Plant struct {
	plant *layer.Plant
}

func New(w *world.World, bindings map[string]Binding) (*Plant, error) {
	if w == nil {
		return nil, ErrNoWorld
	}
	return &Plant{plant: layer.New(w, bindings)}, nil
}
