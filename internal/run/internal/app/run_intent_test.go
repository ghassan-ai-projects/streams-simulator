package app

import (
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestARunWithoutIntentRecordsNeitherPlanNorTermination(t *testing.T) {
	t.Parallel()
	scenario := advanceScenario(1)
	art := executedArtifact(t, hookConfig(t, scenario), scenario)
	if art.Plan != nil || art.Termination != nil {
		t.Fatalf("plan = %+v, termination = %+v, want both absent", art.Plan, art.Termination)
	}
}

func TestARunRecordsThePlanAndTerminationItWasGiven(t *testing.T) {
	t.Parallel()
	scenario := advanceScenario(2)
	r := hookRun(t, scenario)
	end := model.DefaultStartTimeNS + 600e9
	if err := r.SetPlan(model.RunPlan{RequestedEndNS: end, Rate: 60}); err != nil {
		t.Fatal(err)
	}
	at := model.DefaultStartTimeNS + 120e9
	if err := r.SetTermination(model.Termination{Reason: model.TerminationStoppedByOperator, AtNS: at}); err != nil {
		t.Fatal(err)
	}
	if err := r.ExecuteScenario(t.Context(), scenario); err != nil {
		t.Fatal(err)
	}
	art, err := r.End("")
	if err != nil {
		t.Fatal(err)
	}
	if art.Plan == nil || *art.Plan != (model.RunPlan{RequestedEndNS: end, Rate: 60}) {
		t.Fatalf("plan = %+v", art.Plan)
	}
	if art.Termination == nil || *art.Termination != (model.Termination{Reason: model.TerminationStoppedByOperator, AtNS: at}) {
		t.Fatalf("termination = %+v", art.Termination)
	}
}

func TestTheIntentSettersRefuseWhatTheContractDoesNotAllow(t *testing.T) {
	t.Parallel()
	scenario := advanceScenario(1)
	r := hookRun(t, scenario)
	for _, plan := range []model.RunPlan{{RequestedEndNS: 0}, {RequestedEndNS: -1}, {RequestedEndNS: 1, Rate: -1}} {
		if err := r.SetPlan(plan); err == nil || !strings.Contains(err.Error(), "run: plan ") {
			t.Errorf("SetPlan(%+v) = %v, want a refusal naming the plan", plan, err)
		}
	}
	if err := r.SetTermination(model.Termination{Reason: "walked_away"}); err == nil || !strings.Contains(err.Error(), `reason "walked_away"`) {
		t.Errorf("SetTermination of an unknown reason = %v, want the reason named", err)
	}
	if _, err := r.End(""); err != nil {
		t.Fatal(err)
	}
	if err := r.SetTermination(model.Termination{Reason: model.TerminationCompleted}); err == nil || !strings.Contains(err.Error(), "after the run finished") {
		t.Errorf("SetTermination after End = %v, want a refusal", err)
	}
}
