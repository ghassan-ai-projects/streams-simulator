package domain

import (
	"bytes"
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonschema"
	"github.com/ghassan-ai-projects/streams-simulator/internal/schemas"
)

func journalLineViolations(t *testing.T, record map[string]any) []jsonschema.Error {
	t.Helper()
	var doc any
	decoder := json.NewDecoder(bytes.NewReader(schemas.RunJournal()))
	decoder.UseNumber()
	if err := decoder.Decode(&doc); err != nil {
		t.Fatal(err)
	}
	schema, err := jsonschema.Compile(doc)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		t.Fatal(err)
	}
	return schema.Validate(value)
}

func sessionRecord(plan map[string]any) map[string]any {
	return map[string]any{"stream": "observer", "seq": 0, "type": "session", "at_ns": 0, "mode": "connected", "plan": plan}
}

func liveRecord(command map[string]any) map[string]any {
	return map[string]any{"stream": "observer", "seq": 0, "type": "live", "at_ns": 0, "host": "online", "gateway": "not_applicable",
		"devices": []any{}, "commands": []any{command}, "marks": []any{}, "commands_truncated": false}
}

func refusedCommand(digest any) map[string]any {
	return map[string]any{"command_id": "c-1", "device": "P1", "environment": "simulated", "command": "start", "boot_id": "boot-1", "sent": true,
		"receipt": "refused", "reason": "policy_not_allowed", "report": "pending", "effect": "unknown", "effect_evidence": "",
		"sent_at_ns": 1, "updated_at_ns": 1, "policy_digest": digest}
}

func sessionPlan(policies any) map[string]any {
	return map[string]any{"domain": "d", "environment": "simulated", "instances": []any{}, "implements_checked": []any{}, "policies": policies}
}

func TestTheSessionRecordCarriesTheSortedPolicyListAndRefusesAnythingElse(t *testing.T) {
	t.Parallel()
	good := policyFixture
	nine := make([]any, 9)
	for i := range nine {
		nine[i] = good
	}
	cases := []struct {
		name   string
		record map[string]any
		want   string
	}{
		{name: "an empty list", record: sessionRecord(sessionPlan([]any{}))},
		{name: "one digest", record: sessionRecord(sessionPlan([]any{good}))},
		{name: "nine digests", record: sessionRecord(sessionPlan(nine)), want: "maxItems"},
		{name: "a digest in upper case", record: sessionRecord(sessionPlan([]any{strings.ToUpper(good)})), want: "pattern"},
		{name: "a list that is absent", record: sessionRecord(map[string]any{"domain": "d", "environment": "simulated", "instances": []any{},
			"implements_checked": []any{}}), want: "required"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			errs := journalLineViolations(t, tc.record)
			if tc.want == "" && len(errs) > 0 {
				t.Fatalf("violations = %v, want none", errs)
			}
			if tc.want != "" && !slices.ContainsFunc(errs, func(e jsonschema.Error) bool { return e.Keyword == tc.want }) {
				t.Fatalf("violations = %v, want one keyed %q", errs, tc.want)
			}
		})
	}
}

func TestALiveCommandCarriesOnlyAWellFormedPolicyDigest(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		digest any
		bad    bool
	}{
		{name: "a policy digest", digest: policyFixture},
		{name: "free text", digest: "the host's policy", bad: true},
		{name: "a short digest", digest: "sha256:bb", bad: true},
		{name: "a number", digest: 7, bad: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			errs := journalLineViolations(t, liveRecord(refusedCommand(tc.digest)))
			if (len(errs) > 0) != tc.bad {
				t.Fatalf("violations = %v, want a violation: %v", errs, tc.bad)
			}
		})
	}
}
