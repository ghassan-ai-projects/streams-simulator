package component

import (
	"encoding/json"
	"testing"
)

func TestListForAUnitReturnsTheInstrumentsOfTheSameOrAConvertibleUnitExactFirst(t *testing.T) {
	t.Parallel()
	library := builtInLibrary(t)
	cases := []struct {
		name, unit string
		want       []string
	}{
		{"Cel is convertible to the degC of both temperature instruments", "Cel", []string{"streamsim/rtd:degC:convertible", "streamsim/thermistor:degC:convertible"}},
		{"degC is their own unit", "degC", []string{"streamsim/rtd:degC:exact", "streamsim/thermistor:degC:exact"}},
		{"% is a valve position exactly and a humidity by conversion", "%", []string{"streamsim/valve-position:%:exact", "streamsim/humidity:%RH:convertible"}},
		{"an unknown unit matches nothing", "furlong", []string{}},
		{"a unit of another family matches that family only", "kPa", []string{"streamsim/pressure-transmitter:kPa:exact"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			listed := library.ListFor(tc.unit)
			got := make([]string, len(listed))
			for i, entry := range listed {
				got[i] = entry.ID + ":" + entry.Unit + ":" + entry.Match
				if entry.Kind != "instrument" || entry.Range == nil || entry.Range.Min >= entry.Range.Max {
					t.Errorf("entry %+v, want an instrument with its range", entry)
				}
			}
			if !jsonEqual(got, tc.want) {
				t.Fatalf("ListFor(%q) = %v, want %v", tc.unit, got, tc.want)
			}
		})
	}
}

func TestListForNoUnitIsTheWholeLibraryAndEveryInstrumentStatesItsUnitAndRange(t *testing.T) {
	t.Parallel()
	library := builtInLibrary(t)
	all := library.ListFor("")
	if len(all) != len(library.List()) {
		t.Fatalf("ListFor(\"\") lists %d entries, want the %d of List", len(all), len(library.List()))
	}
	for _, entry := range all {
		if (entry.Kind == "instrument") != (entry.Unit != "") || entry.Match != "" {
			t.Errorf("entry %+v: an instrument states its unit, equipment does not, and no match without a unit to match", entry)
		}
	}
	raw, err := json.Marshal(all)
	if err != nil || len(raw) == 0 {
		t.Fatal(err)
	}
	if (*Library)(nil).ListFor("Cel") != nil || (&Library{}).ListFor("Cel") != nil {
		t.Fatal("a missing library lists nothing")
	}
}
