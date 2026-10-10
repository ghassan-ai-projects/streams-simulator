package model

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestDecodeRejectsTrailingJSON(t *testing.T) {
	t.Parallel()
	var got any
	if err := Decode(strings.NewReader(`{"ok":true} {"trailing":true}`), &got); err == nil {
		t.Fatal("decoder accepted two concatenated JSON documents")
	}
	if err := Decode(strings.NewReader(`{"ok":true} trailing`), &got); err == nil {
		t.Fatal("decoder accepted trailing non-JSON data")
	}
}

func TestDecodePreservesLargeNumbers(t *testing.T) {
	t.Parallel()
	var got map[string]any
	if err := Decode(strings.NewReader(`{"seed":18446744073709551615}`), &got); err != nil {
		t.Fatal(err)
	}
	if got["seed"].(json.Number).String() != "18446744073709551615" {
		t.Fatalf("large number lost precision: %#v", got["seed"])
	}
}

func TestTimeFormatAndParseRoundTripNanoseconds(t *testing.T) {
	t.Parallel()
	const ns = DefaultStartTimeNS + 123456789
	text := FormatTime(ns)
	if text != "2026-01-01T00:00:00.123456789Z" {
		t.Fatalf("FormatTime = %q", text)
	}
	got, err := ParseTime(text)
	if err != nil || got != ns {
		t.Fatalf("ParseTime(%q) = %d, %v; want %d", text, got, err, ns)
	}
}

func TestParseTimeNamesTheTimestampItRefused(t *testing.T) {
	t.Parallel()
	_, err := ParseTime("yesterday")
	if err == nil || !strings.HasPrefix(err.Error(), `model: invalid date-time "yesterday": `) {
		t.Fatalf("err = %v", err)
	}
}

func TestDecodeBytesAppliesTheSameStrictnessAsDecode(t *testing.T) {
	t.Parallel()
	var got map[string]any
	if err := DecodeBytes([]byte(`{"a":1}`), &got); err != nil || got["a"].(json.Number).String() != "1" {
		t.Fatalf("got %v, %v", got, err)
	}
	err := DecodeBytes([]byte(`{"a":1} {"b":2}`), &got)
	if err == nil || err.Error() != "model: trailing JSON document" {
		t.Fatalf("err = %v", err)
	}
}

func TestValidateVerdictAndArtifactRefuseNonConformingDocuments(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		validate func([]byte) error
		raw      string
		want     string
	}{
		{"verdict not JSON", ValidateVerdict, `not json`, "model: verdict not valid JSON"},
		{"verdict wrong shape", ValidateVerdict, `{"unexpected":true}`, "model: verdict fails consumer-verdict-v0.1"},
		{"artifact not JSON", ValidateRunArtifact, `{`, "model: artifact not valid JSON"},
		{"artifact wrong shape", ValidateRunArtifact, `[]`, "model: artifact fails run-artifact-v0.1"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if err := tc.validate([]byte(tc.raw)); err == nil || !strings.HasPrefix(err.Error(), tc.want) {
				t.Fatalf("err = %v, want prefix %q", err, tc.want)
			}
		})
	}
}

func TestF1InputAcceptsTheBareStateAndTheObjectForm(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		raw  string
		want F1Input
	}{
		{"bare string", `"temperature"`, F1Input{State: "temperature", Coef: 1, CoefSet: true}},
		{"object without coef", `{"state":"temperature"}`, F1Input{State: "temperature", Coef: 1}},
		{"object with coef", `{"state":"temperature","coef":-2.5}`, F1Input{State: "temperature", Coef: -2.5, CoefSet: true}},
		{"explicit zero coef", `{"state":"temperature","coef":0}`, F1Input{State: "temperature", Coef: 0, CoefSet: true}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var got F1Input
			if err := json.Unmarshal([]byte(tc.raw), &got); err != nil || got != tc.want {
				t.Fatalf("got %+v, %v; want %+v", got, err, tc.want)
			}
		})
	}
	var bad F1Input
	if err := json.Unmarshal([]byte(`[1]`), &bad); err == nil || !strings.HasPrefix(err.Error(), "UnmarshalJSON: ") {
		t.Fatalf("an array is neither form: %v", err)
	}
}

func TestGroundTruthIsPositiveOnlyForAnExpectedNonNegativeEpisode(t *testing.T) {
	t.Parallel()
	cases := []struct {
		record GroundTruthRecord
		want   bool
	}{
		{GroundTruthRecord{ExpectedEpisode: true}, true},
		{GroundTruthRecord{ExpectedEpisode: true, IsNegativeClass: true}, false},
		{GroundTruthRecord{}, false},
	}
	for _, tc := range cases {
		if got := tc.record.IsPositive(); got != tc.want {
			t.Errorf("%+v: IsPositive = %v, want %v", tc.record, got, tc.want)
		}
	}
}

func TestCurrentPlatformNamesTheBuild(t *testing.T) {
	t.Parallel()
	p := CurrentPlatform()
	if p.GOOS == "" || p.GOARCH == "" || !strings.HasPrefix(p.GoVersion, "go") {
		t.Fatalf("platform = %+v", p)
	}
}
