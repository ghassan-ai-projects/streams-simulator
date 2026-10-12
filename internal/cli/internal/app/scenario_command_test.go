package app

import (
	"encoding/json"
	"flag"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

// scriptedPondRun runs the pond from flags with a fault, a perturbation and an
// effector call, and returns its output directory.
func scriptedPondRun(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	call := `{"pond_id":"site-a/pond-1","level":1}`
	invoke(t, "run", "--domains-dir", testsupport.DomainsDir(), "--adapters-dir", testsupport.AdaptersDir(), "--domain", pondDomain,
		"--seed", "9", "--duration", "1800", "--sink", "file", "--out", dir, "--fault", "site-a/pond-1=aerator_failure@600",
		"--perturb", "drop@60@120", "--effector", "start_aerator@site-a/pond-1@30@"+call).mustSucceed(t).document(t, "run")
	return dir
}

// artifactIdentity is what must match between two runs that are the same run.
func artifactIdentity(t *testing.T, dir string) map[string]any {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(dir, "run.json"))
	if err != nil {
		t.Fatal(err)
	}
	var artifact map[string]any
	if err := json.Unmarshal(raw, &artifact); err != nil {
		t.Fatal(err)
	}
	return map[string]any{"trace": artifact["expected_trace_digest"], "world": artifact["world_digest"],
		"commands": artifact["command_log"], "config": artifact["world_config"]}
}

func TestAnExportedScenarioRunsTheSameRun(t *testing.T) {
	t.Parallel()
	original := scriptedPondRun(t)
	scenarioFile := filepath.Join(t.TempDir(), "s.json")
	exported := invoke(t, "scenario", "from-run", "--domains-dir", testsupport.DomainsDir(), "--adapters-dir", testsupport.AdaptersDir(), filepath.Join(original, "run.json"), "--out", scenarioFile).mustSucceed(t).document(t, "scenario_export")
	if exported["commands"] != float64(4) || exported["omitted"] != float64(0) {
		t.Fatalf("export = %v, want 4 commands and none omitted", exported)
	}
	again := t.TempDir()
	invoke(t, "run", "--domains-dir", testsupport.DomainsDir(), "--adapters-dir", testsupport.AdaptersDir(),
		"--scenario", scenarioFile, "--sink", "file", "--out", again).mustSucceed(t).document(t, "run")
	got, want := artifactIdentity(t, again), artifactIdentity(t, original)
	gotRaw, _ := json.Marshal(got)
	wantRaw, _ := json.Marshal(want)
	if string(gotRaw) != string(wantRaw) {
		t.Fatalf("the scenario run differs from the original:\n got %s\nwant %s", gotRaw, wantRaw)
	}
	invoke(t, "verify", "--domains-dir", testsupport.DomainsDir(), "--adapters-dir", testsupport.AdaptersDir(),
		filepath.Join(again, "run.json")).mustSucceed(t).document(t, "replay")
}

func TestAScenarioFileFixesTheFlagsItDefines(t *testing.T) {
	t.Parallel()
	got := invoke(t, "run", "--scenario", "s.json", "--domain", pondDomain, "--seed", "3")
	if got.code != 2 || got.failure(t)["code"] != "usage_invalid" {
		t.Fatalf("exit %d failure %v; want usage_invalid naming the conflicting flags", got.code, got.failure(t))
	}
	if message := got.failure(t)["message"]; message != "run: --scenario defines the run; remove --domain, --seed" {
		t.Fatalf("message = %v", message)
	}
}

// Every run flag is either defined by a scenario, and refused beside
// --scenario, or an execution choice, and accepted with it.
func TestEveryRunFlagIsEitherDefinedByTheScenarioOrAnExecutionChoice(t *testing.T) {
	t.Parallel()
	defined := map[string]bool{"domain": true, "adapter": true, "seed": true, "profile": true, "start-time": true,
		"duration": true, "fault": true, "perturb": true, "effector": true}
	all := newFlagSet("run", io.Discard)
	registerScenarioFlags(all, &runOptions{})
	registerExecutionFlags(all, &runOptions{})
	all.VisitAll(func(registered *flag.Flag) {
		value := registered.DefValue
		if value == "" || registered.Name == "sink" {
			value = "file"
		}
		_, err := parseRunOptions([]string{"--scenario", "s.json", "--out", "o", "--" + registered.Name, value}, io.Discard)
		refused := err != nil && strings.Contains(err.Error(), "--scenario defines the run; remove --"+registered.Name)
		if refused != defined[registered.Name] {
			t.Errorf("--%s beside --scenario: %v; defined by the scenario = %v", registered.Name, err, defined[registered.Name])
		}
	})
	if len(scenarioDefinedFlags()) != len(defined) {
		t.Fatalf("scenario-defined flags = %v, want %v", scenarioDefinedFlags(), defined)
	}
}

func TestAScenarioFileThatBreaksItsContractOrDomainIsScenarioInvalid(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, content, pointer string
	}{
		{"schema", `{"schema_version":"0.1"}`, ""},
		{"domain", `{"schema_version":"0.1","id":"bad-fault","domain":"aquaculture-pond","adapter":"native-jsonl","seed":1,` +
			`"world":{"start_time_ns":0},"commands":[{"op":"fault.inject","args":{"entity_id":"site-a/pond-1",` +
			`"fault":"no_such_fault","onset_ns":0,"params":null}}]}`, "/commands/0/args/fault"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			path := writeFile(t, tc.name+".scenario.json", tc.content)
			got := invoke(t, "run", "--domains-dir", testsupport.DomainsDir(), "--adapters-dir", testsupport.AdaptersDir(), "--scenario", path)
			failure := got.failure(t)
			issues, _ := failure["issues"].([]any)
			first, _ := issues[0].(map[string]any)
			if got.code != 1 || failure["code"] != "scenario_invalid" || failure["source"] != path || first["pointer"] != tc.pointer {
				t.Fatalf("exit %d failure %v; want scenario_invalid from %s with an issue at %q", got.code, failure, path, tc.pointer)
			}
		})
	}
}

func TestScenarioExportNeedsAnArtifactAndAnOutput(t *testing.T) {
	t.Parallel()
	for _, args := range [][]string{{"scenario"}, {"scenario", "from-run"}, {"scenario", "from-run", "run.json"}, {"scenario", "export", "run.json", "--out", "s.json"}} {
		if got := invoke(t, args...); got.code != 2 || got.failure(t)["code"] != "usage_invalid" {
			t.Fatalf("%v: exit %d, want usage_invalid", args, got.code)
		}
	}
}
