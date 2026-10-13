package truth

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	layer "github.com/ghassan-ai-projects/streams-simulator/internal/truth/internal/domain"
)

func BuildRecord(spec *domain.Compiled, solver *Solver, in Injection) (*model.GroundTruthRecord, error) {
	if solver == nil {
		return nil, ErrNoSolver
	}
	return layer.BuildRecord(spec, solver.solver, in)
}

func (s *Store) Seal(runID string, rec *model.GroundTruthRecord) error {
	return s.store.Seal(runID, rec)
}

func (s *Store) Reveal(runID string, unblind bool) (*model.GroundTruthRecord, error) {
	return s.store.Reveal(runID, unblind)
}

func (s *Store) SealStatus(runID string) (sealed, unblinded bool, err error) {
	return s.store.SealStatus(runID)
}
