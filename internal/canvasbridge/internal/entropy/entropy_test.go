package entropy

import (
	"regexp"
	"testing"
)

func TestCodesAndSessionsAreFreshAndWellFormed(t *testing.T) {
	t.Parallel()
	a, b := PairingCode(), PairingCode()
	if a == b || !regexp.MustCompile(`^[A-Z2-7]{12}$`).MatchString(a) {
		t.Fatalf("pairing codes %q and %q", a, b)
	}
	session := SessionID()
	if session == SessionID() || !regexp.MustCompile(`^[0-9a-f]{64}$`).MatchString(session) {
		t.Fatalf("session %q", session)
	}
	if run := RunID(); run == RunID() || !regexp.MustCompile(`^c-[0-9a-f]{8}$`).MatchString(run) {
		t.Fatalf("run %q", run)
	}
}
