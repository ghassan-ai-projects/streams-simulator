package domain

import (
	"errors"
	"fmt"
	"math"
	"slices"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func bound(value float64) *float64 { return &value }

func fitLibrary(t *testing.T) *Library {
	t.Helper()
	return libraryWith(t, instrumentLike(t, "acme/zzz-f", "degF", -58, 1000))
}

func fitLines(summaries []Summary) []string {
	lines := make([]string, len(summaries))
	for i, summary := range summaries {
		lines[i] = fmt.Sprintf("%s:%s:%t:%s:%s", summary.ID, summary.Match, summary.Fit.OK, summary.Fit.Code, summary.Fit.Member)
	}
	return lines
}

func TestEveryInstrumentOfAUnitStatesWhetherItFitsTheReadingAndWhyNotFitsLast(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		unit      string
		low, high *float64
		want      []string
	}{
		{"a reading both instruments cover, the exact unit first", "degF", bound(0), bound(200),
			[]string{"acme/zzz-f:exact:true::", "acme/probe:convertible:true::"}},
		{"a reading the converted range of one instrument does not cover", "degF", bound(0), bound(800),
			[]string{"acme/zzz-f:exact:true::", "acme/probe:convertible:false:range_narrower:eu_range"}},
		{"a reading beyond both", "degF", bound(0), bound(2000),
			[]string{"acme/zzz-f:exact:false:range_narrower:eu_range", "acme/probe:convertible:false:range_narrower:eu_range"}},
		{"a reading that states no range", "degC", nil, nil,
			[]string{"acme/probe:exact:false:range_unstated:eu_range", "acme/zzz-f:convertible:false:range_unstated:eu_range"}},
		{"an offset conversion at the edge of the range", "K", bound(223.15), bound(673.15),
			[]string{"acme/probe:convertible:true::", "acme/zzz-f:convertible:true::"}},
		{"a conversion into an instrument that states no range", "W", bound(0), bound(200000), []string{"acme/meter:convertible:true::"}},
		{"an instrument that needs the range of a reading that has none", "kW", nil, nil, []string{"acme/meter:exact:false:range_unstated:eu_range"}},
		{"a unit no family knows", "furlong", bound(0), bound(1), []string{}},
	}
	library := fitLibrary(t)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			listed, err := library.ListForReading(tc.unit, tc.low, tc.high)
			if got := fitLines(listed); err != nil || !slices.Equal(got, tc.want) {
				t.Fatalf("ListForReading(%q) = %v, %v, want %v", tc.unit, got, err, tc.want)
			}
		})
	}
}

func TestAnInstrumentOfAnotherQuantityOrAReadingThatIsNoNumberNamesItsMember(t *testing.T) {
	t.Parallel()
	library := baseLibrary(t)
	probe := library.instruments["acme/probe@1"]
	cases := []struct {
		name         string
		metric       kindMetric
		code, member string
	}{
		{"a unit of another family", kindMetric{DataType: "number", Unit: "kW", EURange: &Range{Min: 0, Max: 1}}, FitUnitIncompatible, "unit"},
		{"a unit no family holds", kindMetric{DataType: "number", Unit: "furlong", EURange: &Range{Min: 0, Max: 1}}, FitUnitIncompatible, "unit"},
		{"a boolean reading", kindMetric{DataType: "boolean"}, FitReadingNotNumber, "data_type"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			verdict := library.fit(probe, tc.metric)
			if verdict.OK || verdict.Code != tc.code || verdict.Member != tc.member || verdict.err() == nil {
				t.Fatalf("fit = %+v, want a refusal %s at %s", verdict, tc.code, tc.member)
			}
		})
	}
}

func TestTheRangeOfAReadingIsBothBoundsInAUnitAndBelowTheMaximum(t *testing.T) {
	t.Parallel()
	library := baseLibrary(t)
	cases := []struct {
		name      string
		unit      string
		low, high *float64
	}{
		{"a minimum alone", "degC", bound(0), nil},
		{"a maximum alone", "degC", nil, bound(10)},
		{"a minimum above the maximum", "degC", bound(10), bound(0)},
		{"an empty range", "degC", bound(5), bound(5)},
		{"a range that is not a number", "degC", bound(math.NaN()), bound(1)},
		{"an infinite range", "degC", bound(0), bound(math.Inf(1))},
		{"a range without its unit", "", bound(0), bound(1)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if listed, err := library.ListForReading(tc.unit, tc.low, tc.high); !errors.Is(err, ErrReadingRange) || listed != nil {
				t.Fatalf("ListForReading = %v, %v, want ErrReadingRange", listed, err)
			}
		})
	}
}

