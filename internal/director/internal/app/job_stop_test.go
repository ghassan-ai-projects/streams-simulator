package app

import (
	"context"
	"errors"
	"strings"
	"testing"

	rules "github.com/ghassan-ai-projects/streams-simulator/internal/director/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestStoppingAJobMidRunEndsItEarlyWithTheExecutedPrefix(t *testing.T) {
	t.Parallel()
	pacer := newGatedPacer()
	d := pacedBy(shippedDirector(t, t.Context()), pacer)
	req := pondJob()
	started := startJob(t, d, req)
	pacer.release(3)
	for range 4 {
		<-pacer.asked
	}
	stopping, err := d.StopJob(started.WorldID)
	if err != nil || stopping.State != rules.JobStopping {
		t.Fatalf("StopJob = %+v, %v, want the job stopping", stopping, err)
	}
	status := awaitJob(t, d, started.WorldID)
	scenario := rules.PlanScenario(req, "native-jsonl", started.StartNS)
	art := loadRunArtifact(t, d, started.WorldID)
	reached := started.StartNS + 3*120e9
	if status.State != rules.JobStopped || status.ReachedNS != reached || len(art.CommandLog) != 3 {
		t.Fatalf("status = %+v with %d logged commands, want stopped after exactly 3 advances", status, len(art.CommandLog))
	}
	if got, want := commandsJSON(t, loggedCommands(art)), commandsJSON(t, scenario.Commands[:3]); got != want {
		t.Fatalf("command log is not the executed prefix:\n got %s\nwant %s", got, want)
	}
	if art.Incomplete || art.Termination == nil || art.Termination.Reason != model.TerminationStoppedByOperator || art.Termination.AtNS != reached {
		t.Fatalf("termination = %+v incomplete %v, want stopped by operator at the clock", art.Termination, art.Incomplete)
	}
	if art.Plan == nil || art.Plan.RequestedEndNS != started.StartNS+3600e9 {
		t.Fatalf("plan = %+v, want the requested end kept beside the early stop", art.Plan)
	}
	requireVerifies(t, d, started.WorldID)
}

func TestStoppingBeforeTheFirstAdvanceStillRunsTheStartActionsAndOneAdvance(t *testing.T) {
	t.Parallel()
	pacer := newGatedPacer()
	d := pacedBy(shippedDirector(t, t.Context()), pacer)
	req := pondJob()
	req.Faults[0].AtS = 0
	started := startJob(t, d, req)
	<-pacer.asked
	if _, err := d.StopJob(started.WorldID); err != nil {
		t.Fatal(err)
	}
	status := awaitJob(t, d, started.WorldID)
	art := loadRunArtifact(t, d, started.WorldID)
	ops := []string{}
	for _, command := range art.CommandLog {
		ops = append(ops, command.Op)
	}
	if status.State != rules.JobStopped || strings.Join(ops, " ") != "fault.inject clock.advance" {
		t.Fatalf("state %s, commands %v, want the time-0 fault and one advance", status.State, ops)
	}
	if status.ReachedNS <= started.StartNS || art.Termination == nil || art.Termination.AtNS != status.ReachedNS {
		t.Fatalf("reached %d termination %+v, want the stop recorded at the first boundary", status.ReachedNS, art.Termination)
	}
	requireVerifies(t, d, started.WorldID)
}

func TestStopJobIsIdempotentAndNamesUnknownWorlds(t *testing.T) {
	t.Parallel()
	pacer := newGatedPacer()
	d := pacedBy(shippedDirector(t, t.Context()), pacer)
	started := startJob(t, d, pondJob())
	<-pacer.asked
	for range 3 {
		if _, err := d.StopJob(started.WorldID); err != nil {
			t.Fatalf("repeated StopJob: %v", err)
		}
	}
	awaitJob(t, d, started.WorldID)
	finished, err := d.StopJob(started.WorldID)
	if err != nil || finished.State != rules.JobStopped {
		t.Fatalf("StopJob after the end = %+v, %v, want the final state untouched", finished, err)
	}
	var tool *rules.ToolError
	if _, err := d.StopJob("w-99"); !errors.As(err, &tool) || tool.Code != rules.CodeWorldNotFound {
		t.Fatalf("unknown world: %v", err)
	}
	if _, err := d.JobStatus("w-99"); !errors.As(err, &tool) || tool.Code != rules.CodeWorldNotFound {
		t.Fatalf("unknown world status: %v", err)
	}
}

func TestDestroyingARunningJobsWorldStopsTheJobFirst(t *testing.T) {
	t.Parallel()
	pacer := newGatedPacer()
	d := pacedBy(shippedDirector(t, t.Context()), pacer)
	started := startJob(t, d, pondJob())
	<-pacer.asked
	if _, err := d.DestroyWorld(started.WorldID); err != nil {
		t.Fatal(err)
	}
	if d.World(started.WorldID) != nil {
		t.Fatal("the world is still registered")
	}
	status, err := d.JobStatus(started.WorldID)
	if err != nil || status.State != rules.JobStopped {
		t.Fatalf("job after destroy = %+v, %v, want it stopped and still listed", status, err)
	}
	art := loadRunArtifact(t, d, started.WorldID)
	if art.Incomplete || art.Termination == nil || art.Termination.Reason != model.TerminationStoppedByOperator {
		t.Fatalf("run = incomplete %v termination %+v, want a complete run stopped by the operator", art.Incomplete, art.Termination)
	}
}

func TestAnExecutorErrorLeavesAnIncompleteRunWithTheReason(t *testing.T) {
	t.Parallel()
	pacer := newGatedPacer()
	d := pacedBy(shippedDirector(t, t.Context()), pacer)
	started := startJob(t, d, pondJob())
	<-pacer.asked
	if _, err := d.Advance(t.Context(), started.WorldID, started.StartNS+3000e9, false); err != nil {
		t.Fatal(err)
	}
	pacer.release(1)
	status := awaitJob(t, d, started.WorldID)
	if status.State != rules.JobFailed || !strings.Contains(status.Failure, "backwards") {
		t.Fatalf("status = %+v, want a failed job naming the clock going backwards", status)
	}
	art := loadRunArtifact(t, d, started.WorldID)
	if !art.Incomplete || !strings.Contains(art.Error, "backwards") || art.Termination != nil {
		t.Fatalf("run = incomplete %v error %q termination %+v, want an incomplete run with the reason", art.Incomplete, art.Error, art.Termination)
	}
}

func TestShuttingTheDirectorDownEndsTheJobAsFailed(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(t.Context())
	pacer := newGatedPacer()
	d := pacedBy(shippedDirector(t, ctx), pacer)
	started := startJob(t, d, pondJob())
	<-pacer.asked
	cancel()
	status := awaitJob(t, d, started.WorldID)
	if status.State != rules.JobFailed || !strings.Contains(status.Failure, "canceled") {
		t.Fatalf("status = %+v, want a failed job after the director's context ended", status)
	}
	if art := loadRunArtifact(t, d, started.WorldID); !art.Incomplete {
		t.Fatal("the run must be recorded incomplete")
	}
}
