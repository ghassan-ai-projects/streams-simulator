package world

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
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

// Acknowledge is what the consumer learns from an invocation: its
// acknowledgement, without the failure mode, effect flag or latency.
func Acknowledge(result *InvokeResult) *model.EffectorAcknowledgement {
	return layer.Acknowledge(result)
}

// KnownFailureMode reports whether mode is in the failure-mode vocabulary.
func KnownFailureMode(mode FailureMode) bool {
	return layer.KnownFailureMode(mode)
}

func PopulationIDs(spec *domain.Compiled, firstSetCount int) []string {
	return layer.PopulationIDs(spec, firstSetCount)
}
