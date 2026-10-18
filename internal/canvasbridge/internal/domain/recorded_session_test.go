package domain

import "testing"

func TestRecordedSessionAcceptsOnlyExplicitConnectedMetadata(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{
		`{"type":"session","mode":"evaluate","plan":{}}`,
		`{"type":"session","mode":"connected"}`,
		`{"type":"wire","mode":"connected","plan":{}}`,
		`{"type":"session","mode":"connected","plan":`,
		`{"type":"session","stream":"observer","mode":"connected","plan":null}`,
		`{"type":"session","stream":"director","mode":"connected","plan":{}}`,
		`{"type":"wire","stream":"observer"}` + "\n" + `{"type":"session","stream":"observer","mode":"connected","plan":{}}`,
	} {
		if plan, ok := RecordedSession([]byte(raw)); ok {
			t.Fatalf("unproven session classified connected: %s => %s", raw, plan)
		}
	}
}

func TestRecordedSessionIsTheFirstSessionRecordAfterTheOpening(t *testing.T) {
	t.Parallel()
	head := `{"type":"run.opened","stream":"observer"}` + "\n" + `{"type":"observation","stream":"observer"}` + "\n" +
		`{"type":"session","stream":"observer","mode":"connected","plan":{"instances":[]}}` + "\n"
	if plan, ok := RecordedSession([]byte(head)); !ok || string(plan) != `{"instances":[]}` {
		t.Fatalf("recorded session = %s, %v; want the plan of the first session record", plan, ok)
	}
}
