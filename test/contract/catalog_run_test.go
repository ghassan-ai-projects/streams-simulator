package contract

import (
	"context"
	"fmt"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/lint"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

const (
	catalogSeed    = 11
	catalogRunNS   = 2 * 3600e9
	catalogLintHrs = 24
)

func checkCatalogLint(t *testing.T, p catalogPlant) {
	report, err := lint.Check(p.spec, lint.Options{Seed: 1, Hours: catalogLintHrs})
	if err != nil {
		t.Fatal(err)
	}
	for _, finding := range report.Findings {
		switch finding.Severity {
		case "error":
			t.Errorf("lint error %s on %s %s%s at %.0f s: %s", finding.Code, finding.Entity, finding.Channel, finding.State, finding.AtS, finding.Detail)
		default:
			t.Logf("lint warning %s on %s %s%s: %s", finding.Code, finding.Entity, finding.Channel, finding.State, finding.Detail)
		}
	}
}

func checkCatalogReplay(t *testing.T, p catalogPlant) {
	ctx := context.Background()
	native, err := adapter.Load(testsupport.Adapter("native-jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	r, err := run.New(ctx, run.Config{Domain: p.spec, Adapter: native, Seed: catalogSeed, SinkName: model.SinkInproc,
		TimeMode: model.TimeStepped, StartTimeNS: model.DefaultStartTimeNS})
	if err != nil {
		t.Fatal(err)
	}
	driveCatalogRun(t, r, p)
	if _, err := r.Advance(ctx, model.DefaultStartTimeNS+catalogRunNS, false); err != nil {
		t.Fatal(err)
	}
	art, err := r.End("")
	if err != nil || art.Counts.Emitted == 0 {
		t.Fatalf("end = %+v, %v", art, err)
	}
	result, err := run.ReplayArtifact(ctx, art, p.spec, nil, "")
	if err != nil || !result.Matches {
		t.Fatalf("replay = %+v (%v)", result, err)
	}
}

func driveCatalogRun(t *testing.T, r *run.Run, p catalogPlant) {
	t.Helper()
	entity := p.entities[0]
	for i, effector := range p.kind().Effectors {
		at := model.DefaultStartTimeNS + int64(600+i*60)*1e9
		args := catalogArgs(effector.ArgsSchema, entity, 0)
		if _, err := r.InvokeEffector(effector.Name, entity, fmt.Sprintf("replay-%d", i), args, at); err != nil {
			t.Logf("effector %s refused: %v", effector.Name, err)
		}
	}
	for i, fault := range p.kind().Faults {
		at := model.DefaultStartTimeNS + int64(1800+i*60)*1e9
		if _, err := r.InjectFault(entity, fault.ID, at, nil); err != nil {
			t.Fatalf("inject %s: %v", fault.ID, err)
		}
	}
}
