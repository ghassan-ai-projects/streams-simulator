package domain

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func churningHeader(t *testing.T) *domain.Compiled {
	t.Helper()
	raw, err := os.ReadFile(testsupport.CoupledDomain())
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	units := doc["instances"].([]any)[0].(map[string]any)
	units["churn"] = map[string]any{"births_per_hour": 20, "mean_lifetime_s": 900}
	for unit := 4; unit <= 40; unit++ {
		doc["links"] = append(doc["links"].([]any), map[string]any{"from": fmt.Sprintf("unit-%d", unit), "to": "hdr-a", "type": "feed"})
	}
	header := doc["kinds"].([]any)[1].(map[string]any)
	header["faults"] = append(header["faults"].([]any), map[string]any{"id": "inflow_loss",
		"affects":       []any{map[string]any{"state": "inflow", "delta": -0.0003, "multiplicative": true}},
		"observability": map[string]any{"detector": map[string]any{"form": "single_channel_snr", "channel": "header.level"}, "first_observable_snr": 3.0, "unavoidable_snr": 10.0},
		"onset":         map[string]any{"magnitude": 1.0, "shape": "step"}})
	if raw, err = json.Marshal(doc); err != nil {
		t.Fatal(err)
	}
	spec, err := domain.Parse(raw, "churning-header.json")
	if err != nil {
		t.Fatal(err)
	}
	return spec
}

func churnLabel(t *testing.T, spec *domain.Compiled, solverSeed, worldSeed uint64) int64 {
	t.Helper()
	start := model.DefaultStartTimeNS
	rec, err := BuildRecord(spec, NewSolver(spec, solverSeed, 10e9, 6*3600e9), Injection{ScenarioID: "churn", Seed: worldSeed,
		EntityID: "hdr-a", FaultID: "inflow_loss", OnsetNS: start + 1800e9, StartNS: start,
		EntityIDs: []string{"unit-1", "unit-2", "unit-3", "hdr-a", "hdr-b", "plant-1"}})
	if err != nil {
		t.Fatal(err)
	}
	return rec.FirstObservableTimeNS
}

func TestTheOracleWorldsReplayTheChurnOfTheScenarioWorldSeed(t *testing.T) {
	t.Parallel()
	spec := churningHeader(t)
	if churnLabel(t, spec, 1, 1) == churnLabel(t, spec, 2, 2) {
		t.Fatal("the fixture's label does not depend on the churn of its seed, so it cannot tell which seed the oracle used")
	}
	if got, want := churnLabel(t, spec, 1, 2), churnLabel(t, spec, 2, 2); got != want {
		t.Fatalf("a solver seeded 1 labels a scenario of world seed 2 at %d, the scenario's own seed gives %d", got, want)
	}
}
