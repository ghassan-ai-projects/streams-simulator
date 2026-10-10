package world

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	layer "github.com/ghassan-ai-projects/streams-simulator/internal/world/internal/domain"
)

// World is one seeded simulated world. ID, Spec and StartNS are fixed at
// construction.
type World struct {
	ID      string
	Spec    *domain.Compiled
	StartNS int64

	world *layer.World
}

// New builds a world from a compiled domain, a seed, an id and a start time.
// A nil spec is refused with ErrNoSpec.
func New(spec *domain.Compiled, seed uint64, id string, startNS int64, opts Options) (*World, error) {
	if spec == nil {
		return nil, ErrNoSpec
	}
	inner, err := layer.New(spec, seed, id, startNS, opts)
	if err != nil {
		return nil, err
	}
	return &World{ID: inner.ID, Spec: inner.Spec, StartNS: inner.StartNS, world: inner}, nil
}

// RenderID renders one entity id from the domain's id template, for hosts
// that must enumerate a domain's entity ids without building a world.
func RenderID(tmpl string, n int) string {
	return layer.RenderID(tmpl, n)
}
