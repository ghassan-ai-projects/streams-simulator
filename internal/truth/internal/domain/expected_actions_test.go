package domain

import (
	"encoding/json"
	"os"
	"slices"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func tankLeakExpecting(t *testing.T, actions ...any) *domain.Compiled {
	t.Helper()
	raw, err := os.ReadFile(testsupport.TwoKindDomain())
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	tank := doc["kinds"].([]any)[0].(map[string]any)
	tank["faults"].([]any)[0].(map[string]any)["expected_actions"] = actions
	if raw, err = json.Marshal(doc); err != nil {
		t.Fatal(err)
	}
	spec, err := domain.Parse(raw, "two-kind.json")
	if err != nil {
		t.Fatal(err)
	}
	return spec
}

func tankLeakLabel(t *testing.T, spec *domain.Compiled, fault string) *model.GroundTruthRecord {
	t.Helper()
	_, ids := twoKindFixture(t)
	start := model.DefaultStartTimeNS
	rec, err := BuildRecord(spec, NewSolver(spec, 42, 60*1e9, 6*3600*1e9), Injection{ScenarioID: "two-kind/0001", Seed: 42,
		EntityID: "tank-2", FaultID: fault, OnsetNS: start + 3600*1e9, StartNS: start, EntityIDs: ids})
	if err != nil {
		t.Fatal(err)
	}
	return rec
}

func TestALabelExpectsAnActionOfItsOwnKindOnTheFaultedInstanceAndOfAnotherKindOnEachOfItsInstances(t *testing.T) {
	t.Parallel()
	spec := tankLeakExpecting(t, map[string]any{"kind": "pump", "effector": "boost"}, map[string]any{"kind": "tank", "effector": "refill"})
	want := []model.ExpectedAction{{EntityID: "pump-1", Effector: "boost"}, {EntityID: "pump-2", Effector: "boost"},
		{EntityID: "pump-3", Effector: "boost"}, {EntityID: "pump-4", Effector: "boost"}, {EntityID: "tank-2", Effector: "refill"}}
	if got := tankLeakLabel(t, spec, "tank_leak").ExpectedActions; !slices.Equal(got, want) {
		t.Fatalf("expected actions = %v, want %v", got, want)
	}
	if got := tankLeakLabel(t, spec, "tank_nuisance").ExpectedActions; got != nil {
		t.Fatalf("a fault that declares no expected action labels %v", got)
	}
}

func TestASealedLabelKeepsItsExpectedActionsApartFromTheCaller(t *testing.T) {
	t.Parallel()
	rec := &model.GroundTruthRecord{ExpectedActions: []model.ExpectedAction{{EntityID: "pump-1", Effector: "boost"}}}
	out := cloneRecord(rec)
	rec.ExpectedActions[0].EntityID = "pump-9"
	if out.ExpectedActions[0].EntityID != "pump-1" {
		t.Fatalf("clone = %v, want its own copy of the expected actions", out.ExpectedActions)
	}
}
