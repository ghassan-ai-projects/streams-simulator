package app

import (
	"errors"
	"testing"
)

func TestAnIssueNamesTheFlagBehindIt(t *testing.T) {
	t.Parallel()
	origins := []string{`--fault "a=f@1"`, "--duration 60"}
	cases := map[string]string{
		"/commands/0/args/fault": `--fault "a=f@1"`,
		"/commands/1/args/to_ns": "--duration 60",
		"/commands/1":            "--duration 60",
		"/commands/x":            "the run flags",
		"/commands/-1":           "the run flags",
		"/commands/7/args/to_ns": "the run flags",
		"/world/profile":         "--profile",
		"/world":                 "the run flags",
		"":                       "the run flags",
	}
	for pointer, want := range cases {
		if got := flagOrigin(pointer, origins); got != want {
			t.Errorf("flagOrigin(%q) = %q, want %q", pointer, got, want)
		}
	}
}

// A failure that is not a scenario violation is passed on, not restated as a
// usage failure.
func TestAFailureThatIsNotAViolationIsPassedOn(t *testing.T) {
	t.Parallel()
	cause := errors.New("disk full")
	err := flagScenarioFailure(cause, nil)
	if _, usage := asUsageError(err); usage || !errors.Is(err, cause) {
		t.Fatalf("flagScenarioFailure = %v, want the cause passed on", err)
	}
}
