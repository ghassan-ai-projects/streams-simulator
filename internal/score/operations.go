package score

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	layer "github.com/ghassan-ai-projects/streams-simulator/internal/score/internal/domain"
)

func Score(ev Evidence, gt *model.GroundTruthRecord) (*Scorecard, error) {
	return layer.Score(ev, gt)
}

func ScoreRun(ev Evidence, gts []*model.GroundTruthRecord) (*RunScore, error) {
	return layer.ScoreRun(ev, gts)
}

func Document(graded *RunScore) any {
	return layer.Document(graded)
}
