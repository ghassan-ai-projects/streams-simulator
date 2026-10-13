package domain

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

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
