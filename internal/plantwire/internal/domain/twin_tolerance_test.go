package domain

import (
	"errors"
	"math"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

var thermalTemperature = BoardMetric{Device: "zone-01", Kind: "thermal_sensor", Metric: "temperature"}

func TestAMetricsToleranceIsTheDeviceKeyThenTheKindKeyThenTheSpan(t *testing.T) {
	t.Parallel()
	span := MetricFacts{Unit: "Cel", Range: &EURange{Min: 0, Max: 50}}
	tests := []struct {
		name     string
		declared map[string]float64
		facts    MetricFacts
		source   string
		value    float64
		none     bool
	}{
		{"the device key wins over the kind key", map[string]float64{"zone-01.temperature": 1, "thermal_sensor.temperature": 4}, span,
			ToleranceDeclared, 1, false},
		{"the kind key applies when no device key is declared", map[string]float64{"thermal_sensor.temperature": 4}, span,
			ToleranceDeclared, 4, false},
		{"a declared value needs no eu_range", map[string]float64{"zone-01.temperature": 0.5}, MetricFacts{}, ToleranceDeclared, 0.5, false},
		{"the default is the fraction of the eu_range span", nil, span, ToleranceDefaultSpan, 2.5, false},
		{"a metric with neither a declaration nor a range has none", nil, MetricFacts{}, ToleranceNone, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ResolveTolerance(tt.declared, DefaultSpanFraction, thermalTemperature, tt.facts)
			if got.Source != tt.source || (got.Value == nil) != tt.none {
				t.Fatalf("tolerance = %+v, want source %s and none %v", got, tt.source, tt.none)
			}
			if !tt.none && *got.Value != tt.value {
				t.Fatalf("tolerance value = %v, want %v", *got.Value, tt.value)
			}
		})
	}
}

func TestTheDefaultToleranceScalesWithTheFraction(t *testing.T) {
	t.Parallel()
	facts := MetricFacts{Range: &EURange{Min: 20, Max: 90}}
	got := ResolveTolerance(nil, 0.1, thermalTemperature, facts)
	if got.Value == nil || math.Abs(*got.Value-7) > 1e-9 {
		t.Fatalf("tolerance = %+v, want 0.1 of the 70 span, 7", got)
	}
}

func TestParseTolerancesReadsCommaLists(t *testing.T) {
	t.Parallel()
	declared, err := ParseTolerances([]string{"zone-01.temperature=2, fan-01.level_percent=30", "thermal_sensor.humidity=5"})
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]float64{"zone-01.temperature": 2, "fan-01.level_percent": 30, "thermal_sensor.humidity": 5}
	if len(declared) != len(want) {
		t.Fatalf("declared = %v, want %v", declared, want)
	}
	for key, value := range want {
		if declared[key] != value {
			t.Errorf("%s = %v, want %v", key, declared[key], value)
		}
	}
}

func TestParseTolerancesAcceptsADottedDeviceId(t *testing.T) {
	t.Parallel()
	declared, err := ParseTolerances([]string{"zone.01.temperature=2"})
	if err != nil || declared["zone.01.temperature"] != 2 {
		t.Fatalf("ParseTolerances = %v, %v, want the dotted device id as the key", declared, err)
	}
}

func TestParseTolerancesRefusesEveryMalformedDeclaration(t *testing.T) {
	t.Parallel()
	for _, item := range []string{
		"zone-01.temperature",
		"temperature=2",
		"=2",
		"zone-01.temperature=",
		"zone-01.temperature=abc",
		"zone-01.temperature=-1",
		"zone-01.temperature=NaN",
		"zone-01.temperature=+Inf",
		"zone-01.temperature=1,zone-01.temperature=2",
		"zone-01.temperature=1,,zone-01.humidity=2",
	} {
		if declared, err := ParseTolerances([]string{item}); !errors.Is(err, ErrTwinTolerance) || declared != nil {
			t.Errorf("ParseTolerances(%q) = %v, %v, want ErrTwinTolerance", item, declared, err)
		}
	}
}

func TestCheckTolerancesRefusesAKeyThatNamesNoMetricOfTheBoard(t *testing.T) {
	t.Parallel()
	metrics := []BoardMetric{thermalTemperature}
	if err := CheckTolerances(map[string]float64{"zone-01.temperature": 1, "thermal_sensor.temperature": 1}, metrics); err != nil {
		t.Fatalf("CheckTolerances of known keys = %v", err)
	}
	err := CheckTolerances(map[string]float64{"zone-02.temperature": 1}, metrics)
	if !errors.Is(err, ErrTwinTolerance) || !strings.Contains(err.Error(), "zone-02.temperature") {
		t.Fatalf("CheckTolerances = %v, want ErrTwinTolerance naming zone-02.temperature", err)
	}
}

func TestSpanFractionOfAcceptsOnlyAFractionInTheUnitRange(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in   float64
		want float64
		fail bool
	}{
		{0, DefaultSpanFraction, false},
		{0.1, 0.1, false},
		{1, 1, false},
		{-0.1, 0, true},
		{1.5, 0, true},
		{math.NaN(), 0, true},
	}
	for _, tt := range tests {
		got, err := SpanFractionOf(tt.in)
		if tt.fail {
			if !errors.Is(err, ErrTwinTolerance) {
				t.Errorf("SpanFractionOf(%v) = %v, %v, want ErrTwinTolerance", tt.in, got, err)
			}
			continue
		}
		if err != nil || got != tt.want {
			t.Errorf("SpanFractionOf(%v) = %v, %v, want %v", tt.in, got, err, tt.want)
		}
	}
}

func TestTwinImplementsNeedsOneDomainKindForTheBoardsKinds(t *testing.T) {
	t.Parallel()
	spec := &model.DomainSpec{ID: "arduino-thermal", Kinds: []model.Kind{{Name: "thermal_sensor", Implements: "thermal_sensor"}}}
	if err := TwinImplements(spec, []BoardDevice{{Name: "fan-01", Kind: "fan"}, {Name: "zone-01", Kind: "thermal_sensor"}}); err != nil {
		t.Fatalf("TwinImplements with one shared kind = %v", err)
	}
	err := TwinImplements(spec, []BoardDevice{{Name: "fan-01", Kind: "fan"}, {Name: "led-01", Kind: "indicator"}, {Name: "fan-02", Kind: "fan"}})
	if !errors.Is(err, ErrTwinKinds) || !strings.Contains(err.Error(), "board kinds fan, indicator") {
		t.Fatalf("TwinImplements = %v, want ErrTwinKinds listing the board kinds once each", err)
	}
}

func TestMetricFactsReadTheUnitAndTheRangeOfAKindMetric(t *testing.T) {
	t.Parallel()
	kind := map[string]any{"metrics": []any{
		map[string]any{"name": "temperature", "unit": "Cel", "eu_range": map[string]any{"min": 0.0, "max": 50.0}},
		map[string]any{"name": "flat", "unit": "x", "eu_range": map[string]any{"min": 5.0, "max": 5.0}},
		map[string]any{"name": "inverted", "eu_range": map[string]any{"min": 9.0, "max": 1.0}},
		map[string]any{"name": "bare"},
	}}
	if got := MetricFactsOf(kind, "temperature"); got.Unit != "Cel" || got.Range == nil || got.Range.Max != 50 {
		t.Errorf("temperature facts = %+v, want Cel over 0 to 50", got)
	}
	for _, name := range []string{"flat", "inverted", "bare", "missing"} {
		if got := MetricFactsOf(kind, name); got.Range != nil {
			t.Errorf("%s facts = %+v, want no usable range", name, got)
		}
	}
}
