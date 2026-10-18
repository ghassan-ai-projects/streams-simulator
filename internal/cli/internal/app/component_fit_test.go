package app

import (
	"path/filepath"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func fitOfEntry(t *testing.T, args []string, id string) map[string]any {
	t.Helper()
	doc := invoke(t, args...).mustSucceed(t).document(t, "component_list")
	entries, _ := doc["components"].([]any)
	for _, raw := range entries {
		if entry, _ := raw.(map[string]any); entry["id"] == id {
			fit, _ := entry["fit"].(map[string]any)
			return fit
		}
	}
	t.Fatalf("component list %v lists no %s: %v", args, id, doc)
	return nil
}

func TestComponentListByUnitAndRangeStatesWhetherEachInstrumentFitsTheReading(t *testing.T) {
	t.Parallel()
	const transmitter = "streamsim/pressure-transmitter"
	cases := []struct {
		name string
		args []string
		ok   bool
		code any
	}{
		{"a reading in Pa inside the range after conversion", []string{"component", "list", "--unit", "Pa", "--min", "0", "--max", "500000"}, true, nil},
		{"a reading in Pa above the range after conversion", []string{"component", "list", "--unit", "Pa", "--min", "0", "--max", "7000000"}, false, "range_narrower"},
		{"a reading with a negative lower bound", []string{"component", "list", "--unit", "kPa", "--min", "-50", "--max", "100"}, true, nil},
		{"a reading that states no range", []string{"component", "list", "--unit", "Pa"}, false, "range_unstated"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			fit := fitOfEntry(t, tc.args, transmitter)
			if fit["ok"] != tc.ok || fit["code"] != tc.code {
				t.Fatalf("fit of %s = %v, want ok %v and code %v", transmitter, fit, tc.ok, tc.code)
			}
		})
	}
}

func TestComponentListRefusesARangeThatIsNoRangeOrHasNoUnit(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		args []string
		code string
		exit int
	}{
		{"a minimum alone", []string{"component", "list", "--unit", "kPa", "--min", "0"}, "usage_invalid", 2},
		{"a minimum above the maximum", []string{"component", "list", "--unit", "kPa", "--min", "9", "--max", "1"}, "usage_invalid", 2},
		{"a range without its unit", []string{"component", "list", "--min", "0", "--max", "1"}, "usage_invalid", 2},
		{"a range that is not a number", []string{"component", "list", "--unit", "kPa", "--min", "NaN", "--max", "1"}, "usage_invalid", 2},
		{"a range on another verb", []string{"component", "show", "streamsim/rtd", "--min", "0", "--max", "1"}, "usage_invalid", 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := invoke(t, tc.args...)
			if got.code != tc.exit || got.failure(t)["code"] != tc.code {
				t.Fatalf("exit %d with %v, want exit %d and %s", got.code, got.failure(t), tc.exit, tc.code)
			}
		})
	}
}

func TestKindTwinSuggestNamesOnlyTheKindFilesItIsGiven(t *testing.T) {
	t.Parallel()
	fan := filepath.Join(testsupport.ThermalBoardKindsDir(), "fan.kind.json")
	broken := writeFile(t, "pump_monitor.kind.json", `{"kind": "pump_monitor"}`)
	doc := invoke(t, "kind", "twin", "--suggest", fan).mustSucceed(t).document(t, "kind_suggest")
	choices, _ := doc["choices"].(map[string]any)
	kinds, _ := choices["kinds"].([]any)
	if choices["id"] != "fan-twin" || choices["title"] != "Twin of fan" || len(kinds) != 1 {
		t.Fatalf("kind twin --suggest of the fan kind file = %v, want a twin of the fan alone", choices)
	}
	failure := invoke(t, "kind", "twin", "--suggest", fan, broken).failure(t)
	if failure["code"] != "kind_failed" {
		t.Fatalf("suggest with a malformed kind file in the list = %v, want kind_failed", failure)
	}
}
