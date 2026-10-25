package domain

import (
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func richPair() (*Report, *Report) {
	a, b := sealedRun(7), sealedRun(8)
	b.Run.Name, b.Run.Purpose, b.Run.Rate = "Fan failure", "Does the setpoint help?", 60
	b.Truth.Faults[0].ClearedNS = pointer(ns(200))
	b.Truth.Faults[0].ClearedBy = &[]string{"command"}[0]
	b.Truth.Perturbations[0].Params = map[string]any{"rate": 0.9}
	b.Truth.Perturbations = append(b.Truth.Perturbations, Perturbation{Seq: 5, AtNS: ns(150), Action: "clear", PerturbID: "p-0"})
	a.Truth.Environment = []Environment{{Seq: 3, AtNS: ns(30), Target: "site", Fault: "heat_wave", Params: map[string]any{"delta": 3.0}}}
	b.Truth.EffectorCalls[0].Outcome = &Outcome{Refusal: "interlock"}
	b.Truth.EffectorCalls = append(b.Truth.EffectorCalls, EffectorCall{AtNS: ns(300), EntityID: "pump-1", Effector: "stop"})
	b.Truth.KPIs.Entities[0].KPIs = []model.KPIValue{{State: "cop", Mean: 4, Min: 3, Max: 5, Samples: 60}, {State: "flow", Mean: 1, Min: 1, Max: 1, Samples: 60}}
	b.Truth.KPIs.Entities[0].Alarms = nil
	b.Truth.KPIs.Entities = append(b.Truth.KPIs.Entities, model.EntityKPIs{EntityID: "pump-1", Alarms: []model.AlarmKPI{{State: "trip", Priority: "low", Activations: 1}}})
	return a, b
}

func TestTheComparisonMarkdownListsInputsLimitsPlanAndResultsWithoutRanking(t *testing.T) {
	t.Parallel()
	a, b := richPair()
	text := string(ComparisonMarkdown(Compare(a, b)))
	for _, want := range []string{
		"# Run comparison: r-1 and r-1\n\nRun A is the reference. This document lists what differs; it does not rank the runs.",
		"| Seed | 7 | 8 |", "| Name |  | Fan failure |", "| Pace |  | 60 simulated s per wall s |",
		"- `different_seed`: The runs use different seeds (7 and 8)", "- `single_run_per_side`: ",
		"### Faults", "#### Different", "| Run | Entity | Fault | Onset | Cleared | Differs in |", "cleared\\_by, cleared\\_ns",
		"### Perturbations", "| clear | p-0 |", "### Environment", "#### Only in A", "heat\\_wave",
		"### Effector calls", "refused: interlock", "#### Only in B",
		"Scope: whole run of each run. Simulated length covered by KPI truth: A 00:10:00, B 00:10:00.",
		"| chiller-1 | cop | both | 5.000 | 4.000 | -1.000 | -20.0% | 4.000..6.000 | 3.000..5.000 |",
		"| chiller-1 | flow | only\\_in\\_b |  | 1.000 |",
		"| chiller-1 | high\\_temp | only\\_in\\_a | 2 |  |", "| pump-1 | trip | only\\_in\\_b |  | 1 |",
		"| A | 2 | 2 | 0 | 0 | 0 | 0 |", "| Delta | 0 | 0 | 0 | 0 | 0 | 0 |",
		"## Windowed results\n\nnot_available: ",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("markdown lacks %q:\n%s", want, text)
		}
	}
	for _, forbidden := range []string{"winner", "better", "worse", "best"} {
		if strings.Contains(strings.ToLower(text), forbidden) {
			t.Errorf("markdown ranks the runs with %q", forbidden)
		}
	}
	if !strings.HasSuffix(text, "\n") || strings.HasSuffix(text, "\n\n") {
		t.Errorf("markdown must end in exactly one newline:\n%q", text[len(text)-30:])
	}
}

