package domain

import (
	"math"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestCheckRunPlan(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		plan model.RunPlan
		ok   bool
	}{
		{"as fast as possible", model.RunPlan{RequestedEndNS: 1}, true},
		{"paced", model.RunPlan{RequestedEndNS: 60e9, Rate: 120}, true},
		{"no end", model.RunPlan{}, false},
		{"negative end", model.RunPlan{RequestedEndNS: -1}, false},
		{"negative rate", model.RunPlan{RequestedEndNS: 1, Rate: -0.5}, false},
		{"NaN rate", model.RunPlan{RequestedEndNS: 1, Rate: math.NaN()}, false},
		{"infinite rate", model.RunPlan{RequestedEndNS: 1, Rate: math.Inf(1)}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if err := CheckRunPlan(tc.plan); (err == nil) != tc.ok {
				t.Fatalf("CheckRunPlan(%+v) = %v, want ok=%v", tc.plan, err, tc.ok)
			}
		})
	}
}

func TestCheckTermination(t *testing.T) {
	t.Parallel()
	for reason, ok := range map[string]bool{model.TerminationCompleted: true, model.TerminationStoppedByOperator: true, "": false, "crashed": false} {
		if err := CheckTermination(model.Termination{Reason: reason}); (err == nil) != ok {
			t.Errorf("CheckTermination(%q) = %v, want ok=%v", reason, err, ok)
		}
	}
}
