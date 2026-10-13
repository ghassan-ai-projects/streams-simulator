package domain

import (
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestAFeedbackLoopNeedsASampledReadOnTheSourceStepGrid(t *testing.T) {
	t.Parallel()
	link := func(d map[string]any) map[string]any {
		return at(d, "kinds", 0, "dynamics", 1, "f1", "inputs", 1, "link").(map[string]any)
	}
	cases := []struct {
		name    string
		change  func(d map[string]any)
		pointer string
	}{
		{"a loop closed by a sampled read", func(map[string]any) {}, ""},
		{"a loop with no sampled read", func(d map[string]any) { delete(link(d), "sample_s") }, "/kinds/0"},
		{"a sample off the source's 2 s step", func(d map[string]any) { link(d)["sample_s"] = 5.0 }, "/kinds/0/dynamics/1/f1/inputs/1/link/sample_s"},
		{"a sample finer than a millisecond", func(d map[string]any) { link(d)["sample_s"] = 0.0005 }, "/kinds/0/dynamics/1/f1/inputs/1/link/sample_s"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			violationIn(t, testsupport.LoopDomain(), tc.change, tc.pointer)
		})
	}
}
