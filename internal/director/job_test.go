package director_test

import (
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/director"
)

func pondJob(rate float64) director.JobRequest {
	return director.JobRequest{Domain: "aquaculture-pond", Seed: 4, DurationS: 1800, Rate: rate,
		Faults: []director.JobFault{{EntityID: "site-a/pond-1", Fault: "aerator_failure", AtS: 300}}}
}

func settled(t *testing.T, d *director.Director, worldID string) director.JobStatus {
	t.Helper()
	deadline := time.After(time.Minute)
	for {
		status, err := d.JobStatus(worldID)
		if err != nil {
			t.Fatal(err)
		}
		if status.State != director.JobRunning && status.State != director.JobStopping {
			return status
		}
		select {
		case <-deadline:
			t.Fatalf("job %s never finished: %+v", worldID, status)
		default:
			runtime.Gosched()
		}
	}
}

func TestFacadeRunsAnUnpacedJobToCompletion(t *testing.T) {
	t.Parallel()
	d := newTestDirector(t)
	started, err := d.StartJob(pondJob(0))
	if err != nil || started.State != director.JobRunning || started.WorldID == "" {
		t.Fatalf("StartJob = %+v, %v", started, err)
	}
	status := settled(t, d, started.WorldID)
	if status.State != director.JobCompleted || status.ReachedNS != status.RequestedEndNS {
		t.Fatalf("status = %+v, want completed at the requested end", status)
	}
	if listed := d.ListJobs(); len(listed) != 1 || listed[0].WorldID != started.WorldID {
		t.Fatalf("ListJobs = %+v", listed)
	}
}

func TestFacadeStopsAPacedJobOnTheWallPacer(t *testing.T) {
	t.Parallel()
	d := newTestDirector(t)
	started, err := d.StartJob(pondJob(1))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := d.StartJob(pondJob(1)); err == nil || !strings.HasPrefix(err.Error(), string(director.CodeJobActive)) {
		t.Fatalf("a second job must be refused while one runs: %v", err)
	}
	if _, err := d.StopJob(started.WorldID); err != nil {
		t.Fatal(err)
	}
	if _, err := d.DestroyWorld(started.WorldID); err != nil {
		t.Fatal(err)
	}
	status, err := d.JobStatus(started.WorldID)
	if err != nil || status.State != director.JobStopped || status.ReachedNS <= status.StartNS || status.ReachedNS >= status.RequestedEndNS {
		t.Fatalf("status = %+v, %v, want a stopped job that ran part of its plan", status, err)
	}
}
