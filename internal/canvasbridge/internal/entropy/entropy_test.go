package entropy

import (
	"regexp"
	"testing"
)

func TestCodesAndSessionsAreFreshAndWellFormed(t *testing.T) {
	t.Parallel()
	a, errA := PairingCode()
	b, errB := PairingCode()
	if errA != nil || errB != nil || a == b || !regexp.MustCompile(`^[A-Z2-7]{12}$`).MatchString(a) {
		t.Fatalf("pairing codes %q and %q (%v, %v)", a, b, errA, errB)
	}
	session, err := SessionID()
	if err != nil || !regexp.MustCompile(`^[0-9a-f]{64}$`).MatchString(session) {
		t.Fatalf("session %q, %v", session, err)
	}
}
