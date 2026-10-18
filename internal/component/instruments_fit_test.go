package component

import (
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"testing"
)

func bound(value float64) *float64 { return &value }

func fitOf(t *testing.T, library *Library, unit string, low, high *float64) map[string]Fit {
	t.Helper()
	listed, err := library.ListFor(unit, low, high)
	if err != nil {
		t.Fatal(err)
	}
	fits := map[string]Fit{}
	for _, entry := range listed {
		if entry.Fit == nil {
			t.Fatalf("entry %+v carries no fit", entry)
		}
		fits[entry.ID] = *entry.Fit
	}
	return fits
}

func TestAPressureTransmitterInKilopascalFitsAReadingInPascalThatItsRangeCovers(t *testing.T) {
	t.Parallel()
	library := builtInLibrary(t)
	cases := []struct {
		name      string
		low, high *float64
		want      Fit
	}{
		{"a reading inside the range after conversion", bound(0), bound(500000), Fit{OK: true}},
		{"a reading above the range after conversion", bound(0), bound(7000000), Fit{Code: FitRangeNarrower, Member: "eu_range"}},
		{"a reading that states no range", nil, nil, Fit{Code: FitRangeUnstated, Member: "eu_range"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := fitOf(t, library, "Pa", tc.low, tc.high)["streamsim/pressure-transmitter"]; got != tc.want {
				t.Fatalf("pressure-transmitter fit = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestAFlowMeterInCubicMetresPerHourFitsAReadingInLitresPerMinuteOnlyWhenTheReadingStatesItsRange(t *testing.T) {
	t.Parallel()
	library := builtInLibrary(t)
	with := fitOf(t, library, "L/min", bound(0), bound(2000))
	without := fitOf(t, library, "L/min", nil, nil)
	for _, id := range []string{"streamsim/mag-flow-meter", "streamsim/vortex-flow-meter"} {
		if with[id] != (Fit{OK: true}) || without[id] != (Fit{Code: FitRangeUnstated, Member: "eu_range"}) {
			t.Errorf("%s: with a range %+v, without %+v, want it fitting only with the range", id, with[id], without[id])
		}
	}
}

func TestTheFitOfEachInstrumentIsListedAfterTheOnesThatFitAndSerialisedWithItsStableCodes(t *testing.T) {
	t.Parallel()
	listed, err := builtInLibrary(t).ListFor("%", bound(0), bound(100))
	if err != nil {
		t.Fatal(err)
	}
	order := make([]string, len(listed))
	for i, entry := range listed {
		order[i] = entry.ID
	}
	if want := []string{"streamsim/valve-position", "streamsim/humidity"}; !slices.Equal(order, want) {
		t.Fatalf("order = %v, want the exact valve position before the convertible humidity", order)
	}
	raw, err := json.Marshal(listed[0])
	if err != nil || !strings.Contains(string(raw), `"fit":{"ok":true}`) || !strings.Contains(string(raw), `"match":"exact"`) {
		t.Fatalf("entry JSON = %s, %v, want the fit as {ok:true} beside the match", raw, err)
	}
}

func TestARangeThatIsNoRangeIsRefusedBeforeAnyInstrumentIsJudged(t *testing.T) {
	t.Parallel()
	library := builtInLibrary(t)
	for _, bounds := range [][2]*float64{{bound(1), nil}, {nil, bound(1)}, {bound(5), bound(5)}} {
		if listed, err := library.ListFor("kPa", bounds[0], bounds[1]); !errors.Is(err, ErrReadingRange) || listed != nil {
			t.Errorf("ListFor with %v = %v, %v, want ErrReadingRange", bounds, listed, err)
		}
	}
}
