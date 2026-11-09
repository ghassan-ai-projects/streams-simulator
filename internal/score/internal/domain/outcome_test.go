package domain

import (
	"errors"
	"reflect"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func conveyorWeights(tie float64) *model.OutcomeWeights {
	return &model.OutcomeWeights{Tie: tie,
		Boundaries: []model.BoundaryWeight{{Kind: "belt", Name: "stranded", PerViolation: 4, PerHour: 2}},
		Trips:      []model.TripWeight{{Kind: "drive", State: "inverter_ok", PerTrip: 0.5}},
		KPIs:       []model.KPIWeight{{Kind: "belt", State: "delivered", PerUnit: -0.01}, {Kind: "drive", State: "life_used", PerUnit: 0.5}},
		Calls:      []model.CallWeight{{Effector: "set_feed", PerCall: 0.25}}}
}

func belt(delivered float64, boundaries ...model.BoundaryOutcome) model.EntityOutcome {
	return model.EntityOutcome{EntityID: "B-1", Kind: "belt", Boundaries: boundaries, KPIs: []model.KPIOutcome{{State: "delivered", Start: 1000, End: 1000 + delivered}}}
}

func drive(id string, used float64, trips ...model.TripOutcome) model.EntityOutcome {
	return model.EntityOutcome{EntityID: id, Kind: "drive", Trips: trips, KPIs: []model.KPIOutcome{{State: "life_used", Start: 10, End: 10 + used}}}
}

func facts(calls []model.OutcomeCall, entities ...model.EntityOutcome) *model.OutcomeFacts {
	if calls == nil {
		calls = []model.OutcomeCall{}
	}
	return &model.OutcomeFacts{HorizonNS: 1, Entities: entities, Calls: calls}
}

var (
	stranded     = model.BoundaryOutcome{Name: "stranded", Severity: "major", Violations: 1, FirstNS: 5, DurationS: 5400}
	inverterTrip = model.TripOutcome{State: "inverter_ok", FirstNS: 3, Count: 2}
	feedCall     = model.OutcomeCall{Effector: "set_feed", EntityID: "F-1", AtNS: 2, Accepted: true}
	refusedCall  = model.OutcomeCall{Effector: "set_feed", EntityID: "F-1", AtNS: 3}
)

func TestTheCostOfARunIsTheSumOfItsWeightedTerms(t *testing.T) {
	t.Parallel()
	failing := facts(nil, belt(50, stranded), drive("D-1", 2, inverterTrip))
	scored, err := Outcome(conveyorWeights(0), []model.OutcomeRun{{Mode: model.OutcomeAsRun, Facts: failing}, {Mode: model.OutcomeDoNothing, Facts: failing}})
	if err != nil {
		t.Fatal(err)
	}
	want := []model.CostTerm{
		{Term: model.TermBoundaryHours, Kind: "belt", Name: "stranded", EntityID: "B-1", Quantity: 1.5, Cost: 3},
		{Term: model.TermBoundaryViolations, Kind: "belt", Name: "stranded", EntityID: "B-1", Quantity: 1, Cost: 4},
		{Term: model.TermKPIs, Kind: "belt", Name: "delivered", EntityID: "B-1", Quantity: 50, Cost: -0.5},
		{Term: model.TermKPIs, Kind: "drive", Name: "life_used", EntityID: "D-1", Quantity: 2, Cost: 1},
		{Term: model.TermTrips, Kind: "drive", Name: "inverter_ok", EntityID: "D-1", Quantity: 2, Cost: 1},
	}
	cost := scored.Runs[0].Cost
	if !reflect.DeepEqual(cost.Terms, want) || cost.J != 3+4-0.5+1+1 {
		t.Fatalf("cost %+v, want J 8.5 from %+v", cost, want)
	}
}

func TestCallsCostOnlyWhenDoingNothingWasClean(t *testing.T) {
	t.Parallel()
	calls := []model.OutcomeCall{feedCall, feedCall, refusedCall}
	for _, tc := range []struct {
		name      string
		nothing   *model.OutcomeFacts
		valid     bool
		callsCost float64
	}{
		{"a clean story", facts(nil, belt(50), drive("D-1", 2)), false, 0.5},
		{"a story that fails", facts(nil, belt(50, stranded), drive("D-1", 2)), true, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			scored, err := Outcome(conveyorWeights(0), []model.OutcomeRun{{Mode: model.OutcomeAsRun, Facts: facts(calls, belt(50), drive("D-1", 2))},
				{Mode: model.OutcomeDoNothing, Facts: tc.nothing}})
			if err != nil {
				t.Fatal(err)
			}
			if got := scored.Runs[0].Cost.J - (-0.5 + 1); scored.Valid != tc.valid || got != tc.callsCost {
				t.Fatalf("valid %v, calls cost %v; want %v and %v (two accepted calls at 0.25, the refused one free)", scored.Valid, got, tc.valid, tc.callsCost)
			}
		})
	}
}

