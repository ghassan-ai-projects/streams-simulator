package domain

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

// Evidence is everything the scorer reads about one run: the submitted
// verdict, the delivery ledger, the director-side effector log, the applied
// perturbations, the emitted count and the hidden-state history, plus the
// domain whose faults give the recovery levels. The scorer never reaches
// back into the run that produced it.
type Evidence struct {
	RunID         string
	Domain        *domain.Compiled
	Verdict       *model.Verdict
	Ledger        []model.LedgerRecord
	Calls         []world.EffectorCall
	Perturbations []string
	Emitted       int64
	History       []model.StateSnapshot
	Reproducible  bool
	Unblinded     bool
}
