package domain

import (
	"errors"
	"reflect"
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
			[]model.Issue{model.LocatedIssue("/metrics/0/name", "string does not match pattern "+namePattern)}},
		{"a missing commands member", `{"kind": "k", "metrics": []}`,
			[]model.Issue{model.LocatedIssue("/commands", `missing required property "commands"`)}},
		{"an unknown data_type", `{"kind": "k", "metrics": [{"name": "m", "data_type": "float"}], "commands": []}`,
			[]model.Issue{model.LocatedIssue("/metrics/0/data_type", "value not in enum")}},
		{"every violation, sorted by pointer", `{"kind": "k"}`,
			[]model.Issue{model.LocatedIssue("/commands", `missing required property "commands"`),
				model.LocatedIssue("/metrics", `missing required property "metrics"`)}},
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
		invalid.Issues[0].Located || invalid.Issues[0].Message == "" || errors.Unwrap(invalid) == nil {
		t.Fatalf("ParseKindFiles = %v (%+v), want one unlocated issue naming bad.kind.json", err, invalid)
	}
}
