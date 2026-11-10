package domain

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestARunDeclaresAnOutcomeOnlyWhenItsEmbeddedDomainWeighsOne(t *testing.T) {
	t.Parallel()
	for name, tc := range map[string]struct {
		artifact *model.RunArtifact
		want     bool
	}{
		"no artifact":        {nil, false},
		"no embedded domain": {&model.RunArtifact{}, false},
		"no outcome weights": {&model.RunArtifact{DomainSpec: json.RawMessage(`{"ground_truth":{"negative_class_fraction":0.4}}`)}, false},
		"outcome weights":    {&model.RunArtifact{DomainSpec: json.RawMessage(`{"ground_truth":{"outcome_weights":{"tie":0.1}}}`)}, true},
		"an unreadable spec": {&model.RunArtifact{DomainSpec: json.RawMessage(`[1]`)}, false},
	} {
		if got := DeclaresOutcome(tc.artifact); got != tc.want {
			t.Errorf("%s: declares %v, want %v", name, got, tc.want)
		}
	}
}

func TestTheOutcomeScoreSectionShowsEveryRunAndBaseline(t *testing.T) {
	t.Parallel()
	facts := &model.OutcomeFacts{Entities: []model.EntityOutcome{
		{EntityID: "B-1", Trips: []model.TripOutcome{{State: "stranded", Count: 2}}, Boundaries: []model.BoundaryOutcome{{Name: "stranded", Violations: 2}}},
		{EntityID: "D-1", Trips: []model.TripOutcome{{State: "inverter_ok", Count: 1}}}},
		Calls: []model.OutcomeCall{{Effector: "set_feed"}}}
	result := &model.OutcomeReport{OutcomeScore: model.OutcomeScore{Valid: true, Easy: true,
		Runs: []model.OutcomeRun{{Mode: model.OutcomeAsRun, Facts: facts, Cost: &model.OutcomeCost{J: -1.25}},
			{Mode: model.OutcomeDoNothing, Facts: &model.OutcomeFacts{}}, {Mode: model.PolicyOutcome("b1"), Facts: &model.OutcomeFacts{}}},
		Verdicts: []model.OutcomeVerdict{{Baseline: model.OutcomeDoNothing, Delta: -5.5, Verdict: model.VerdictBeats},
			{Baseline: model.PolicyOutcome("b1"), Delta: 0.05, Verdict: model.VerdictTies}}}}
	scoring := &OutcomeScoring{Status: Scored, Unit: "equivalent production hours", Labels: map[string]string{"policy:b1": "vibration alarm"}, Result: result}
	lines := outcomeScoreLines(scoring)
	for _, want := range []string{"**Summary:** The run beats doing nothing (-5.5000); it ties b1 (0.0500).",
		"| Status | scored |", "| Cost unit | equivalent production hours |",
		"| Easy | yes: a baseline policy ties or beats the run, so the story does not tell the run from that fixed rule |",
		"| Run | Policy | Cost J (equivalent production hours) | Trips | Boundary violations | Calls |",
		"| `as_run` | the run as it ran | -1.2500 | 3 | 2 | 1 |", "| `do_nothing` | nobody acts | - | 0 | 0 | 0 |",
		"| `do_nothing` | nobody acts | -5.5000 | beats |", "| `policy:b1` | vibration alarm | 0.0500 | ties |"} {
		if !slices.Contains(lines, want) {
			t.Errorf("section %q lacks %q", lines, want)
		}
	}
	if got := outcomeScoreLines(&OutcomeScoring{Status: Failed, Reason: "the run does not replay"}); !slices.Contains(got, "| Reason | the run does not replay |") || len(got) != 4 {
		t.Fatalf("a failed outcome = %q, want its status and reason only", got)
	}
	if outcomeScoreLines(nil) != nil {
		t.Fatal("a report without an outcome score has an outcome score section")
	}
}

func TestTheOutcomeScoreSaysWhatTheCostsMeanAndWhatKindOfStoryRan(t *testing.T) {
	t.Parallel()
	tie := 0.1
	control := &OutcomeScoring{Status: Scored, Result: &model.OutcomeReport{OutcomeScore: model.OutcomeScore{Tie: &tie,
		Runs:     []model.OutcomeRun{{Mode: model.OutcomeAsRun, Facts: &model.OutcomeFacts{}}},
		Verdicts: []model.OutcomeVerdict{{Baseline: model.OutcomeDoNothing, Verdict: model.VerdictTies}, {Baseline: model.PolicyOutcome("b3"), Delta: 1, Verdict: model.VerdictLoses}}}}}
	got := strings.Join(outcomeScoreLines(control), "\n")
	for _, want := range []string{"The run ties doing nothing (0.0000); it loses b3 (1.0000).",
		"Costs are J in the domain's own cost unit (the domain does not name it). A delta is the run's J minus the baseline's: a negative delta means the run cost less. Two costs within 0.1 of each other tie.",
		"| Story | Control story: doing nothing is clean, so the story tests no reaction and an unneeded call only costs (`valid: no`) |",
		"| Easy | no: a control story is never easy |", "| Run | Policy | Cost J | Trips |"} {
		if !strings.Contains(got, want) {
			t.Errorf("section:\n%s\nlacks %q", got, want)
		}
	}
	if strings.Contains(got, "Cost unit") || strings.Contains(got, "invalid") {
		t.Errorf("a control story without a unit reads:\n%s", got)
	}
	if text := easyText(true, false); text != "no: the run beats every baseline policy" {
		t.Errorf("a valid story that is not easy reads %q", text)
	}
}

func TestAScoredOutcomeCarriesTheDomainsCostUnitAndPolicyLabels(t *testing.T) {
	t.Parallel()
	artifact := &model.RunArtifact{DomainSpec: json.RawMessage(`{"ground_truth":{"outcome_weights":{"tie":0.1,"unit":"equivalent production hours"},` +
		`"baseline_policies":[{"name":"b1","label":"vibration alarm","when":{},"then":[]},{"name":"b2","when":{},"then":[]}]}}`)}
	scoring := ScoredOutcome(&model.OutcomeReport{}, artifact)
	if scoring.Status != Scored || scoring.Unit != "equivalent production hours" || len(scoring.Labels) != 1 || scoring.Labels["policy:b1"] != "vibration alarm" {
		t.Fatalf("scoring = %+v, want the unit and b1's label only", scoring)
	}
	if bare := ScoredOutcome(&model.OutcomeReport{}, &model.RunArtifact{DomainSpec: json.RawMessage(`[1]`)}); bare.Unit != "" || bare.Labels != nil {
		t.Fatalf("an unreadable domain gives %+v, want no words", bare)
	}
}
