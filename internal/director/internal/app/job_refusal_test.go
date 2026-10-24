package app

import (
	"errors"
	"strings"
	"testing"

	rules "github.com/ghassan-ai-projects/streams-simulator/internal/director/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestARequestThatCannotRunIsRefusedBeforeAnyWorldExists(t *testing.T) {
	t.Parallel()
	for name, tc := range map[string]struct {
		change func(*rules.JobRequest)
		want   string
	}{
		"bad duration":      {func(r *rules.JobRequest) { r.DurationS = 0 }, "duration_s"},
		"unknown fault":     {func(r *rules.JobRequest) { r.Faults[0].Fault = "no-such-fault" }, "no-such-fault"},
		"unknown entity":    {func(r *rules.JobRequest) { r.Faults[0].EntityID = "site-z/none" }, "site-z/none"},
		"unknown effector":  {func(r *rules.JobRequest) { r.Calls[0].Effector = "no-such-effector" }, "no-such-effector"},
		"bad effector args": {func(r *rules.JobRequest) { r.Calls[0].Args = map[string]any{"level": 9} }, "start_aerator"},
		"unknown domain":    {func(r *rules.JobRequest) { r.Domain = "no-such-domain" }, "no-such-domain"},
		"beyond 30 days":    {func(r *rules.JobRequest) { r.DurationS = 31 * 86400; r.Calls[0].AtS = 0; r.Faults[0].AtS = 0 }, "30 days"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			d := pacedBy(shippedDirector(t, t.Context()), freePacer{})
			req := pondJob()
			tc.change(&req)
			_, err := d.StartJob(req)
			var tool *rules.ToolError
			if !errors.As(err, &tool) || !strings.Contains(tool.Msg, tc.want) {
				t.Fatalf("err = %v, want a refusal naming %q", err, tc.want)
			}
			if len(d.Worlds) != 0 || len(d.ListJobs()) != 0 {
				t.Fatalf("%d worlds, %d jobs, want none after a refused start", len(d.Worlds), len(d.ListJobs()))
			}
		})
	}
}

func lineJob() rules.JobRequest {
	return rules.JobRequest{Domain: "discrete-line-oee", Seed: 2, DurationS: 600,
		Faults: []rules.JobFault{{EntityID: "line-1", Fault: "guard_trigger", AtS: 0}},
		Calls: []rules.JobCall{
			{EntityID: "line-1", Effector: "reset_line", Args: map[string]any{"line_id": "line-1"}, AtS: 60},
			{EntityID: "line-1", Effector: "adjust_speed", Args: map[string]any{"line_id": "line-1"}, AtS: 120},
		}}
}

func TestARefusedEffectorCallIsAnOutcomeAndTheJobGoesOn(t *testing.T) {
	t.Parallel()
	d := pacedBy(shippedDirector(t, t.Context()), freePacer{})
	started := startJob(t, d, lineJob())
	status := awaitJob(t, d, started.WorldID)
	if status.State != rules.JobCompleted || len(status.Calls) != 2 {
		t.Fatalf("status = %+v, want a completed job with both calls recorded", status)
	}
	refused, accepted := status.Calls[0], status.Calls[1]
	if refused.Effector != "reset_line" || refused.Accepted || refused.Reason != "interlock_refused" || refused.AtS != 60 {
		t.Fatalf("refused call = %+v, want the interlock refusal as an outcome", refused)
	}
	if accepted.Effector != "adjust_speed" || !accepted.Accepted || accepted.AtS != 120 {
		t.Fatalf("second call = %+v, want it accepted after the refusal", accepted)
	}
	art := loadRunArtifact(t, d, started.WorldID)
	if art.Incomplete || len(art.CommandLog) == 0 || art.Termination == nil || art.Termination.Reason != model.TerminationCompleted {
		t.Fatalf("run = incomplete %v, %d commands, termination %+v, want a complete run", art.Incomplete, len(art.CommandLog), art.Termination)
	}
	requireVerifies(t, d, started.WorldID)
}
