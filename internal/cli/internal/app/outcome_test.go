package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

const shortPlantBScenario = `{"schema_version":"0.1","id":"short-b","domain":"belt-conveyor-drives-b","adapter":"native-jsonl","seed":1,
"world":{"start_time_ns":1767225600000000000},"commands":[
{"op":"clock.advance","args":{"to_ns":1767225660000000000}},
{"op":"fault.inject","args":{"entity_id":"D-3","fault":"conveyor_drive__vfd_fault","onset_ns":1767225660000000000,"params":null}},
{"op":"clock.advance","args":{"to_ns":1767225670000000000}},
{"op":"effector.invoke","args":{"effector":"bulk_feeder__set_feed","entity_id":"F-1","command_id":"c-1","args":{"percent":0},"at_ns":1767225670000000000}},
{"op":"clock.advance","args":{"to_ns":1767226200000000000}}]}`

func examplesDir() string { return filepath.Join(testsupport.RepositoryRoot(), "examples") }

func shortPlantBRun(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	testsupport.WriteFileIn(t, dir, "short.scenario.json", []byte(shortPlantBScenario))
	out := filepath.Join(dir, "run")
	invoke(t, "run", "--scenario", filepath.Join(dir, "short.scenario.json"), "--domains-dir", examplesDir(),
		"--adapters-dir", testsupport.AdaptersDir(), "--kinds-dir", testsupport.BeltConveyorKindsDir(), "--sink", "file", "--out", out).mustSucceed(t)
	return out
}

func TestOutcomeScoresARunAgainstDoingNothingAndEachBaselinePolicy(t *testing.T) {
	t.Parallel()
	doc := invoke(t, "outcome", "--run", shortPlantBRun(t), "--domains-dir", examplesDir(), "--adapters-dir", testsupport.AdaptersDir()).
		mustSucceed(t).document(t, "outcome")
	runs, _ := doc["runs"].([]any)
	verdicts, _ := doc["verdicts"].([]any)
	if len(runs) != 6 || len(verdicts) != 5 || doc["valid"] != true || doc["tie"] != 0.1 || doc["domain"] != "belt-conveyor-drives-b" {
		t.Fatalf("outcome = %v, want as_run, do_nothing and b1-b4, a valid story and five verdicts", doc)
	}
	first, _ := verdicts[0].(map[string]any)
	if first["baseline"] != "do_nothing" || first["verdict"] != "beats" {
		t.Fatalf("first verdict %v, want cutting the feed to beat doing nothing on a weak plant", first)
	}
}

func TestReportCarriesTheOutcomeScoreOnlyWhenAskedFor(t *testing.T) {
	t.Parallel()
	dir := shortPlantBRun(t)
	plain := invoke(t, "report", dir).mustSucceed(t).json(t)
	asked := invoke(t, "report", dir, "--outcome").mustSucceed(t).json(t)
	scored, _ := asked["outcome_score"].(map[string]any)
	if plain["outcome_score"] != nil || scored["status"] != "scored" || scored["result"] == nil {
		t.Fatalf("outcome score %v by default and %v when asked, want none and a scored result", plain["outcome_score"], asked["outcome_score"])
	}
	if got := invoke(t, "report", "compare", "--a", dir, "--b", dir, "--outcome").failure(t); got["code"] != "usage_invalid" {
		t.Fatalf("compare with --outcome = %v, want usage_invalid", got)
	}
}

func TestOutcomeOfADomainWithoutWeightsReportsFactsOnly(t *testing.T) {
	t.Parallel()
	doc := invoke(t, "outcome", "--run", runPondDocument(t), "--domains-dir", testsupport.DomainsDir(), "--adapters-dir", testsupport.AdaptersDir()).
		mustSucceed(t).document(t, "outcome")
	runs, _ := doc["runs"].([]any)
	asRun, _ := runs[0].(map[string]any)
	if len(runs) != 2 || doc["verdicts"] != nil || doc["tie"] != nil || asRun["cost"] != nil || asRun["facts"] == nil {
		t.Fatalf("outcome = %v, want two runs with facts and no cost, tie or verdict", doc)
	}
}

func TestOutcomeFailuresNameWhatIsWrong(t *testing.T) {
	t.Parallel()
	pond := runPondDocument(t)
	editEmbeddedSpec(t, filepath.Join(pond, "run.json"))
	foreign, otherDomains := runPondDocument(t), t.TempDir()
	withoutSpec(t, filepath.Join(foreign, "run.json"))
	copyDomain(t, testsupport.BeltConveyorDomain("a"), otherDomains)
	cases := []struct {
		name string
		args []string
		code string
	}{
		{"no run", []string{"outcome"}, "usage_invalid"},
		{"a missing run folder", []string{"outcome", "--run", filepath.Join(t.TempDir(), "none")}, "file_not_found"},
		{"a run whose domain is unknown", []string{"outcome", "--run", foreign, "--domains-dir", otherDomains, "--adapters-dir", testsupport.AdaptersDir()}, "domain_not_found"},
		{"a run that no longer replays", []string{"outcome", "--run", pond}, "replay_mismatch"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := invoke(t, tc.args...).failure(t); got["code"] != tc.code {
				t.Fatalf("failure %v, want %s", got, tc.code)
			}
		})
	}
}

func withoutSpec(t *testing.T, artifact string) {
	t.Helper()
	raw, err := os.ReadFile(artifact)
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	delete(doc, "domain_spec")
	edited, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	place(t, artifact, edited)
}

func copyDomain(t *testing.T, path, dir string) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	testsupport.WriteFileIn(t, dir, filepath.Base(path), raw)
}
