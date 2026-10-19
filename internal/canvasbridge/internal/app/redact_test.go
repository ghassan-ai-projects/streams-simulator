package app

import (
	"errors"
	"testing"
)

func TestRedactHidesTheWorkspaceRootOfAPlainErrorAndKeepsItsCause(t *testing.T) {
	t.Parallel()
	cause := errors.New("open /work/space/x: denied, then /work/space")
	err := redact(cause, "/work/space")
	if err.Error() != "open x: denied, then ." || !errors.Is(err, cause) {
		t.Fatalf("redact = %q, want the root hidden and the cause kept", err)
	}
}

func TestRedactLeavesAnErrorWithoutTheRootAsItIs(t *testing.T) {
	t.Parallel()
	cause := errors.New("open /elsewhere/x: denied")
	if err := redact(cause, "/work/space"); !errors.Is(err, cause) || err.Error() != cause.Error() {
		t.Fatalf("redact = %v, want the error unchanged", err)
	}
}
