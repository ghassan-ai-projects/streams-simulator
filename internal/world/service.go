package world

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	layer "github.com/ghassan-ai-projects/streams-simulator/internal/world/internal/domain"
)

type World struct {
	ID      string
	Spec    *domain.Compiled
	StartNS int64

	world *layer.World
}

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

func Acknowledge(result *InvokeResult) *model.EffectorAcknowledgement {
	return layer.Acknowledge(result)
}

func PopulationIDs(spec *domain.Compiled, firstSetCount int) []string {
	return layer.PopulationIDs(spec, firstSetCount)
}
