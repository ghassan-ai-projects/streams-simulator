package domain

import (
	"encoding/json"
	"errors"
	"math"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestANumberIsReadOnlyFromAJSONNumberOrADouble(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		value any
		want  float64
		known bool
	}{
		{"a JSON number", json.Number("-2.5"), -2.5, true},
		{"a double", float64(7), 7, true},
		{"a JSON number beyond a double", json.Number("1e999"), 0, false},
		{"a negative JSON number beyond a double", json.Number("-1e999"), 0, false},
		{"an infinite double", math.Inf(1), 0, false},
		{"a double that is no number", math.NaN(), 0, false},
		{"text", "7", 0, false},
		{"nothing", nil, 0, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got, known := numberOf(tc.value); got != tc.want || known != tc.known {
				t.Fatalf("numberOf(%v) = %v, %v, want %v, %v", tc.value, got, known, tc.want, tc.known)
			}
		})
	}
}

func TestARangeIsEmptyOnlyWhenBothBoundsAreKnownAndMinIsNotBelowMax(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		value  any
		empty  bool
		detail string
	}{
		{"min below max", map[string]any{"min": float64(0), "max": float64(1)}, false, ""},
		{"min equal to max", map[string]any{"min": float64(2), "max": float64(2)}, true, "min 2 is not below max 2"},
		{"min above max, as written", map[string]any{"min": json.Number("5"), "max": json.Number("-1")}, true, "min 5 is not below max -1"},
		{"no max", map[string]any{"min": float64(5)}, false, ""},
		{"a bound beyond a double", map[string]any{"min": json.Number("1e999"), "max": float64(1)}, false, ""},
		{"no range", nil, false, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			issue, empty := emptyRange(tc.value, "/metrics/0/eu_range")
			if empty != tc.empty || (empty && (issue.Pointer != "/metrics/0/eu_range" || issue.Message != tc.detail)) {
				t.Fatalf("emptyRange = %+v, %v, want empty %v with %q", issue, empty, tc.empty, tc.detail)
			}
		})
	}
}

func TestCheckKindRefusesAnEmptyRangeOfAHandBuiltEntryAndNamesEachMetric(t *testing.T) {
	t.Parallel()
	entry := map[string]any{"kind": "k", "commands": []any{}, "metrics": []any{
		map[string]any{"name": "a", "data_type": "number", "eu_range": map[string]any{"min": float64(0), "max": float64(0)}},
		map[string]any{"name": "b", "data_type": "number", "instrument_range": map[string]any{"min": float64(3), "max": float64(1)}},
	}}
	invalid, ok := model.AsValidationError(CheckKind(entry))
	want := []model.Issue{model.LocatedIssue("/metrics/0/eu_range", "min 0 is not below max 0"), model.LocatedIssue("/metrics/1/instrument_range", "min 3 is not below max 1")}
	if !ok || !errors.Is(invalid, ErrEmptyRange) || invalid.Contract != model.ContractKindFile || len(invalid.Issues) != 2 || invalid.Issues[0] != want[0] || invalid.Issues[1] != want[1] {
		t.Fatalf("CheckKind = %+v, want a kind-file error caused by ErrEmptyRange with issues %+v", invalid, want)
	}
}

func TestCheckKindRefusesAKindDigestBeforeAnythingElseAndCarriesTheCause(t *testing.T) {
	t.Parallel()
	err := CheckKind(map[string]any{"kind_digest": "sha256:00"})
	invalid, ok := model.AsValidationError(err)
	if !ok || !errors.Is(err, ErrKindDigestInFile) || len(invalid.Issues) != 1 || invalid.Issues[0].Pointer != "/kind_digest" ||
		!strings.HasPrefix(invalid.Summary, "plantwire: kind file fails its schema: /kind_digest: ") {
		t.Fatalf("CheckKind = %v, want ErrKindDigestInFile at /kind_digest and no other issue, though the entry also lacks every required member", err)
	}
}

func TestAMemberNamedTwiceIsFoundInAnyObjectOfAnyList(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		raw  string
		dup  bool
	}{
		{"an object inside a list inside a list", `{"kind":"a","metrics":[],"commands":[{"name":"go","args_schema":{"anyOf":[[{"type":"object","type":"array"}]]}}]}`, true},
		{"a repeated name in different objects", `{"kind":"a","metrics":[{"name":"m","data_type":"number"},{"name":"n","data_type":"number"}],"commands":[]}`, false},
		{"scalars and nulls in lists", `{"kind":"a","metrics":[],"commands":[{"name":"go","args_schema":{"enum":[1,"x",null,true,[2]]}}]}`, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			entry, err := decodeKind([]byte(tc.raw))
			if tc.dup {
				if entry != nil || !errors.Is(err, ErrDuplicateMember) || !strings.Contains(err.Error(), `"type"`) {
					t.Fatalf("decodeKind = %v, %v, want ErrDuplicateMember naming type", entry, err)
				}
				return
			}
			if err != nil || entry["kind"] != "a" {
				t.Fatalf("decodeKind = %v, %v, want the document decoded", entry, err)
			}
		})
	}
}

func TestTheMemberWalkReportsADocumentThatEndsBeforeItIsClosed(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{``, `{"a":`, `{"a":1`, `[1`, `{"a":[1,2}`} {
		err := refuseDuplicateMembers([]byte(raw))
		if err == nil || !strings.HasPrefix(err.Error(), "decode: ") || errors.Is(err, ErrDuplicateMember) {
			t.Errorf("refuseDuplicateMembers(%q) = %v, want a decode error that is not a duplicate member", raw, err)
		}
	}
}
