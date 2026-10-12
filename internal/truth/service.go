package truth

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	layer "github.com/ghassan-ai-projects/streams-simulator/internal/truth/internal/domain"
)

type Solver struct {
	solver *layer.Solver
}

func NewSolver(spec *domain.Compiled, seed uint64, sampleNS, maxHorizonNS int64) (*Solver, error) {
	if spec == nil {
		return nil, ErrNoSpec
	}
	return &Solver{solver: layer.NewSolver(spec, seed, sampleNS, maxHorizonNS)}, nil
}

type Store struct {
	store *layer.Store
}

func NewStore(runIsOpen func(runID string) bool) *Store {
	return &Store{store: layer.NewStore(runIsOpen)}
}
