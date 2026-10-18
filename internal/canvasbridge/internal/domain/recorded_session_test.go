package domain

import "testing"

func TestRecordedSessionAcceptsOnlyExplicitConnectedMetadata(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{
		`{"type":"session","mode":"evaluate","plan":{}}`,
		`{"type":"session","mode":"connected"}`,
		`{"type":"wire","mode":"connected","plan":{}}`,
		`{"type":"session","mode":"connected","plan":`,
	} {
		if plan, ok := RecordedSession([]byte(raw)); ok {
			t.Fatalf("unproven session classified connected: %s => %s", raw, plan)
		}
	}
}
