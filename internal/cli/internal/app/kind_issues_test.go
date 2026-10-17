package app

import "testing"

func TestKindValidateReportsEachSchemaViolationAsAnIssueWithItsPointer(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, raw, pointer, message string
	}{
		{"a bad metric name", `{"kind": "k", "metrics": [{"name": "bad name", "data_type": "number"}], "commands": []}`,
			"/metrics/0/name", "string does not match pattern ^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$"},
		{"a missing commands member", `{"kind": "k", "metrics": []}`, "/commands", `missing required property "commands"`},
		{"an unknown data_type", `{"kind": "k", "metrics": [{"name": "m", "data_type": "float"}], "commands": []}`,
			"/metrics/0/data_type", "value not in enum"},
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
			if issue, _ := issues[0].(map[string]any); issue["pointer"] != tc.pointer || issue["message"] != tc.message || len(issue) != 2 {
				t.Fatalf("issue = %v, want pointer %s and message %q", issues[0], tc.pointer, tc.message)
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
	if body["code"] != "kind_failed" || body["source"] != path || len(issues) != 1 || issue["pointer"] != nil || issue["message"] == "" {
		t.Fatalf("failure = %v, want kind_failed with one issue that has a message and no pointer", body)
	}
}
