package domain

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/randutil"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func (r headerReference) supply(t float64) float64 {
	unlinked := newLagReference([]lagLink{{"avg", 70, 50, "u"}}, 7)
	var supply, from float64
	for ; from+600 <= t; from += 600 {
		supply = lagStep(supply, max(r.average(from), unlinked.value("avg", from)), 900, 600)
	}
	if t > from {
		supply = lagStep(supply, max(r.average(from), unlinked.value("avg", from)), 900, t-from)
	}
	return supply
}

func TestAThreeLevelChainAgreesWithTheReference(t *testing.T) {
	t.Parallel()
	r := newHeaderReference(map[string]float64{"unit-1": 10, "unit-2": 10, "unit-3": 10}, nil)
	for _, s := range []float64{600, 1234, 3600, 7200} {
		w := coupledWorld(t)
		near(t, "plant supply", w.StateValue("plant-1", "supply", atSeconds(s)), r.supply(s))
	}
	w := coupledWorld(t)
	readSteps(w, "unit-2", "flow", 60, 7200)
	near(t, "plant supply after the units ran ahead", w.StateValue("plant-1", "supply", atSeconds(7200)), r.supply(7200))
}

type coupledRead struct {
	entity, state string
}

var coupledReads = []coupledRead{
	{"unit-1", "flow"}, {"unit-3", "flow"}, {"hdr-a", "inflow"}, {"hdr-a", "level"}, {"hdr-a", "avg_flow"},
	{"hdr-a", "units_running"}, {"hdr-b", "avg_flow"}, {"plant-1", "supply"},
}

func TestRandomForwardReadsAgreeWithAWorldReadOnce(t *testing.T) {
	t.Parallel()
	rng := randutil.NewSplitMix64(17)
	w := coupledWorld(t)
	if _, err := w.InjectFault("unit-2", "derate", atSeconds(1500), nil); err != nil {
		t.Fatal(err)
	}
	at := 0.0
	for range 120 {
		at += rng.Float64() * 120
		read := coupledReads[rng.Intn(len(coupledReads))]
		fresh := coupledWorld(t)
		if _, err := fresh.InjectFault("unit-2", "derate", atSeconds(1500), nil); err != nil {
			t.Fatal(err)
		}
		if got, want := w.StateValue(read.entity, read.state, atSeconds(at)), fresh.StateValue(read.entity, read.state, atSeconds(at)); got != want {
			t.Fatalf("%s.%s at %v s = %v after the schedule, %v read once", read.entity, read.state, at, got, want)
		}
	}
}

func relayWorld(t *testing.T) *World {
	t.Helper()
	raw, err := os.ReadFile(testsupport.CoupledDomain())
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	header := doc["kinds"].([]any)[1].(map[string]any)
	dynamics := header["dynamics"].([]any)
	header["dynamics"] = append(append([]any{}, dynamics[:2]...), dynamics[4:]...)
	plant := doc["kinds"].([]any)[2].(map[string]any)
	input := plant["dynamics"].([]any)[0].(map[string]any)["f1"].(map[string]any)["inputs"].([]any)[0].(map[string]any)
	input["state"] = "inflow"
	if raw, err = json.Marshal(doc); err != nil {
		t.Fatal(err)
	}
	spec, err := domain.Parse(raw, "relay.json")
	if err != nil {
		t.Fatal(err)
	}
	w, err := New(spec, 3, "w-relay", model.DefaultStartTimeNS, Options{EmitDisabled: true})
	if err != nil {
		t.Fatal(err)
	}
	return w
}

func TestARelayWithNoHistoryOfItsOwnKeepsItsSourcesHistoryForItsReaders(t *testing.T) {
	t.Parallel()
	r := newHeaderReference(map[string]float64{"unit-1": 10, "unit-2": 10, "unit-3": 10}, nil)
	var supply float64
	for from := 0.0; from < 7200; from += 600 {
		supply = lagStep(supply, max(r.sum(from), 5), 900, 600)
	}
	w := relayWorld(t)
	readSteps(w, "unit-1", "flow", 60, 7200)
	near(t, "plant supply over a relay", w.StateValue("plant-1", "supply", atSeconds(7200)), supply)
}
