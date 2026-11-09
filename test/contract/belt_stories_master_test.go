package contract

import (
	"math"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

var lineFollowers = []string{"D-2", "D-3"}

func (c storyCheck) survivors() []string {
	if c.story.Story == "S6" {
		return lineFollowers
	}
	return survivors
}

func (c storyCheck) masterLoss(policy string) {
	c.t.Helper()
	o, trip := c.run(policy), c.field("trip_s")
	if c.story.Unavoidable {
		c.lineStops(o, policy, trip)
		return
	}
	if lead := o.at("D-2.speed_lead"); lead != trip+referenceSampleS+promoteS || !math.IsInf(o.at("D-2.enabled"), 1) {
		c.errorf("%s: D-2 leads at %.0f s (closed form %.0f s), stopped at %.0f s", policy, lead, trip+referenceSampleS+promoteS, o.at("D-2.enabled"))
	}
	if policy == policyReference {
		c.saved(o)
		return
	}
	c.strandsAtTheDownstreamStop(o)
}

// lineStops checks the stop of the whole line after its master is lost: the stall and the stranded boundary at their
// closed-form times, no protective trip, the belt stranded to the horizon and never restarted (one stall).
func (c storyCheck) lineStops(o storyOutcome, policy string, trip float64) {
	c.t.Helper()
	_, stall, _, stranded := stopAllClock(c.story.regime())
	if o.at("B-1.stall_ok") != trip+stall || o.at("B-1.stranded") != trip+math.Ceil(stranded) {
		c.errorf("%s: stall at %.0f s and stranded at %.0f s, closed form %.0f s and %.0f s", policy, o.at("B-1.stall_ok"), o.at("B-1.stranded"), trip+stall, trip+math.Ceil(stranded))
	}
	if when, what := o.firstOf(storyDrives, storyProtections); !math.IsInf(when, 1) {
		c.errorf("%s: %s tripped at %.0f s", policy, what, when)
	}
	held := c.story.HorizonS - o.at("B-1.stranded")
	if stalls, boundary := tripCount(o.facts, "B-1", "stall_ok"), boundaryOf(o.facts, "B-1", "stranded"); stalls != 1 || boundary.Violations != 1 || boundary.DurationS < held-1 {
		c.errorf("%s: %d stalls, stranded %d times for %.0f s; want one stall and the belt stranded for the last %.0f s", policy, stalls, boundary.Violations, boundary.DurationS, held)
	}
}

func boundaryOf(facts *model.OutcomeFacts, entity, name string) model.BoundaryOutcome {
	at := slices.IndexFunc(facts.Entities, func(e model.EntityOutcome) bool { return e.EntityID == entity })
	boundaries := facts.Entities[at].Boundaries
	if found := slices.IndexFunc(boundaries, func(b model.BoundaryOutcome) bool { return b.Name == name }); found >= 0 {
		return boundaries[found]
	}
	return model.BoundaryOutcome{}
}

// harmless checks a story no reaction can save: its reference does no harm, so it ties doing nothing exactly (a call
// costs only on a story whose do-nothing run is clean).
func (c storyCheck) harmless(verdict model.OutcomeVerdict) {
	c.t.Helper()
	if verdict.Verdict != model.VerdictTies || verdict.Delta != 0 {
		c.errorf("outcome: the reference %s doing nothing by %.4f; want an exact tie", verdict.Verdict, verdict.Delta)
	}
}
