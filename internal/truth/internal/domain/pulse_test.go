package domain

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func sagOnTwoChannels(t *testing.T) *domain.Compiled {
	t.Helper()
	raw, err := os.ReadFile(testsupport.UtilityDomain())
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	for _, k := range doc["kinds"].([]any) {
		if kind := k.(map[string]any); kind["name"] == "chiller" {
			addPeriodicSag(kind)
		}
	}
	if raw, err = json.Marshal(doc); err != nil {
		t.Fatal(err)
	}
	spec, err := domain.Parse(raw, "sag.json")
	if err != nil {
		t.Fatal(err)
	}
	return spec
}

func addPeriodicSag(kind map[string]any) {
	var sag, event map[string]any
	for _, f := range kind["faults"].([]any) {
		if fault := f.(map[string]any); fault["id"] == "voltage_sag" {
			sag = fault
		}
	}
	for _, c := range kind["channels"].([]any) {
		if channel := c.(map[string]any); channel["name"] == "chiller.voltage_event" {
			event = channel
		}
	}
	periodic := map[string]any{}
	for key, value := range event {
		periodic[key] = value
	}
	periodic["name"], periodic["cadence"] = "chiller.voltage", map[string]any{"mode": "periodic", "period_s": 60}
	kind["channels"] = append(kind["channels"].([]any), periodic)
	periodicSag := map[string]any{"id": "voltage_sag_periodic", "affects": sag["affects"], "onset": sag["onset"],
		"observability": map[string]any{"detector": map[string]any{"form": "single_channel_snr", "channel": "chiller.voltage"},
			"first_observable_snr": 3.0, "unavoidable_snr": 10.0}}
	kind["faults"] = append(kind["faults"].([]any), periodicSag)
}

func TestAPulseIsObservableOnlyWhereItsDetectorChannelRecordsIt(t *testing.T) {
	t.Parallel()
	spec := sagOnTwoChannels(t)
	start := model.DefaultStartTimeNS
	onset := start + 600_400_000_000
	ids := []string{"meter-1", "meter-2", "CH1", "AIR1"}
	solver := NewSolver(spec, 1, 60e9, 2*3600e9)
	recorded, err := solver.solve("CH1", "voltage_sag", onset, start, ids, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !recorded.Observable || recorded.FirstObservableNS != onset {
		t.Fatalf("the event-driven channel records the sag at its start: %+v, want observable at %d", recorded, onset)
	}
	unseen, err := solver.solve("CH1", "voltage_sag_periodic", onset, start, ids, nil)
	if err != nil {
		t.Fatal(err)
	}
	if unseen.Observable {
		t.Fatalf("a 60 s periodic channel records nothing inside a 0.3 s sag at 600.4 s, yet the label is %+v", unseen)
	}
}
