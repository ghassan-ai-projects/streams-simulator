package app

import (
	"fmt"
	"sync"
	"testing"
	"time"

	rules "github.com/ghassan-ai-projects/streams-simulator/internal/director/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

var aeratorArgs = map[string]any{"pond_id": pondEntity}

func pausedAt(t *testing.T, d *Director, pacer *gatedPacer, req rules.JobRequest, chunks int) rules.JobStatus {
	t.Helper()
	started := startJob(t, d, req)
	pacer.release(chunks)
	for range chunks + 1 {
		<-pacer.asked
	}
	if _, err := d.PauseJob(started.WorldID); err != nil {
		t.Fatal(err)
	}
	return started
}

func finish(t *testing.T, d *Director, pacer *gatedPacer, worldID string) *model.RunArtifact {
	t.Helper()
	if _, err := d.ResumeJob(worldID); err != nil {
		t.Fatal(err)
	}
	pacer.release(1000)
	awaitJob(t, d, worldID)
	return requireReplaysFromTheRunFile(t, d, worldID)
}

func TestAnEffectorCalledNowRunsAtThePausedInstantAndIsRecordedAsAnImmediateOutcome(t *testing.T) {
	t.Parallel()
	pacer := newGatedPacer()
	d := pacedBy(shippedDirector(t, t.Context()), pacer)
	started := pausedAt(t, d, pacer, pondJob(), 2)
	res, err := d.JobInvokeEffector(started.WorldID, pondEntity, "start_aerator", aeratorArgs)
	if err != nil || res.Call == nil || !res.Call.Accepted || !res.Call.Immediate || res.Call.AtS != 240 || res.State != rules.JobPaused {
		t.Fatalf("JobInvokeEffector = %+v, %v, want an accepted immediate call at 240 s on a still paused job", res, err)
	}
	if len(res.Calls) != 1 || res.Calls[0] != *res.Call {
		t.Fatalf("calls = %+v, want the immediate outcome listed in the job", res.Calls)
	}
	art := finish(t, d, pacer, started.WorldID)
	logged := art.CommandLog[2]
	if logged.Op != model.OpEffectorInvoke || logged.AtNS != started.StartNS+240e9 || logged.Args["command_id"] != "job-now-1" {
		t.Fatalf("command 2 = %+v, want the call logged after the second advance with its own command id", logged)
	}
	if status, _ := d.JobStatus(started.WorldID); len(status.Calls) != 2 || !status.Calls[0].Immediate || status.Calls[1].Immediate {
		t.Fatalf("calls = %+v, want the immediate outcome first and the planned one second", status.Calls)
	}
}

func TestAFaultInjectedNowStartsAtThePausedInstantAndReplays(t *testing.T) {
	t.Parallel()
	pacer := newGatedPacer()
	d := pacedBy(shippedDirector(t, t.Context()), pacer)
	started := pausedAt(t, d, pacer, explorePond(), 2)
	res, err := d.JobInjectFault(started.WorldID, pondEntity, "do_probe_fouling")
	if err != nil || res.FaultID == "" || res.State != rules.JobPaused || len(res.Calls) != 0 {
		t.Fatalf("JobInjectFault = %+v, %v, want a fault id on a still paused job", res, err)
	}
	art := finish(t, d, pacer, started.WorldID)
	at := started.StartNS + 240e9
	logged := art.CommandLog[2]
	if logged.Op != model.OpFaultInject || logged.AtNS != at || fmt.Sprint(logged.Args["onset_ns"]) != fmt.Sprint(at) || logged.Args["fault"] != "do_probe_fouling" {
		t.Fatalf("command 2 = %+v, want the fault logged at the clock with its onset at the clock", logged)
	}
}

func TestARefusedEffectorCallNowIsARecordedOutcomeLikeAPlannedOne(t *testing.T) {
	t.Parallel()
	pacer := newGatedPacer()
	d := pacedBy(shippedDirector(t, t.Context()), pacer)
	req := lineJob()
	req.Rate = 60
	started := pausedAt(t, d, pacer, req, 0)
	res, err := d.JobInvokeEffector(started.WorldID, "line-1", "reset_line", map[string]any{"line_id": "line-1"})
	if err != nil || res.Call == nil || res.Call.Accepted || res.Call.Reason != "interlock_refused" || !res.Call.Immediate {
		t.Fatalf("JobInvokeEffector = %+v, %v, want the interlock refusal as an immediate outcome", res, err)
	}
	finish(t, d, pacer, started.WorldID)
	if status, _ := d.JobStatus(started.WorldID); len(status.Calls) != 3 || !status.Calls[0].Immediate {
		t.Fatalf("calls = %+v, want the refusal then the two planned calls", status.Calls)
	}
}

func TestAnInvalidImmediateCommandIsRefusedAndTheJobIsUndisturbed(t *testing.T) {
	t.Parallel()
	pacer := newGatedPacer()
	d := pacedBy(shippedDirector(t, t.Context()), pacer)
	started := pausedAt(t, d, pacer, explorePond(), 1)
	id := started.WorldID
	_, err := d.JobInjectFault(id, pondEntity, "no-such-fault")
	requireToolCode(t, err, rules.CodeDomainInvalid)
	_, err = d.JobInjectFault(id, "site-z/none", "do_probe_fouling")
	requireToolCode(t, err, rules.CodeDomainInvalid)
	_, err = d.JobInjectFault(id, "", "do_probe_fouling")
	requireToolCode(t, err, rules.CodeInvalidArgs)
	_, err = d.JobInvokeEffector(id, pondEntity, "no-such-effector", nil)
	requireToolCode(t, err, rules.CodeUnknownEffector)
	_, err = d.JobInvokeEffector(id, "site-z/none", "start_aerator", aeratorArgs)
	requireToolCode(t, err, rules.CodeInvalidArgs)
	_, err = d.JobInvokeEffector(id, pondEntity, "", nil)
	requireToolCode(t, err, rules.CodeInvalidArgs)
	_, err = d.JobInvokeEffector("w-99", pondEntity, "start_aerator", aeratorArgs)
	requireToolCode(t, err, rules.CodeWorldNotFound)
	if status, _ := d.JobStatus(id); status.State != rules.JobPaused || len(status.Calls) != 0 || status.ReachedNS != started.StartNS+120e9 {
		t.Fatalf("status = %+v, want the paused job untouched by refused commands", status)
	}
	if art := finish(t, d, pacer, id); art.Incomplete {
		t.Fatal("the run is incomplete after refused immediate commands")
	}
}

func TestImmediateCommandsAreRefusedOnceTheJobIsNotRunningOrPaused(t *testing.T) {
	t.Parallel()
	d := pacedBy(shippedDirector(t, t.Context()), freePacer{})
	started := startJob(t, d, explorePond())
	awaitJob(t, d, started.WorldID)
	_, err := d.JobInjectFault(started.WorldID, pondEntity, "do_probe_fouling")
	requireToolCode(t, err, rules.CodeJobNotRunning)
	_, err = d.JobInvokeEffector(started.WorldID, pondEntity, "start_aerator", aeratorArgs)
	requireToolCode(t, err, rules.CodeJobNotRunning)
}

func TestACommandThatArrivesAfterTheExecutorFinishedIsRefusedBeforeTheRunEnds(t *testing.T) {
	t.Parallel()
	pacer := newGatedPacer()
	d := pacedBy(shippedDirector(t, t.Context()), pacer)
	started := pausedAt(t, d, pacer, explorePond(), 1)
	j := d.jobs.find(started.WorldID)
	j.closeImmediate()
	_, err := d.JobInjectFault(started.WorldID, pondEntity, "do_probe_fouling")
	requireToolCode(t, err, rules.CodeJobNotRunning)
	if _, err := d.StopJob(started.WorldID); err != nil {
		t.Fatal(err)
	}
	awaitJob(t, d, started.WorldID)
}

func TestAnImmediateCommandWaitsForAnAdvanceInFlightWhileStatusStaysLive(t *testing.T) {
	t.Parallel()
	pacer := newGatedPacer()
	d := pacedBy(shippedDirector(t, t.Context()), pacer)
	started := startJob(t, d, explorePond())
	<-pacer.asked
	inAdvance, unblock := make(chan struct{}), make(chan struct{})
	var once sync.Once
	d.RunOf(started.WorldID).SetDeliveryObserver(func(model.SimEvent) {
		once.Do(func() { close(inAdvance) })
		<-unblock
	})
	pacer.release(1)
	<-inAdvance
	applied := make(chan rules.JobImmediate, 1)
	go func() {
		res, _ := d.JobInjectFault(started.WorldID, pondEntity, "do_probe_fouling")
		applied <- res
	}()
	seen := make(chan rules.JobStatus, 1)
	go func() { status, _ := d.JobStatus(started.WorldID); seen <- status }()
	select {
	case status := <-seen:
		if status.ReachedNS != started.StartNS {
			t.Fatalf("status = %+v, want the advance not yet reached", status)
		}
	case <-time.After(20 * time.Second):
		t.Fatal("JobStatus blocked behind an advance and an immediate command")
	}
	close(unblock)
	if res := <-applied; res.FaultID == "" {
		t.Fatalf("immediate result = %+v, want the fault applied once the advance was done", res)
	}
	pacer.release(1000)
	awaitJob(t, d, started.WorldID)
	art := requireReplaysFromTheRunFile(t, d, started.WorldID)
	if logged := art.CommandLog[1]; logged.Op != model.OpFaultInject || logged.AtNS != started.StartNS+120e9 {
		t.Fatalf("command 1 = %+v, want the fault logged after the in-flight advance, at its clock", logged)
	}
}
