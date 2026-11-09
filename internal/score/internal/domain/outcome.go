package domain

import (
	"errors"
	"math"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

var ErrOutcomeRuns = errors.New("score: an outcome needs the as_run and the do_nothing run")

func Outcome(weights *model.OutcomeWeights, runs []model.OutcomeRun) (*model.OutcomeScore, error) {
	asRun, nothing := runIndex(runs, model.OutcomeAsRun), runIndex(runs, model.OutcomeDoNothing)
	if asRun < 0 || nothing < 0 {
		return nil, ErrOutcomeRuns
	}
	clean := isClean(runs[nothing].Facts)
	scored := &model.OutcomeScore{Valid: !clean, Runs: slices.Clone(runs)}
	if weights == nil {
		return scored, nil
	}
	costRuns(scored, weights, asRun, clean)
	return scored, nil
}

func costRuns(scored *model.OutcomeScore, weights *model.OutcomeWeights, asRun int, clean bool) {
	for i := range scored.Runs {
		scored.Runs[i].Cost = runCost(weights, scored.Runs[i].Facts, clean)
	}
	tie := weights.Tie
	scored.Tie, scored.Verdicts = &tie, verdicts(scored.Runs, asRun, tie)
	scored.Easy = scored.Valid && slices.ContainsFunc(scored.Verdicts, matchedByAPolicy)
}

func matchedByAPolicy(v model.OutcomeVerdict) bool {
	_, isPolicy := v.Baseline.Policy()
	return isPolicy && v.Verdict != model.VerdictBeats
}

func runIndex(runs []model.OutcomeRun, mode model.OutcomeMode) int {
	return slices.IndexFunc(runs, func(r model.OutcomeRun) bool { return r.Mode == mode })
}

func isClean(facts *model.OutcomeFacts) bool {
	return !slices.ContainsFunc(facts.Entities, func(e model.EntityOutcome) bool { return len(e.Trips) > 0 || len(e.Boundaries) > 0 })
}

func verdicts(runs []model.OutcomeRun, asRun int, tie float64) []model.OutcomeVerdict {
	var judged []model.OutcomeVerdict
	for i, baseline := range runs {
		if i != asRun {
			delta := runs[asRun].Cost.J - baseline.Cost.J
			judged = append(judged, model.OutcomeVerdict{Baseline: baseline.Mode, Delta: delta, Verdict: verdictOf(delta, tie)})
		}
	}
	return judged
}

func verdictOf(delta, tie float64) string {
	switch {
	case math.Abs(delta) <= tie:
		return model.VerdictTies
	case delta < 0:
		return model.VerdictBeats
	}
	return model.VerdictLoses
}
