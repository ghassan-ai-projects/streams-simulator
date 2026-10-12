package score

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	layer "github.com/ghassan-ai-projects/streams-simulator/internal/score/internal/domain"
)

// Score evaluates one run's evidence against its sealed label. A run with no
// submitted verdict is refused.
func Score(ev Evidence, gt *model.GroundTruthRecord) (*Scorecard, error) {
	return layer.Score(ev, gt)
}
