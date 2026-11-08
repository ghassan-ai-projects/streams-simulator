package contract

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/score"
)

func (c storyCheck) scoreOutcome(policy string) {
	c.t.Helper()
	nothing, asRun := c.run(policyNone).facts, c.run(policyNone).facts
	if policy == policyReference {
		asRun = c.run(policyReference).facts
	}
	weights := beltPlant(c.t, c.story.Plant).Spec.GroundTruth.OutcomeWeights
	scored, err := score.Outcome(weights, []model.OutcomeRun{{Mode: model.OutcomeAsRun, Facts: asRun}, {Mode: model.OutcomeDoNothing, Facts: nothing}})
	if err != nil {
		c.t.Fatal(err)
	}
	verdict := scored.Verdicts[0]
	c.t.Logf("outcome %s seed %.0f %s: J %.4f, do nothing J %.4f, delta %.4f (%s), valid %v", c.story.ID, c.row["seed"], policy,
		scored.Runs[0].Cost.J, scored.Runs[1].Cost.J, verdict.Delta, verdict.Verdict, scored.Valid)
	switch {
	case c.story.NegativeClass && (scored.Valid || verdict.Verdict != model.VerdictTies):
		c.errorf("outcome: valid %v and %s; a negative story leaves doing nothing clean and ties", scored.Valid, verdict.Verdict)
	case !c.story.NegativeClass && !scored.Valid:
		c.errorf("outcome: doing nothing is clean, so the story tests nothing")
	case policy == policyReference && verdict.Verdict != model.VerdictBeats:
		c.errorf("outcome: the reference policy %s doing nothing by %.4f", verdict.Verdict, verdict.Delta)
	}
}
