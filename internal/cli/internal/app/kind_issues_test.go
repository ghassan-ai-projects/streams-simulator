package app

import (
	"reflect"
	"testing"
)

func TestKindValidateReportsEachSchemaViolationAsAnIssueWithItsPointerKeywordAndMember(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, raw string
		want      map[string]any
	}{
		{"a bad metric name", `{"kind": "k", "metrics": [{"name": "bad name", "data_type": "number"}], "commands": []}`,
			map[string]any{"pointer": "/metrics/0/name", "message": "string does not match pattern ^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$", "keyword": "pattern"}},
		{"a missing commands member", `{"kind": "k", "metrics": []}`,
			map[string]any{"pointer": "/commands", "message": `missing required property "commands"`, "keyword": "required", "member": "commands"}},
		{"an unknown data_type", `{"kind": "k", "metrics": [{"name": "m", "data_type": "float"}], "commands": []}`,
			map[string]any{"pointer": "/metrics/0/data_type", "message": "value not in enum", "keyword": "enum"}},
		{"an empty range", `{"kind": "k", "metrics": [{"name": "m", "data_type": "number", "eu_range": {"min": 1, "max": 1}}], "commands": []}`,
			map[string]any{"pointer": "/metrics/0/eu_range", "message": "min 1 is not below max 1", "keyword": "kind_range_empty", "member": "eu_range"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			path := writeFile(t, "bad.kind.json", tc.raw)
			got := invoke(t, "kind", "validate", path)
			body := got.failure(t)
			issues, _ := body["issues"].([]any)
			if got.code != 1 || body["code"] != "kind_failed" || body["source"] != path || len(issues) != 1 {
				t.Fatalf("exit %d with %v, want kind_failed from %s with one issue", got.code, body, path)
			}
			if !reflect.DeepEqual(issues[0], tc.want) {
				t.Fatalf("issue = %v, want %v", issues[0], tc.want)
			}
		})
	}
}

func TestKindValidateReportsAFileThatIsNotJSONAsAnIssueWithoutAPointer(t *testing.T) {
	t.Parallel()
	path := writeFile(t, "broken.kind.json", "{")
	body := invoke(t, "kind", "validate", path).failure(t)
	issues, _ := body["issues"].([]any)
	issue, _ := issues[0].(map[string]any)
	if body["code"] != "kind_failed" || body["source"] != path || len(issues) != 1 || issue["pointer"] != nil || issue["message"] == "" || issue["keyword"] != "json" {
		t.Fatalf("failure = %v, want kind_failed with one issue that has a message and the keyword json and no pointer", body)
	}
}

func TestDomainValidateReportsTheKeywordOfEachSchemaViolationAndOfEachRule(t *testing.T) {
	t.Parallel()
	bad := writeFile(t, "bad.domain.json", `{"schema_version": "0.2"}`)
	issues, _ := invoke(t, "domain", "validate", bad).failure(t)["issues"].([]any)
	if len(issues) == 0 {
		t.Fatal("a domain without its members must be refused with issues")
	}
	for _, raw := range issues {
		issue, _ := raw.(map[string]any)
		if issue["keyword"] == nil || issue["pointer"] == nil {
			t.Errorf("issue %v, want a pointer and a keyword", issue)
		}
	}
}
