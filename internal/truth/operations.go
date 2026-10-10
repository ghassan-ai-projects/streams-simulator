package truth

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	layer "github.com/ghassan-ai-projects/streams-simulator/internal/truth/internal/domain"
)

// BuildRecord assembles one sealed label for a scenario.
func BuildRecord(spec *domain.Compiled, solver *Solver, in Injection) (*model.GroundTruthRecord, error) {
	if solver == nil {
		return nil, ErrNoSolver
	}
	return layer.BuildRecord(spec, solver.solver, in)
}

// Seal records the label for a run and seals it.
func (s *Store) Seal(runID string, rec *model.GroundTruthRecord) error {
	return s.store.Seal(runID, rec)
}

// Reveal returns the sealed label. unblind permits revealing on an open
// run, permanently stamping it.
func (s *Store) Reveal(runID string, unblind bool) (*model.GroundTruthRecord, error) {
	return s.store.Reveal(runID, unblind)
}

// SealStatus reports the sealing state of a run: an unknown run is an error;
// for a known run sealed is always true and unblinded tells whether a reveal
// stamped it.
func (s *Store) SealStatus(runID string) (sealed, unblinded bool, err error) {
	return s.store.SealStatus(runID)
}