func TestEachBaselineIsBeatenTiedOrLostWithinTheTie(t *testing.T) {
	t.Parallel()
	nothing := facts(nil, belt(50, stranded), drive("D-1", 2))
	runs := []model.OutcomeRun{
		{Mode: model.OutcomeAsRun, Facts: facts(nil, belt(40), drive("D-1", 2))},
		{Mode: model.OutcomeDoNothing, Facts: nothing},
		{Mode: model.PolicyOutcome("close"), Facts: facts(nil, belt(45), drive("D-1", 2))},
		{Mode: model.PolicyOutcome("better"), Facts: facts(nil, belt(60), drive("D-1", 1))},
	}
	scored, err := Outcome(conveyorWeights(0.1), runs)
	if err != nil {
		t.Fatal(err)
	}
	want := []model.OutcomeVerdict{
		{Baseline: model.OutcomeDoNothing, Delta: (-0.4 + 1) - (3 + 4 - 0.5 + 1), Verdict: model.VerdictBeats},
		{Baseline: model.PolicyOutcome("close"), Delta: (-0.4 + 1) - (-0.45 + 1), Verdict: model.VerdictTies},
		{Baseline: model.PolicyOutcome("better"), Delta: (-0.4 + 1) - (-0.6 + 0.5), Verdict: model.VerdictLoses},
	}
	if !closeVerdicts(scored.Verdicts, want) || scored.Tie == nil || *scored.Tie != 0.1 || !scored.Valid || !scored.Easy {
		t.Fatalf("verdicts %+v (tie %v, valid %v, easy %v), want %+v and easy", scored.Verdicts, scored.Tie, scored.Valid, scored.Easy, want)
	}
}

func TestAStoryIsEasyWhenSomeBaselinePolicyDoesAsWellAsTheRun(t *testing.T) {
	t.Parallel()
	failing, clean := facts(nil, belt(50, stranded), drive("D-1", 2)), facts(nil, belt(50), drive("D-1", 2))
	run := facts(nil, belt(40), drive("D-1", 2))
	for _, tc := range []struct {
		name    string
		nothing *model.OutcomeFacts
		policy  *model.OutcomeFacts
		easy    bool
	}{
		{"every policy beaten", failing, failing, false},
		{"a policy that ties", failing, facts(nil, belt(45), drive("D-1", 2)), true},
		{"a policy that wins", failing, facts(nil, belt(60), drive("D-1", 1)), true},
		{"doing nothing ties on a clean story", clean, failing, false},
		{"a policy that ties on a clean story", clean, run, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			scored, err := Outcome(conveyorWeights(0.1), []model.OutcomeRun{{Mode: model.OutcomeAsRun, Facts: run},
				{Mode: model.OutcomeDoNothing, Facts: tc.nothing}, {Mode: model.PolicyOutcome("b1"), Facts: tc.policy}})
			if err != nil {
				t.Fatal(err)
			}
			if scored.Easy != tc.easy {
				t.Fatalf("easy %v with verdicts %+v (valid %v), want %v", scored.Easy, scored.Verdicts, scored.Valid, tc.easy)
			}
		})
	}
}

func closeVerdicts(got, want []model.OutcomeVerdict) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i].Baseline != want[i].Baseline || got[i].Verdict != want[i].Verdict || got[i].Delta-want[i].Delta > 1e-12 || want[i].Delta-got[i].Delta > 1e-12 {
			return false
		}
	}
	return true
}

func TestADomainWithoutWeightsKeepsItsFactsAndHasNoCost(t *testing.T) {
	t.Parallel()
	runs := []model.OutcomeRun{{Mode: model.OutcomeAsRun, Facts: facts(nil, belt(50))}, {Mode: model.OutcomeDoNothing, Facts: facts(nil, belt(50, stranded))}}
	scored, err := Outcome(nil, runs)
	if err != nil || !scored.Valid || scored.Easy || scored.Tie != nil || scored.Verdicts != nil || scored.Runs[0].Cost != nil || scored.Runs[1].Facts != runs[1].Facts {
		t.Fatalf("score %+v (%v), want the facts, valid, and no cost, tie or verdict", scored, err)
	}
	if _, err := Outcome(nil, runs[:1]); !errors.Is(err, ErrOutcomeRuns) {
		t.Fatalf("err = %v, want ErrOutcomeRuns without a do_nothing run", err)
	}
}
