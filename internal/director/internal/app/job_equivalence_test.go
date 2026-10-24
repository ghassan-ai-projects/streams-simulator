package app

import (
	"path/filepath"
	"testing"

	rules "github.com/ghassan-ai-projects/streams-simulator/internal/director/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func plans() map[string]rules.JobRequest {
	return map[string]rules.JobRequest{
		"pond with a fault and a call":                          pondJob(),
		"pond as fast as possible":                              func() rules.JobRequest { r := pondJob(); r.Rate = 0; return r }(),
		"line with an interlock refusal":                        lineJob(),
		"pond with a recovery, a delivery condition and a name": recoveryPondJob(),
	}
}

func TestAJobAndRunScenarioOfTheSamePlanProduceTheSameTrace(t *testing.T) {
	t.Parallel()
	for name, req := range plans() {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			d := pacedBy(shippedDirector(t, t.Context()), freePacer{})
			started := startJob(t, d, req)
			awaitJob(t, d, started.WorldID)
			jobArt := loadRunArtifact(t, d, started.WorldID)

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
			if jobArt.ExpectedTraceDigest == "" || jobArt.ExpectedTraceDigest != summary.TraceDigest {
				t.Fatalf("trace digest: job %s, RunScenario %s", jobArt.ExpectedTraceDigest, summary.TraceDigest)
			}
			if jobArt.WorldDigest != scenarioArt.WorldDigest || commandsJSON(t, jobArt.CommandLog) != commandsJSON(t, scenarioArt.CommandLog) {
				t.Fatalf("world digest %s/%s or command log differ", jobArt.WorldDigest, scenarioArt.WorldDigest)
			}
			requireVerifies(t, d, started.WorldID)
		})
	}
}
