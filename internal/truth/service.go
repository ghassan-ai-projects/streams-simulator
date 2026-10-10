package truth

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	layer "github.com/ghassan-ai-projects/streams-simulator/internal/truth/internal/domain"
)

// Solver computes the observability timestamps of a fault by running the
// world twice — once clean, once faulted, both noiseless — and finding when
// the detector quantity crosses the declared SNR thresholds.
type Solver struct {
	solver *layer.Solver
}

// NewSolver builds a solver. sampleNS is the resolution of the onset scan;
// maxHorizonNS caps how far the scan looks before declaring a fault
// unobservable. A nil spec is refused with ErrNoSpec.
func NewSolver(spec *domain.Compiled, seed uint64, sampleNS, maxHorizonNS int64) (*Solver, error) {
	if spec == nil {
		return nil, ErrNoSpec
	}
	return &Solver{solver: layer.NewSolver(spec, seed, sampleNS, maxHorizonNS)}, nil
}

// Store holds the sealed truth for runs under the director role. It is
// deliberately a separate object from any operator-facing view.
type Store struct {
	store *layer.Store
}

// NewStore builds an empty truth store. runIsOpen reports whether a run is
// still open; Reveal refuses an open run unless it is unblinded. A nil check
// counts every run as open, so a store built without one refuses to reveal
// rather than leaking the label.
func NewStore(runIsOpen func(runID string) bool) *Store {
	return &Store{store: layer.NewStore(runIsOpen)}
}
