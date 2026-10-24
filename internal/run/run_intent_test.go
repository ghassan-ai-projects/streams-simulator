package run_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func stoppedRun(t *testing.T, stopAt int, intent bool) (string, *model.RunArtifact) {
	t.Helper()
	start := model.DefaultStartTimeNS
	scenario := scenarioFor(start)
	scenario.Commands = nil
	for i := 1; i <= 6; i++ {
		scenario.Commands = append(scenario.Commands, model.ScenarioCommand{Op: model.OpClockAdvance, Args: map[string]any{"to_ns": start + int64(i)*600e9}})
	}
	spec, adapterSpec, err := run.LoadInputs(testsupport.DomainsDir(), testsupport.AdaptersDir(), scenario.Domain, scenario.Adapter)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	r, err := run.New(t.Context(), run.Config{Domain: spec, Adapter: adapterSpec, Seed: scenario.Seed, SinkName: model.SinkInproc,
		TimeMode: model.TimeStepped, StartTimeNS: start, StartTimeSet: true, JournalDir: dir})
	if err != nil {
		t.Fatal(err)
	}
	hooks := run.ScenarioHooks{Before: func(_ context.Context, index int, _ model.ScenarioCommand) (bool, error) { return index == stopAt, nil }}
	outcome, err := r.ExecuteScenarioWith(t.Context(), scenario, hooks)
	if err != nil || outcome.Executed != stopAt || !outcome.Stopped {
		t.Fatalf("outcome = %+v, %v", outcome, err)
	}
	if intent {
		recordIntent(t, r, start+int64(stopAt)*600e9)
	}
	art, err := r.End(dir)
	if err != nil {
		t.Fatal(err)
	}
	return dir, art
}

func recordIntent(t *testing.T, r *run.Run, at int64) {
	t.Helper()
	if err := r.SetPlan(model.RunPlan{RequestedEndNS: model.DefaultStartTimeNS + 3600e9, Rate: 120}); err != nil {
		t.Fatal(err)
	}
	if err := r.SetTermination(model.Termination{Reason: model.TerminationStoppedByOperator, AtNS: at}); err != nil {
		t.Fatal(err)
	}
}

func TestAStoppedRunWritesItsIntentAndStillVerifiesByReplay(t *testing.T) {
	t.Parallel()
	dir, art := stoppedRun(t, 3, true)
	if len(art.CommandLog) != 3 {
		t.Fatalf("command log holds %d commands, want the 3 that ran", len(art.CommandLog))
	}
	loaded, err := run.LoadArtifact(filepath.Join(dir, "run.json"))
	if err != nil || loaded.Plan == nil || loaded.Plan.RequestedEndNS != model.DefaultStartTimeNS+3600e9 || loaded.Plan.Rate != 120 {
		t.Fatalf("loaded plan = %+v, %v", loaded.Plan, err)
	}
	if loaded.Termination == nil || loaded.Termination.Reason != model.TerminationStoppedByOperator || loaded.Termination.AtNS != model.DefaultStartTimeNS+1800e9 {
		t.Fatalf("loaded termination = %+v", loaded.Termination)
	}
	raw, err := os.ReadFile(filepath.Join(dir, "run.json"))
	if err != nil || model.ValidateRunArtifact(raw) != nil {
		t.Fatalf("run.json does not validate: %v", model.ValidateRunArtifact(raw))
	}
	res, err := run.ReplayFile(t.Context(), filepath.Join(dir, "run.json"), run.Directories{Domains: testsupport.DomainsDir(), Adapters: testsupport.AdaptersDir()})
	if err != nil || !res.Matches || res.GotDigest != art.ExpectedTraceDigest {
		t.Fatalf("replay = %+v, %v, want a match with %s", res, err, art.ExpectedTraceDigest)
	}
}

func TestARunThatRecordedNoIntentWritesNoPlanOrTermination(t *testing.T) {
	t.Parallel()
	dir, art := stoppedRun(t, 2, false)
	raw, err := os.ReadFile(filepath.Join(dir, "run.json"))
	if err != nil {
		t.Fatal(err)
	}
	if art.Plan != nil || art.Termination != nil || bytes.Contains(raw, []byte(`"plan"`)) || bytes.Contains(raw, []byte(`"termination"`)) {
		t.Fatalf("run.json carries intent nobody set: %s", raw)
	}
}

func TestTheArtifactSchemaRefusesIntentOutsideTheContract(t *testing.T) {
	t.Parallel()
	dir, _ := stoppedRun(t, 2, true)
	raw, err := os.ReadFile(filepath.Join(dir, "run.json"))
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	cases := map[string]func(){
		"an unknown reason":        func() { doc["termination"].(map[string]any)["reason"] = "walked_away" },
		"a termination without at": func() { delete(doc["termination"].(map[string]any), "at_ns") },
		"a plan without an end":    func() { delete(doc["plan"].(map[string]any), "requested_end_ns") },
		"a negative rate":          func() { doc["plan"].(map[string]any)["rate"] = -1 },
		"an extra plan member":     func() { doc["plan"].(map[string]any)["pace"] = 1 },
	}
	for name, mutate := range cases {
		mutate()
		bad, err := json.Marshal(doc)
		if err != nil {
			t.Fatal(err)
		}
		if model.ValidateRunArtifact(bad) == nil {
			t.Errorf("%s: the schema accepted it", name)
		}
		if err := json.Unmarshal(raw, &doc); err != nil {
			t.Fatal(err)
		}
	}
}
