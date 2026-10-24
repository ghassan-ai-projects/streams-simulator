package app

import (
	"path/filepath"
	"testing"

	rules "github.com/ghassan-ai-projects/streams-simulator/internal/director/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func controlPlans() map[string]rules.JobRequest {
	line := lineJob()
	line.Rate = 60
	return map[string]rules.JobRequest{"pond with a fault and a call": pondJob(), "line with an interlock refusal": line}
}

func driveWithControls(t *testing.T, d *Director, pacer *gatedPacer, worldID string) {
	t.Helper()
	must := func(_ rules.JobStatus, err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	<-pacer.asked
	must(d.PauseJob(worldID))
	must(d.StepJob(worldID))
	must(d.StepJob(worldID))
	must(d.SetJobRate(worldID, 10))
	must(d.ResumeJob(worldID))
	<-pacer.asked
	must(d.PauseJob(worldID))
	must(d.StepJob(worldID))
	must(d.SetJobRate(worldID, 0))
	must(d.ResumeJob(worldID))
	pacer.release(1000)
}

func TestPausesStepsAndRateChangesNeverAlterTheWorldOrTheCommandLog(t *testing.T) {
	t.Parallel()
	for name, req := range controlPlans() {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			straight := pacedBy(shippedDirector(t, t.Context()), freePacer{})
			plain := startJob(t, straight, req)
			awaitJob(t, straight, plain.WorldID)
			plainArt := loadRunArtifact(t, straight, plain.WorldID)

			pacer := newGatedPacer()
			controlled := pacedBy(shippedDirector(t, t.Context()), pacer)
			started := startJob(t, controlled, req)
			driveWithControls(t, controlled, pacer, started.WorldID)
			status := awaitJob(t, controlled, started.WorldID)
			art := requireReplaysFromTheRunFile(t, controlled, started.WorldID)

			if status.State != rules.JobCompleted || status.Rate != 0 {
				t.Fatalf("status = %+v, want the controlled job completed at the last rate set", status)
			}
			if plainArt.ExpectedTraceDigest == "" || art.ExpectedTraceDigest != plainArt.ExpectedTraceDigest || art.WorldDigest != plainArt.WorldDigest {
				t.Fatalf("trace digest straight %s controlled %s, want equal", plainArt.ExpectedTraceDigest, art.ExpectedTraceDigest)
			}
			if commandsJSON(t, art.CommandLog) != commandsJSON(t, plainArt.CommandLog) {
				t.Fatal("the command log of the controlled job differs from the straight run")
			}
			requireMatchesScenarioRun(t, req, art)
		})
	}
}

func requireMatchesScenarioRun(t *testing.T, req rules.JobRequest, jobArt *model.RunArtifact) {
	t.Helper()
	scenario := rules.PlanScenario(req, "native-jsonl", model.DefaultStartTimeNS)
	dir := t.TempDir()
	summary, err := run.RunScenario(t.Context(), &scenario, run.ScenarioRunOptions{DomainsDir: testsupport.DomainsDir(),
		AdaptersDir: testsupport.AdaptersDir(), SinkName: model.SinkInproc, OutDir: dir, Explore: req.Explore})
	if err != nil {
		t.Fatal(err)
	}
	scenarioArt, err := run.LoadArtifact(filepath.Join(dir, model.FileRunArtifact))
	if err != nil {
		t.Fatal(err)
	}
	if summary.TraceDigest != jobArt.ExpectedTraceDigest || commandsJSON(t, scenarioArt.CommandLog) != commandsJSON(t, jobArt.CommandLog) {
		t.Fatalf("RunScenario trace %s, controlled job %s, want one implementation", summary.TraceDigest, jobArt.ExpectedTraceDigest)
	}
}
