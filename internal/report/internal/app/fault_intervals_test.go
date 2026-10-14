package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonschema"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/schemas"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

const second = int64(1e9)

func clearedRun(t *testing.T, dir string, explore bool) (start int64) {
	t.Helper()
	spec, err := domain.Load(testsupport.Domain("aquaculture-pond"))
	if err != nil {
		t.Fatal(err)
	}
	native, err := adapter.Load(testsupport.Adapter("native-jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	start = model.DefaultStartTimeNS
	r, err := run.New(t.Context(), run.Config{Domain: spec, Adapter: native, Seed: 5, SinkName: model.SinkInproc,
		TimeMode: model.TimeStepped, StartTimeNS: start, StartTimeSet: true, Explore: explore, JournalDir: dir})
	if err != nil {
		t.Fatal(err)
	}
	id, err := r.InjectFault(pond, "aerator_failure", start+60*second, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Advance(t.Context(), start+120*second, false); err != nil {
		t.Fatal(err)
	}
	if err := r.ClearFault(id, start+300*second); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Advance(t.Context(), start+400*second, false); err != nil {
		t.Fatal(err)
	}
	if _, err := r.End(dir); err != nil {
		t.Fatal(err)
	}
	return start
}

func TestReportGivesEachFaultAsAnIntervalInExploreAndEvaluateRuns(t *testing.T) {
	t.Parallel()
	for _, explore := range []bool{true, false} {
		dir := t.TempDir()
		start := clearedRun(t, dir, explore)
		built, err := Build(t.Context(), dir)
		if err != nil {
			t.Fatal(err)
		}
		if built.Truth == nil || len(built.Truth.Faults) != 1 {
			t.Fatalf("explore=%v: truth = %+v, want one fault", explore, built.Truth)
		}
		fault := built.Truth.Faults[0]
		if fault.FaultID != "f-0" || fault.EntityID != pond || fault.Fault != "aerator_failure" || fault.InjectedAtNS != start ||
			fault.OnsetNS != start+60*second || fault.ClearedNS == nil || *fault.ClearedNS != start+300*second ||
			fault.ClearedBy == nil || *fault.ClearedBy != model.ClearedByCommand {
			t.Fatalf("explore=%v: fault = %+v, want f-0 cleared by command at +300 s with onset +60 s", explore, fault)
		}
	}
}

func TestReportFaultIntervalsMatchTheRunReportContract(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	clearedRun(t, dir, false)
	schema, err := jsonschema.CompileJSON(schemas.RunReport())
	if err != nil {
		t.Fatal(err)
	}
	var doc any
	if err := model.DecodeBytes(render(t, dir, JSONFormat), &doc); err != nil {
		t.Fatal(err)
	}
	if errs := schema.Validate(doc); len(errs) > 0 {
		t.Fatalf("report does not match run-report-v0.1:\n%s", jsonschema.FormatErrors(errs))
	}
	if text := string(render(t, dir, JSONFormat)); !strings.Contains(text, `"cleared_by": "command"`) {
		t.Fatalf("report lacks the cleared_by of the fault:\n%s", text)
	}
}

func TestAnOpenEvaluateRunsFaultIntervalsStaySealedAndAnOpenExploreRunsDoNot(t *testing.T) {
	t.Parallel()
	for _, explore := range []bool{true, false} {
		dir := t.TempDir()
		clearedRun(t, dir, explore)
		if err := os.Remove(filepath.Join(dir, "run.json")); err != nil {
			t.Fatal(err)
		}
		built, err := Build(t.Context(), dir)
		if err != nil {
			t.Fatal(err)
		}
		if explore != (built.Truth != nil && len(built.Truth.Faults) == 1) {
			t.Fatalf("explore=%v: truth = %+v, want the interval only for the explore run", explore, built.Truth)
		}
		if !explore && strings.Contains(string(render(t, dir, JSONFormat)), "f-0") {
			t.Fatal("the report of an open evaluate run names a fault id")
		}
	}
}
