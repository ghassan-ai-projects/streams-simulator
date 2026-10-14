package domain

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"slices"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
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
		"onset":         map[string]any{"magnitude": 1.0, "shape": "step"}},
		map[string]any{"id": "inflow_dip",
			"affects":       []any{map[string]any{"state": "inflow", "delta": -0.0003, "multiplicative": true}},
			"observability": map[string]any{"detector": map[string]any{"form": "single_channel_snr", "channel": "header.level"}, "first_observable_snr": 3.0, "unavoidable_snr": 10.0},
			"onset":         map[string]any{"magnitude": 1.0, "shape": "pulse", "duration_s": 3600}})
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

var churnIDs = []string{"unit-1", "unit-2", "unit-3", "hdr-a", "hdr-b", "plant-1"}

func runStyleWorld(t *testing.T, spec *domain.Compiled, seed uint64) *world.World {
	t.Helper()
	w, err := world.New(spec, seed, "", model.DefaultStartTimeNS, world.Options{Noiseless: true, EmitDisabled: true, InitialEntities: churnIDs})
	if err != nil {
		t.Fatal(err)
	}
	return w
}

func TestTheOracleWorldIsBuiltAsARunBuildsItsWorld(t *testing.T) {
	t.Parallel()
	spec := churningHeader(t)
	oracle, err := NewSolver(spec, 1, 60e9, 3600e9).forWorldSeed(7).oracleWorld(model.DefaultStartTimeNS, churnIDs, false)
	if err != nil {
		t.Fatal(err)
	}
	run := runStyleWorld(t, spec, 7)
	for _, w := range []*world.World{oracle, run} {
		if _, _, err := w.Advance(model.DefaultStartTimeNS + 3*3600e9); err != nil {
			t.Fatal(err)
		}
	}
	if got, want := oracle.EntityIDs(), run.EntityIDs(); !slices.Equal(got, want) {
		t.Fatalf("the oracle world births %v, a run of the same seed %v", got, want)
	}
}

func referencePulseLabel(t *testing.T, spec *domain.Compiled, seed uint64, onset int64) int64 {
	t.Helper()
	clean, faulted := runStyleWorld(t, spec, seed), runStyleWorld(t, spec, seed)
	if _, err := faulted.InjectFault("hdr-a", "inflow_dip", onset, nil); err != nil {
		t.Fatal(err)
	}
	for at := onset; at <= onset+6*3600e9; at += 60e9 {
		for _, w := range []*world.World{clean, faulted} {
			if _, _, err := w.Advance(at); err != nil {
				t.Fatal(err)
			}
		}
		if math.Abs(faulted.Reading("hdr-a", "header.level", at)-clean.Reading("hdr-a", "header.level", at)) >= 3*0.05 {
			return at
		}
	}
	return 0
}

func TestAPulseInAChurningDomainIsLabelledFromWorldsAtTheSameInstant(t *testing.T) {
	t.Parallel()
	spec := churningHeader(t)
	onset := model.DefaultStartTimeNS + 1800e9
	for _, seed := range []uint64{1, 2} {
		rec, err := BuildRecord(spec, NewSolver(spec, 99, 10e9, 6*3600e9), Injection{ScenarioID: "dip", Seed: seed,
			EntityID: "hdr-a", FaultID: "inflow_dip", OnsetNS: onset, StartNS: model.DefaultStartTimeNS, EntityIDs: churnIDs})
		if err != nil {
			t.Fatal(err)
		}
		if want := referencePulseLabel(t, spec, seed, onset); rec.FirstObservableTimeNS != want || want <= onset {
			t.Fatalf("seed %d: first observable at %d, the reference at %d (onset %d)", seed, rec.FirstObservableTimeNS, want, onset)
		}
	}
}