func fitKind(unit string, low, high *float64) string {
	limits := ""
	if low != nil && high != nil {
		limits = fmt.Sprintf(`,"eu_range":{"min":%v,"max":%v}`, *low, *high)
	}
	return fmt.Sprintf(`{"kind":"probe_kind","metrics":[{"name":"reading","data_type":"number","unit":%q%s,"resolution":0.1,"writable":false}],"commands":[]}`, unit, limits)
}

func fitChoices(ref string) string {
	return fmt.Sprintf(`{"schema_version":"0.1","id":"fit-twin","version":"0.1.0","title":"Fit","kinds":[{"kind":"probe_kind","instance":"probe-01","metrics":{"reading":{"instrument":%q}}}]}`, ref)
}

func TestTheFitOfAnInstrumentPredictsExactlyWhatTheTwinScaffoldAccepts(t *testing.T) {
	t.Parallel()
	library := fitLibrary(t)
	readings := []struct {
		unit      string
		low, high *float64
	}{
		{"degF", bound(0), bound(200)}, {"degF", bound(0), bound(800)}, {"degF", bound(0), bound(2000)}, {"degC", nil, nil},
		{"K", bound(223.15), bound(673.15)}, {"K", bound(100), bound(673.15)}, {"degC", bound(0), bound(100)}, {"W", bound(0), bound(200000)}, {"kW", nil, nil},
	}
	for _, reading := range readings {
		listed, err := library.ListForReading(reading.unit, reading.low, reading.high)
		if err != nil || len(listed) == 0 {
			t.Fatalf("ListForReading(%q) = %v, %v", reading.unit, listed, err)
		}
		for _, entry := range listed {
			ref := fmt.Sprintf("%s@%d", entry.ID, entry.Version)
			_, scaffoldErr := ScaffoldTwin(library, kindDocuments(t, fitKind(reading.unit, reading.low, reading.high)), []byte(fitChoices(ref)), "fit.twin.json")
			if entry.Fit.OK {
				if scaffoldErr != nil {
					t.Errorf("%s for %s: fit says ok, the scaffold refuses: %v", ref, reading.unit, scaffoldErr)
				}
				continue
			}
			invalid := asViolation(t, scaffoldErr, model.ContractTwinChoices)
			if len(invalid.Issues) != 1 || invalid.Issues[0].Member != entry.Fit.Member || !errors.Is(scaffoldErr, ErrInstrumentMismatch) {
				t.Errorf("%s for %s: fit says %+v, the scaffold says %v (%+v)", ref, reading.unit, entry.Fit, scaffoldErr, invalid.Issues)
			}
		}
	}
}

func TestATwinReadsThroughAnInstrumentOfAnotherUnitInTheUnitOfItsKindFile(t *testing.T) {
	t.Parallel()
	kinds := heaterKinds(t, func(doc map[string]any) { fileMetric(doc)["unit"] = "degF" })
	_, doc := twinOK(t, kinds, heaterChoicesJSON)
	kind := member(t, doc, "kinds", 0)
	temperature := named(t, member(t, kind, "channels"), "channel", "name", "temperature")
	sigma := number(t, member(t, temperature, "noise", "sigma"))
	if member(t, temperature, "unit") != "degF" || math.Abs(sigma-0.09) > 1e-9 {
		t.Fatalf("temperature channel = %v, want the kind file's degF and the probe's 0.05 degC noise as 0.09 degF", temperature)
	}
	offset := named(t, member(t, kind, "faults"), "fault", "id", "heater__temperature__offset")
	if delta := number(t, member(t, offset, "affects", 0, "delta")); math.Abs(delta-3.6) > 1e-9 {
		t.Fatalf("offset delta = %v, want the probe's 2 degC as 3.6 degF", delta)
	}
	if bias := named(t, member(t, kind, "state"), "state", "name", "temperature__bias"); member(t, bias, "unit") != "degF" {
		t.Fatalf("bias state = %v, want the unit of the kind file", bias)
	}
}

func TestAnUnconvertedInstrumentKeepsItsNoiseAndFaultMagnitudes(t *testing.T) {
	t.Parallel()
	kind := member(t, heaterTwin(t), "kinds", 0)
	offset := named(t, member(t, kind, "faults"), "fault", "id", "heater__temperature__offset")
	if delta := number(t, member(t, offset, "affects", 0, "delta")); delta != 2 {
		t.Fatalf("offset delta = %v, want the probe's 2 degC as it is", delta)
	}
}
