package app

import (
	"errors"
	"strings"
	"testing"
)

func TestADigestMismatchKeepsItsErrorAndNamesItsVerdict(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"domain":  "the recorded domain spec does not match its digest",
		"adapter": "the recorded adapter spec does not match its digest",
		"world":   "the recorded specs replay to a different world than the recorded world digest",
	}
	for subject, verdict := range cases {
		t.Run(subject, func(t *testing.T) {
			t.Parallel()
			err := &digestMismatch{subject, "sha256:a", "sha256:b"}
			if !errors.Is(err, ErrDigestMismatch) || !strings.HasPrefix(err.Error(), "run: "+subject+" digest mismatch: ") {
				t.Fatalf("error = %v", err)
			}
			if !strings.HasPrefix(err.verdict(), verdict) || !strings.Contains(err.verdict(), "sha256:a") || !strings.Contains(err.verdict(), "sha256:b") {
				t.Fatalf("verdict = %q", err.verdict())
			}
		})
	}
}
