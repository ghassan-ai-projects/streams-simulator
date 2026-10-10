package perturb

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	layer "github.com/ghassan-ai-projects/streams-simulator/internal/perturb/internal/domain"
)

// Layer applies active perturbations to a native event stream.
type Layer struct {
	layer *layer.Layer
}

// New builds a perturbation layer for a world. The domain spec is needed
// only for contract knowledge (declared enums, ranges, units, attacker-
// controlled channels); consumer knowledge never enters here. A nil spec is
// refused with ErrNoSpec.
func New(worldID string, seed uint64, spec *domain.Compiled) (*Layer, error) {
	if spec == nil {
		return nil, ErrNoSpec
	}
	return &Layer{layer: layer.New(worldID, seed, spec)}, nil
}
