package app

import (
	"sync"
	"testing"
	"time"

	rules "github.com/ghassan-ai-projects/streams-simulator/internal/director/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestStatusProgressesWithEachAdvanceAndNeverBlocksBehindOne(t *testing.T) {
	t.Parallel()
	pacer := newGatedPacer()
	d := pacedBy(shippedDirector(t, t.Context()), pacer)
	started := startJob(t, d, pondJob())
	<-pacer.asked
	inAdvance, unblock := make(chan struct{}), make(chan struct{})
	var once sync.Once
	d.RunOf(started.WorldID).SetDeliveryObserver(func(model.SimEvent) {
		once.Do(func() { close(inAdvance) })
		<-unblock
	})
	pacer.release(1)
	<-inAdvance
	got := make(chan rules.JobStatus, 1)
	go func() {
		status, _ := d.JobStatus(started.WorldID)
		got <- status
	}()
	select {
	case status := <-got:
		if status.State != rules.JobRunning || status.ReachedNS != started.StartNS {
			t.Fatalf("status during an advance = %+v, want running and not yet reached", status)
		}
	case <-time.After(20 * time.Second):
		t.Fatal("JobStatus blocked behind an in-flight advance")
	}
	close(unblock)
	pacer.release(1)
	<-pacer.asked
	if status, _ := d.JobStatus(started.WorldID); status.ReachedNS != started.StartNS+120e9 {
		t.Fatalf("reached = %d, want the first chunk reached once its advance finished", status.ReachedNS)
	}
	pacer.release(1000)
	awaitJob(t, d, started.WorldID)
}

func TestStatusAndStopAreSafeUnderConcurrentUse(t *testing.T) {
	t.Parallel()
	d := pacedBy(shippedDirector(t, t.Context()), freePacer{})
	started := startJob(t, d, pondJob())
	var wg sync.WaitGroup
	for range 8 {
		wg.Add(2)
		go func() { defer wg.Done(); _, _ = d.JobStatus(started.WorldID); _ = d.ListJobs() }()
		go func() { defer wg.Done(); _, _ = d.StopJob(started.WorldID) }()
	}
	wg.Wait()
	status := awaitJob(t, d, started.WorldID)
	if rules.JobActive(status.State) {
		t.Fatalf("status = %+v, want a finished job", status)
	}
}

func TestAFinishedJobsStatusIsAStableSnapshot(t *testing.T) {
	t.Parallel()
	d := pacedBy(shippedDirector(t, t.Context()), freePacer{})
	started := startJob(t, d, pondJob())
	first := awaitJob(t, d, started.WorldID)
	first.Calls[0].Effector = "tampered"
	if again, _ := d.JobStatus(started.WorldID); again.Calls[0].Effector != "start_aerator" {
		t.Fatalf("a caller changed the job's own record: %+v", again.Calls)
	}
}
