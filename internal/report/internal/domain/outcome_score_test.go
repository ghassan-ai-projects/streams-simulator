package domain

import (
	"encoding/json"
	"slices"
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
	result := &model.OutcomeReport{OutcomeScore: model.OutcomeScore{Valid: true,
		Runs: []model.OutcomeRun{{Mode: model.OutcomeAsRun, Facts: facts, Cost: &model.OutcomeCost{J: -1.25}},
			{Mode: model.OutcomeDoNothing, Facts: &model.OutcomeFacts{}}},
		Verdicts: []model.OutcomeVerdict{{Baseline: model.OutcomeDoNothing, Delta: -5.5, Verdict: model.VerdictBeats}}}}
	lines := outcomeScoreLines(&OutcomeScoring{Status: Scored, Result: result})
	for _, want := range []string{"| Status | scored |", "| Valid | yes |", "| `as_run` | -1.2500 | 3 | 2 | 1 |",
		"| `do_nothing` | - | 0 | 0 | 0 |", "| `do_nothing` | -5.5000 | beats |"} {
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
