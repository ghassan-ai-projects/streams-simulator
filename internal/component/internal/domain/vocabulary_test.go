package domain

import (
	"math"
	"strings"
	"testing"
)

func vocabularyOnly(raw string) testFiles { return customerFiles(nil, nil, raw) }

func loadWithVocabulary(t *testing.T, raw string) (*Library, error) {
	t.Helper()
	library, err := LoadLibrary([]Directory{builtIn(t), vocabularyOnly(raw).directory("customer")})
	return library, err
}

func TestVocabularyMergesEntryByEntry(t *testing.T) {
	t.Parallel()
	library, err := loadWithVocabulary(t, `{"schema_version": "0.1",
	  "unit_families": {"temperature": {"base": "degC", "units": {"degR": {"scale": 0.5555555555555556, "offset": -273.15}}},
	    "pressure": {"base": "kPa", "units": {"kPa": {"scale": 1}, "bar": {"scale": 100}}}},
	  "port_types": {"steam": {"unit_family": "pressure"}},
	  "tag_letters": {"PT": {"measures": ["pressure"]}, "TT": {"measures": ["pressure"]}}}`)
	if err != nil {
		t.Fatal(err)
	}
	vocabulary := library.vocabulary
	if !vocabulary.hasUnit("temperature", "degR") || !vocabulary.hasUnit("temperature", "K") {
		t.Error("the added unit did not join the existing family")
	}
	if !vocabulary.hasUnit("pressure", "bar") || vocabulary.PortTypes["steam"].UnitFamily != "pressure" || vocabulary.PortTypes["thermal_load"].UnitFamily != "power" {
		t.Errorf("vocabulary = %+v, want the new family and port type beside the built-in ones", vocabulary)
	}
	if !vocabulary.tagMeasures("TT", "temperature") || !vocabulary.tagMeasures("TT", "pressure") || !vocabulary.tagMeasures("PT", "pressure") {
		t.Errorf("tags = %+v, want TT to measure both families and PT to exist", vocabulary.TagLetters)
	}
}

func TestVocabularyMergeRefusesEntriesDefinedDifferently(t *testing.T) {
	t.Parallel()
	cases := []struct{ name, vocabulary, message string }{
		{"a unit with another scale", `{"schema_version": "0.1", "unit_families": {"power": {"base": "kW", "units": {"kW": {"scale": 1}, "W": {"scale": 0.002}}}}}`,
			`unit "W" of family "power" is defined differently`},
		{"a unit with another offset", `{"schema_version": "0.1", "unit_families": {"temperature": {"base": "degC", "units": {"K": {"scale": 1, "offset": -273}}}}}`,
			`unit "K" of family "temperature" is defined differently`},
		{"a family with another base", `{"schema_version": "0.1", "unit_families": {"power": {"base": "W", "units": {"W": {"scale": 1}}}}}`,
			`unit family "power": base unit "W" differs from the existing "kW"`},
		{"a port type of another family", `{"schema_version": "0.1", "unit_families": {}, "port_types": {"thermal_load": {"unit_family": "temperature"}}}`,
			`port type "thermal_load" is defined differently`},
		{"a port type of an unknown family", `{"schema_version": "0.1", "unit_families": {}, "port_types": {"steam": {"unit_family": "mass"}}}`,
			`port type "steam" names unit family "mass", which no vocabulary defines`},
		{"a tag measuring an unknown family", `{"schema_version": "0.1", "unit_families": {}, "tag_letters": {"PT": {"measures": ["mass"]}}}`,
			`tag "PT" measures unit family "mass", which no vocabulary defines`},
		{"a base unit that is not the identity", `{"schema_version": "0.1", "unit_families": {"mass": {"base": "kg", "units": {"kg": {"scale": 2}}}}}`,
			`the base unit "kg" must be a unit with scale 1 and no offset`},
		{"a base unit that is not a unit", `{"schema_version": "0.1", "unit_families": {"mass": {"base": "kg", "units": {"g": {"scale": 1}}}}}`,
			`the base unit "kg" must be a unit with scale 1 and no offset`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := loadWithVocabulary(t, tc.vocabulary)
			invalid := asViolation(t, err, "component-vocabulary-v0.1")
			if invalid.Source != "customer/vocabulary.json" || !strings.Contains(err.Error(), tc.message) {
				t.Fatalf("error = %v (source %s), want %q in customer/vocabulary.json", err, invalid.Source, tc.message)
			}
		})
	}
}

func TestVocabularyRefusesAFileThatBreaksItsSchema(t *testing.T) {
	t.Parallel()
	_, err := loadWithVocabulary(t, `{"schema_version": "0.1"}`)
	wantViolation(t, err, "component-vocabulary-v0.1")
}

func TestVocabularyConvertsBetweenUnitsOfAFamily(t *testing.T) {
	t.Parallel()
	vocabulary := baseLibrary(t).vocabulary
	cases := []struct {
		family, from, to string
		gain, offset     float64
	}{
		{"temperature", "degC", "degC", 1, 0},
		{"temperature", "degC", "K", 1, 273.15},
		{"temperature", "K", "degC", 1, -273.15},
		{"temperature", "degC", "degF", 1.8, 32},
		{"power", "kW", "W", 1000, 0},
	}
	for _, tc := range cases {
		conv, err := vocabulary.convert(tc.family, tc.from, tc.to)
		if err != nil || math.Abs(conv.gain-tc.gain) > 1e-9 || math.Abs(conv.offset-tc.offset) > 1e-9 {
			t.Errorf("%s %s to %s = %+v (%v), want gain %v offset %v", tc.family, tc.from, tc.to, conv, err, tc.gain, tc.offset)
		}
	}
	if conv, _ := vocabulary.convert("temperature", "degC", "degC"); !conv.identity() {
		t.Error("equal units are not the identity")
	}
	if _, err := vocabulary.convert("temperature", "degC", "kW"); err == nil || !strings.Contains(err.Error(), `is not a unit of family "temperature"`) {
		t.Errorf("converting to a unit of another family gave %v, want the refusal", err)
	}
}
