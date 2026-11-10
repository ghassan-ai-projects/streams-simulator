package domain

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func instrumentedPlant(t *testing.T, flowSigma float64) *domain.Compiled {
	t.Helper()
	raw, err := os.ReadFile(testsupport.CoupledDomain())
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	unit := doc["kinds"].([]any)[0].(map[string]any)
	for _, channel := range unit["channels"].([]any) {
		if channel := channel.(map[string]any); channel["name"] == "unit.flow" {
			channel["noise"].(map[string]any)["sigma"] = flowSigma
		}
	}
	fault := func(id string, observation map[string]any) map[string]any {
		return map[string]any{"id": id, "onset": map[string]any{"shape": "step", "magnitude": 1.0}, "observation": observation,
			"observability": map[string]any{"detector": map[string]any{"form": "single_channel_snr", "channel": "unit.flow"}, "first_observable_snr": 3.0, "unavoidable_snr": 10.0}}
	}
	unit["faults"] = append(unit["faults"].([]any),
		fault("flow_dropout", map[string]any{"channel": "unit.flow", "form": "dropout"}),
		fault("flow_frozen", map[string]any{"channel": "unit.flow", "form": "freeze"}))
	if flowSigma > 0 {
		unit["faults"] = append(unit["faults"].([]any), fault("flow_noisy", map[string]any{"channel": "unit.flow", "form": "noise", "factor": 5.0}))
	}
	if raw, err = json.Marshal(doc); err != nil {
		t.Fatal(err)
	}
	spec, err := domain.Parse(raw, "instrumented.json")
	if err != nil {
		t.Fatal(err)
	}
	return spec
}

func TestInstrumentFaultsAreObservableByWhatTheyDoToTheChannel(t *testing.T) {
	t.Parallel()
	spec := instrumentedPlant(t, 0.05)
	ids := []string{"unit-1", "unit-2", "unit-3", "hdr-a", "hdr-b", "plant-1"}
	start := model.DefaultStartTimeNS
	onset := start + 600e9
	cases := []struct {
		fault       string
		unavoidable bool
	}{{"flow_dropout", true}, {"flow_noisy", false}, {"flow_frozen", true}}
	for _, tc := range cases {
		res, err := NewSolver(spec, 42, 60e9, 3600e9).solve("unit-1", tc.fault, onset, start, ids, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !res.Observable || res.FirstObservableNS != onset {
			t.Fatalf("%s: observable %v first at %d, want at onset %d", tc.fault, res.Observable, res.FirstObservableNS, onset)
		}
		if res.unavoidableFound != tc.unavoidable {
			t.Fatalf("%s: unavoidable found %v, want %v (a fivefold noise is SNR 4)", tc.fault, res.unavoidableFound, tc.unavoidable)
		}
	}
}

func TestAFrozenReadingOfANoiselessChannelIsNotObservableAtItsInjection(t *testing.T) {
	t.Parallel()
	start := model.DefaultStartTimeNS
	onset := start + 600e9
	res, err := NewSolver(instrumentedPlant(t, 0), 42, 60e9, 3600e9).solve("unit-1", "flow_frozen", onset, start, []string{"unit-1", "unit-2", "unit-3", "hdr-a", "hdr-b", "plant-1"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.Observable && res.FirstObservableNS == onset {
		t.Fatal("a frozen reading of a channel without noise loses nothing at its injection, yet it is observable there")
	}
}
