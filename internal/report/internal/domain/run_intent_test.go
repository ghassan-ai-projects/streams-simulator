package domain

import (
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestARunReportsTheRequestedEndAndTerminationItsArtifactRecorded(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name        string
		plan        *model.RunPlan
		termination *model.Termination
		wantEnd     int64
		wantText    string
	}{
		{"none", nil, nil, 0, ""},
		{"plan only", &model.RunPlan{RequestedEndNS: 20000}, nil, 20000, ""},
		{"termination only", nil, &model.Termination{Reason: model.TerminationCompleted, AtNS: 9000}, 0, "Completed at "},
		{"stopped", &model.RunPlan{RequestedEndNS: 20000}, &model.Termination{Reason: model.TerminationStoppedByOperator, AtNS: 4000}, 20000, "Stopped by operator at "},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			report, err := Assemble(Folder{Artifact: endedArtifact(func(art *model.RunArtifact) { art.Plan, art.Termination = tc.plan, tc.termination })}, nil)
			if err != nil {
				t.Fatal(err)
			}
			run := report.Run
			if run.RequestedEndNS != tc.wantEnd || (run.Termination == nil) != (tc.termination == nil) {
				t.Fatalf("run = %+v, want requested end %d", run, tc.wantEnd)
			}
			if got := terminationText(run); !strings.HasPrefix(got, tc.wantText) {
				t.Fatalf("terminationText = %q, want prefix %q", got, tc.wantText)
			}
		})
	}
}
