package app

import (
	"testing"
	"time"

	rules "github.com/ghassan-ai-projects/streams-simulator/internal/director/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestStoppingAPausedJobEndsItAtTheReachedInstantAndTheRunVerifies(t *testing.T) {
	t.Parallel()
	pacer := newGatedPacer()
	d := pacedBy(shippedDirector(t, t.Context()), pacer)
	started := startJob(t, d, pondJob())
	pacer.release(2)
	for range 3 {
		<-pacer.asked
	}
	if _, err := d.PauseJob(started.WorldID); err != nil {
		t.Fatal(err)
	}
	if stopping, err := d.StopJob(started.WorldID); err != nil || stopping.State != rules.JobStopping {
		t.Fatalf("StopJob = %+v, %v, want the paused job stopping", stopping, err)
	}
	status := awaitJob(t, d, started.WorldID)
	art := loadRunArtifact(t, d, started.WorldID)
	reached := started.StartNS + 2*120e9
	if status.State != rules.JobStopped || status.ReachedNS != reached || len(art.CommandLog) != 2 {
		t.Fatalf("status = %+v with %d commands, want stopped after exactly the 2 chunks run before the pause", status, len(art.CommandLog))
	}
	if art.Termination == nil || art.Termination.Reason != model.TerminationStoppedByOperator || art.Termination.AtNS != reached || art.Incomplete {
		t.Fatalf("termination = %+v, want stopped by the operator at the paused instant", art.Termination)
	}
	requireVerifies(t, d, started.WorldID)
}

func TestStoppingAJobPausedBeforeItsFirstAdvanceStillRunsOneAdvance(t *testing.T) {
	t.Parallel()
	pacer := newGatedPacer()
	d := pacedBy(shippedDirector(t, t.Context()), pacer)
	started := startJob(t, d, pondJob())
	<-pacer.asked
	if _, err := d.PauseJob(started.WorldID); err != nil {
		t.Fatal(err)
	}
	if _, err := d.StopJob(started.WorldID); err != nil {
		t.Fatal(err)
	}
	status := awaitJob(t, d, started.WorldID)
	if status.State != rules.JobStopped || status.ReachedNS != started.StartNS+120e9 {
		t.Fatalf("status = %+v, want stopped after the one advance a stop always allows", status)
	}
	requireVerifies(t, d, started.WorldID)
}

func TestDestroyingAPausedJobsWorldReleasesItsGoroutine(t *testing.T) {
	t.Parallel()
	pacer := newGatedPacer()
	d := pacedBy(shippedDirector(t, t.Context()), pacer)
	started := startJob(t, d, pondJob())
	<-pacer.asked
	if _, err := d.PauseJob(started.WorldID); err != nil {
		t.Fatal(err)
	}
	if _, err := d.DestroyWorld(started.WorldID); err != nil {
		t.Fatal(err)
	}
	select {
	case <-d.jobs.find(started.WorldID).done:
	case <-time.After(time.Minute):
		t.Fatal("the paused job's goroutine is still parked after its world was destroyed")
	}
	if status, _ := d.JobStatus(started.WorldID); status.State != rules.JobStopped {
		t.Fatalf("status = %+v, want stopped", status)
	}
}

func TestShuttingTheDirectorDownReleasesAPausedJobAsFailed(t *testing.T) {
	t.Parallel()
	ctx, cancel := contextWithCancel(t)
	pacer := newGatedPacer()
	d := pacedBy(shippedDirector(t, ctx), pacer)
	started := startJob(t, d, pondJob())
	<-pacer.asked
	if _, err := d.PauseJob(started.WorldID); err != nil {
		t.Fatal(err)
	}
	cancel()
	if status := awaitJob(t, d, started.WorldID); status.State != rules.JobFailed {
		t.Fatalf("status = %+v, want a failed job once the director's context ended", status)
	}
}

func TestStatusNeverBlocksWhileAJobIsPausedOrAPauseWaitsForTheBoundary(t *testing.T) {
	t.Parallel()
	pacer := newGatedPacer()
	d := pacedBy(shippedDirector(t, t.Context()), pacer)
	started := startJob(t, d, pondJob())
	<-pacer.asked
	if _, err := d.PauseJob(started.WorldID); err != nil {
		t.Fatal(err)
	}
	got := make(chan rules.JobStatus, 1)
	go func() { status, _ := d.JobStatus(started.WorldID); got <- status }()
	select {
	case status := <-got:
		if status.State != rules.JobPaused {
			t.Fatalf("status = %+v, want paused", status)
		}
	case <-time.After(20 * time.Second):
		t.Fatal("JobStatus blocked behind a paused job")
	}
	if _, err := d.StopJob(started.WorldID); err != nil {
		t.Fatal(err)
	}
	awaitJob(t, d, started.WorldID)
}
