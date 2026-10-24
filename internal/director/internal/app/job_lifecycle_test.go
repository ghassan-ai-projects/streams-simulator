package app

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	rules "github.com/ghassan-ai-projects/streams-simulator/internal/director/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestStartJobReturnsAtOnceWithTheJobRunningOnAnOrdinaryWorld(t *testing.T) {
	t.Parallel()
	pacer := newGatedPacer()
	d := pacedBy(shippedDirector(t, t.Context()), pacer)
	status := startJob(t, d, pondJob())
	if status.WorldID != "w-1" || status.RunID != "r-1" || status.State != rules.JobRunning {
		t.Fatalf("status = %+v, want running job on w-1", status)
	}
	if status.StartNS != model.DefaultStartTimeNS || status.ReachedNS != status.StartNS || status.RequestedEndNS != status.StartNS+3600e9 || status.Rate != 60 || len(status.Calls) != 0 {
		t.Fatalf("status = %+v, want the plan echoed and nothing reached", status)
	}
	world := d.World("w-1")
	if world == nil || world.Token == "" || world.Run.ID != "r-1" {
		t.Fatalf("world record = %+v, want an ordinary registered world with a token", world)
	}
	if _, err := os.Stat(filepath.Join(d.OutDir, "w-1")); err != nil {
		t.Fatalf("the run folder is missing: %v", err)
	}
	pacer.release(1000)
	awaitJob(t, d, "w-1")
}

func TestAJobRunsItsPlanToCompletionAndRecordsIntent(t *testing.T) {
	t.Parallel()
	d := pacedBy(shippedDirector(t, t.Context()), freePacer{})
	req := pondJob()
	started := startJob(t, d, req)
	status := awaitJob(t, d, started.WorldID)
	end := started.StartNS + 3600e9
	if status.State != rules.JobCompleted || status.ReachedNS != end || status.Failure != "" {
		t.Fatalf("status = %+v, want completed at the requested end", status)
	}
	if len(status.Calls) != 1 || status.Calls[0].EntityID != pondEntity || status.Calls[0].Effector != "start_aerator" || status.Calls[0].AtS != 1800 || !status.Calls[0].Accepted || status.Calls[0].EffectETANS == 0 {
		t.Fatalf("calls = %+v, want the accepted aerator call with its ETA", status.Calls)
	}
	art := loadRunArtifact(t, d, started.WorldID)
	if art.Incomplete || art.Plan == nil || art.Plan.RequestedEndNS != end || art.Plan.Rate != 60 {
		t.Fatalf("run.json plan = %+v incomplete %v, want the requested end and rate", art.Plan, art.Incomplete)
	}
	if art.Termination == nil || art.Termination.Reason != model.TerminationCompleted || art.Termination.AtNS != end {
		t.Fatalf("termination = %+v, want completed at the end", art.Termination)
	}
	scenario := rules.PlanScenario(req, "native-jsonl", started.StartNS)
	if got, want := commandsJSON(t, loggedCommands(art)), commandsJSON(t, scenario.Commands); got != want {
		t.Fatalf("command log differs from the planned scenario:\n got %s\nwant %s", got, want)
	}
	requireVerifies(t, d, started.WorldID)
}

func TestAnUnpacedJobNeverAsksThePacer(t *testing.T) {
	t.Parallel()
	pacer := newGatedPacer()
	d := pacedBy(shippedDirector(t, t.Context()), pacer)
	req := pondJob()
	req.Rate = 0
	started := startJob(t, d, req)
	if status := awaitJob(t, d, started.WorldID); status.State != rules.JobCompleted {
		t.Fatalf("status = %+v, want completed without any pacing", status)
	}
	if len(pacer.asked) != 0 {
		t.Fatalf("the pacer was asked %d times for an unpaced job", len(pacer.asked))
	}
}

func TestAPacedJobAsksThePacerOncePerAdvanceWithTheElapsedSimulatedTime(t *testing.T) {
	t.Parallel()
	pacer := newGatedPacer()
	d := pacedBy(shippedDirector(t, t.Context()), pacer)
	req := rules.JobRequest{Domain: "aquaculture-pond", Seed: 1, DurationS: 60, Rate: 10}
	started := startJob(t, d, req)
	pacer.release(100)
	awaitJob(t, d, started.WorldID)
	asked := len(pacer.asked)
	first, last := <-pacer.asked, int64(0)
	for range asked - 1 {
		last = <-pacer.asked
	}
	if asked != 30 || first != 2e9 || last != 60e9 {
		t.Fatalf("pacer asked %d times, first %d last %d, want 30 waits from 2 s to 60 s", asked, first, last)
	}
}

func TestSecondJobIsRefusedWhileOneRunsAndAcceptedAfterwards(t *testing.T) {
	t.Parallel()
	pacer := newGatedPacer()
	d := pacedBy(shippedDirector(t, t.Context()), pacer)
	first := startJob(t, d, pondJob())
	var tool *rules.ToolError
	if _, err := d.StartJob(pondJob()); !errors.As(err, &tool) || tool.Code != rules.CodeJobActive {
		t.Fatalf("second start = %v, want a %s refusal", err, rules.CodeJobActive)
	}
	if len(d.Worlds) != 1 {
		t.Fatalf("%d worlds, want the refused start to create none", len(d.Worlds))
	}
	pacer.release(1000)
	awaitJob(t, d, first.WorldID)
	second := startJob(t, d, pondJob())
	if second.WorldID != "w-2" {
		t.Fatalf("second job world = %s, want w-2", second.WorldID)
	}
	awaitJob(t, d, second.WorldID)
	if listed := d.ListJobs(); len(listed) != 2 || listed[0].WorldID != "w-1" || listed[1].WorldID != "w-2" || listed[0].State != rules.JobCompleted {
		t.Fatalf("listed = %+v, want both finished jobs in start order", listed)
	}
}
