package perturb

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	layer "github.com/ghassan-ai-projects/streams-simulator/internal/perturb/internal/domain"
)

type Layer struct {
	layer *layer.Layer
}

func New(worldID string, seed uint64, spec *domain.Compiled) (*Layer, error) {
	if spec == nil {
		return nil, ErrNoSpec
	}
	return &Layer{layer: layer.New(worldID, seed, spec)}, nil
}
