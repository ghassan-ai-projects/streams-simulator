package score

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	layer "github.com/ghassan-ai-projects/streams-simulator/internal/score/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

// Score evaluates one run's evidence against its sealed label. A run with no
// submitted verdict is refused.
func Score(ev Evidence, gt *model.GroundTruthRecord) (*Scorecard, error) {
	return layer.Score(ev, gt)
}

// Offline scores from artifacts alone: a verdict, the sealed label, the
// delivery ledger and, when available, the effector calls and applied
// perturbations. It shares the online scorer's policies for every metric
// both can compute.
func Offline(v *model.Verdict, gt *model.GroundTruthRecord, ledger []model.LedgerRecord, calls []world.EffectorCall, perturbations []string) *Scorecard {
	return layer.Offline(v, gt, ledger, calls, perturbations)
}
