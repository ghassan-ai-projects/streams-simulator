package domain

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestAKindFileThatBreaksTheSchemaReportsEachViolationWithItsPointer(t *testing.T) {
	t.Parallel()
	const namePattern = "^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$"
	cases := []struct {
		name string
		raw  string
		want []model.Issue
	}{
		{"a bad metric name", `{"kind": "k", "metrics": [{"name": "bad name", "data_type": "number"}], "commands": []}`,
			[]model.Issue{model.RuleIssue("/metrics/0/name", "pattern", "", "string does not match pattern "+namePattern)}},
		{"a missing commands member", `{"kind": "k", "metrics": []}`,
			[]model.Issue{model.RuleIssue("/commands", "required", "commands", `missing required property "commands"`)}},
		{"an unknown data_type", `{"kind": "k", "metrics": [{"name": "m", "data_type": "float"}], "commands": []}`,
			[]model.Issue{model.RuleIssue("/metrics/0/data_type", "enum", "", "value not in enum")}},
		{"a kind_digest that is stale", `{"kind": "k", "kind_digest": "sha256:00", "metrics": [], "commands": []}`, []model.Issue{kindDigestIssue}},
		{"a kind_digest that is right", `{"kind": "k", "kind_digest": "` + emptyKindDigest + `", "metrics": [], "commands": []}`, []model.Issue{kindDigestIssue}},
		{"an eu_range with min equal to max", `{"kind": "k", "metrics": [{"name": "m", "data_type": "number", "eu_range": {"min": 5, "max": 5}}], "commands": []}`,
			[]model.Issue{model.RuleIssue("/metrics/0/eu_range", "kind_range_empty", "eu_range", "min 5 is not below max 5")}},
		{"ranges that run backwards, sorted by pointer", `{"kind": "k", "metrics": [{"name": "a", "data_type": "number"},
			{"name": "b", "data_type": "number", "instrument_range": {"min": 10, "max": 2.5}, "eu_range": {"min": 1, "max": 0}}], "commands": []}`,
			[]model.Issue{model.RuleIssue("/metrics/1/eu_range", "kind_range_empty", "eu_range", "min 1 is not below max 0"),
				model.RuleIssue("/metrics/1/instrument_range", "kind_range_empty", "instrument_range", "min 10 is not below max 2.5")}},
		{"every violation, sorted by pointer", `{"kind": "k"}`,
			[]model.Issue{model.RuleIssue("/commands", "required", "commands", `missing required property "commands"`),
				model.RuleIssue("/metrics", "required", "metrics", `missing required property "metrics"`)}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, _, err := ParseKindFiles([]KindSource{{Path: "k.kind.json", Raw: []byte(tc.raw)}})
			invalid, ok := model.AsValidationError(err)
			if !ok || invalid.Contract != model.ContractKindFile || invalid.Source != "k.kind.json" || !reflect.DeepEqual(invalid.Issues, tc.want) {
				t.Fatalf("ParseKindFiles = %v, want a kind-file violation from k.kind.json with issues %+v", err, tc.want)
			}
		})
	}
}

func TestAKindFileThatIsNotJSONIsOneUnlocatedIssue(t *testing.T) {
	t.Parallel()
	_, _, err := ParseKindFiles([]KindSource{{Path: "bad.kind.json", Raw: []byte("{")}})
	invalid, ok := model.AsValidationError(err)
	if !ok || invalid.Contract != model.ContractKindFile || invalid.Source != "bad.kind.json" || len(invalid.Issues) != 1 ||
		invalid.Issues[0].Located || invalid.Issues[0].Message == "" || invalid.Issues[0].Keyword != "json" || errors.Unwrap(invalid) == nil {
		t.Fatalf("ParseKindFiles = %v (%+v), want one unlocated issue naming bad.kind.json", err, invalid)
	}
}

const emptyKindDigest = "sha256:540c36216dd6364d21e22923590e1cfc5791f7a06c51774cd509214b0451a177"

var kindDigestIssue = model.RuleIssue("/kind_digest", "kind_digest_forbidden", "kind_digest", "a kind file must not carry kind_digest: it is computed from the file, so remove the member")

func TestAKindFileWithAKindDigestOrAnEmptyRangeNamesWhyItIsRefused(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		raw   string
		cause error
		text  string
	}{
		{"a kind_digest", `{"kind": "k", "kind_digest": "` + emptyKindDigest + `", "metrics": [], "commands": []}`, ErrKindDigestInFile,
			"plantwire: kind file fails its schema: /kind_digest: a kind file must not carry kind_digest"},
		{"an empty range", `{"kind": "k", "metrics": [{"name": "m", "data_type": "number", "eu_range": {"min": 3, "max": 3}}], "commands": []}`, ErrEmptyRange,
			"plantwire: kind file fails its schema: /metrics/0/eu_range: min 3 is not below max 3"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, _, err := ParseKindFiles([]KindSource{{Path: "k.kind.json", Raw: []byte(tc.raw)}})
			if !errors.Is(err, tc.cause) || !strings.Contains(err.Error(), tc.text) {
				t.Fatalf("ParseKindFiles = %v, want %v with the text %q", err, tc.cause, tc.text)
			}
		})
	}
}

func TestAKindFileWithRangesThatRunUpwardsIsAccepted(t *testing.T) {
	t.Parallel()
	raw := `{"kind": "k", "metrics": [{"name": "m", "data_type": "number", "eu_range": {"min": -1, "max": 1}, "instrument_range": {"min": -2.5, "max": 2.5}}], "commands": []}`
	files, _, err := ParseKindFiles([]KindSource{{Path: "k.kind.json", Raw: []byte(raw)}})
	if err != nil || len(files) != 1 {
		t.Fatalf("ParseKindFiles = %v, %v, want the one kind file accepted", files, err)
	}
}