func TestPairedItemsOfEveryKindNameTheFieldsThatDiffer(t *testing.T) {
	t.Parallel()
	a, b := richPair()
	a.Truth.Environment = []Environment{{Seq: 3, AtNS: ns(30), Target: "site", Fault: "heat_wave", Params: map[string]any{"delta": 3.0}}}
	b.Truth.Environment = []Environment{{Seq: 3, AtNS: ns(30), Target: "site", Fault: "heat_wave", Params: map[string]any{"delta": 5.0}}}
	c := Compare(a, b)
	if got := c.Plan.Environment.Different[0].Differences; len(got) != 1 || got[0] != "params" {
		t.Errorf("environment differences = %v, want params", got)
	}
	if got := c.Plan.Perturbations.Different[0].Differences; len(got) != 1 || got[0] != "params" {
		t.Errorf("perturbation differences = %v, want params", got)
	}
	if got := c.Plan.EffectorCalls.Different[0].Differences; len(got) != 1 || got[0] != "outcome" {
		t.Errorf("call differences = %v, want outcome", got)
	}
	if len(c.Plan.Perturbations.OnlyInB) != 1 || c.Plan.Perturbations.OnlyInB[0].PerturbID != "p-0" {
		t.Errorf("perturbations = %+v, want the clear only in B", c.Plan.Perturbations)
	}
}

func TestAComparisonWithoutKPITruthRendersNoneAndAReportWithoutTruthCountsAsEmpty(t *testing.T) {
	t.Parallel()
	a, b := sealedRun(7), sealedRun(7)
	a.Truth.KPIs, b.Truth.KPIs = nil, nil
	text := string(ComparisonMarkdown(Compare(a, b)))
	for _, want := range []string{"### KPIs\n\nNone.", "### Alarms\n\nNone.", "### Alarm system\n\nNone.", "A none, B none"} {
		if !strings.Contains(text, want) {
			t.Errorf("markdown lacks %q:\n%s", want, text)
		}
	}
	untruthed := Compare(&Report{Run: Run{Mode: "explore"}}, &Report{Run: Run{Mode: "explore"}})
	if len(untruthed.Plan.Faults.Identical)+len(untruthed.Results.KPIs) != 0 || untruthed.A.KPITruth {
		t.Fatalf("a report without truth compared as %+v, want an empty plan and no results", untruthed)
	}
	if got := idOf(nil) + digestOf(nil) + requestedSeconds(RunInputs{}); got != "unknownnot recorded" {
		t.Errorf("unknown references and an absent request read as %q", got)
	}
}

func TestEncodingAComparisonFailsOnlyForValuesJSONCannotHold(t *testing.T) {
	t.Parallel()
	a, b := sealedRun(7), sealedRun(7)
	b.Truth.Environment = []Environment{{Params: map[string]any{"bad": make(chan int)}}}
	if _, err := ComparisonJSON(Compare(a, b)); err == nil || !strings.Contains(err.Error(), "encode comparison") {
		t.Fatalf("err = %v, want the encode failure named", err)
	}
	if _, err := JSON(&Report{Truth: &Truth{Environment: []Environment{{Params: map[string]any{"bad": make(chan int)}}}}}); err == nil || !strings.Contains(err.Error(), "report: encode") {
		t.Fatalf("err = %v, want the report encode failure named", err)
	}
}

func TestTheHandoverNamesAnUnendedRunAnIncompleteRunAndARequestedEnd(t *testing.T) {
	t.Parallel()
	open := &Report{Run: Run{Mode: "explore"}, Verification: Verification{Status: NotChecked}}
	if got := executionOutcome(open.Run); got != "Not ended: no run.json" {
		t.Errorf("open run outcome = %q", got)
	}
	incomplete := Run{Ended: true, IncompleteReason: "the sink failed"}
	if got := executionOutcome(incomplete); got != "Ended incomplete: the sink failed" {
		t.Errorf("incomplete outcome = %q", got)
	}
	plain := Run{Ended: true, Complete: true, StartTimeNS: startNS, EndTimeNS: ns(60)}
	if got := executionOutcome(plain); !strings.HasPrefix(got, "Ended at ") {
		t.Errorf("outcome without a termination = %q", got)
	}
	planned := plain
	planned.Termination, planned.RequestedEndNS = &model.Termination{Reason: model.TerminationCompleted, AtNS: ns(60)}, ns(60)
	if got := executionOutcome(planned); !strings.Contains(got, "Completed at ") || !strings.Contains(got, ", requested end ") {
		t.Errorf("planned outcome = %q", got)
	}
}
