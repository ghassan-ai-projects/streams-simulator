package app

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func runScenarioFile(t *testing.T, path string) outcome {
	t.Helper()
	return invoke(t, "run", "--domains-dir", testsupport.DomainsDir(), "--adapters-dir", testsupport.AdaptersDir(), "--scenario", path)
}

const pondScenarioHead = `{"schema_version":"0.1","id":"s","domain":"aquaculture-pond","adapter":"native-jsonl",`

func TestAScenarioFileThatCannotBeHeldOrTimedIsScenarioInvalid(t *testing.T) {
	t.Parallel()
	cases := []struct{ name, content, pointer string }{
		{"seed beyond 64 bits", pondScenarioHead + `"seed":1e30,"world":{"start_time_ns":1},"commands":[]}`, "/seed"},
		{"time with an exponent", pondScenarioHead + `"seed":1,"world":{"start_time_ns":1},` +
			`"commands":[{"op":"clock.advance","args":{"to_ns":1e3}}]}`, "/commands/0/args/to_ns"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			path := writeFile(t, "s.json", tc.content)
			got := runScenarioFile(t, path)
			failure := got.failure(t)
			issues, _ := failure["issues"].([]any)
			if got.code != 1 || failure["code"] != "scenario_invalid" || failure["source"] != path || len(issues) != 1 ||
				issues[0].(map[string]any)["pointer"] != tc.pointer {
				t.Fatalf("exit %d failure %v; want scenario_invalid at %s", got.code, failure, tc.pointer)
			}
		})
	}
}

func TestAMissingScenarioFileIsNamed(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "absent.json")
	got := runScenarioFile(t, path)
	if got.code != 1 || got.failure(t)["code"] != "file_not_found" || got.failure(t)["source"] != path {
		t.Fatalf("exit %d failure %v; want file_not_found naming %s", got.code, got.failure(t), path)
	}
}

func TestScenarioExportFailuresNameWhatFailed(t *testing.T) {
	t.Parallel()
	absent := filepath.Join(t.TempDir(), "absent.json")
	missing := invoke(t, "scenario", "from-run", "--domains-dir", testsupport.DomainsDir(), "--adapters-dir", testsupport.AdaptersDir(), absent, "--out", filepath.Join(t.TempDir(), "s.json"))
	if missing.code != 1 || missing.failure(t)["code"] != "file_not_found" {
		t.Fatalf("export of a missing artifact: exit %d failure %v", missing.code, missing.failure(t))
	}
	blocker := writeFile(t, "not-a-dir", "x")
	unwritable := invoke(t, "scenario", "from-run", "--domains-dir", testsupport.DomainsDir(), "--adapters-dir", testsupport.AdaptersDir(), filepath.Join(scriptedPondRun(t), "run.json"), "--out", filepath.Join(blocker, "s.json"))
	if unwritable.code != 1 || unwritable.failure(t)["code"] != "io_failed" {
		t.Fatalf("export to an unwritable path: exit %d failure %v", unwritable.code, unwritable.failure(t))
	}
}

// An artifact whose command log holds a command no scenario can express
// exports nothing: the export checks its own contract before writing.
func TestAnExportThatWouldBreakTheScenarioContractWritesNothing(t *testing.T) {
	t.Parallel()
	dir := scriptedPondRun(t)
	artifact := filepath.Join(dir, "run.json")
	raw, err := os.ReadFile(artifact)
	if err != nil {
		t.Fatal(err)
	}
	tampered := strings.Replace(string(raw), `"perturbation": "drop"`, `"perturbation": ""`, 1)
	if tampered == string(raw) {
		t.Fatal("the run's artifact has no drop perturbation to tamper with")
	}
	artifact = writeFile(t, "run.json", tampered)
	out := filepath.Join(t.TempDir(), "s.json")
	got := invoke(t, "scenario", "from-run", "--domains-dir", testsupport.DomainsDir(), "--adapters-dir", testsupport.AdaptersDir(), artifact, "--out", out)
	if got.code != 1 || got.failure(t)["code"] != "scenario_invalid" || !strings.Contains(got.stderr, "exported scenario") ||
		got.failure(t)["source"] != artifact {
		t.Fatalf("exit %d failure %v; want scenario_invalid for the export", got.code, got.failure(t))
	}
	if _, err := os.Stat(out); !os.IsNotExist(err) {
		t.Fatalf("a scenario that breaks its contract was written: %v", err)
	}
}

// An artifact whose commands satisfy the scenario contract but not the
// domain exports nothing: the export is checked as run --scenario would
// check it, and the violation names the artifact.
func TestAnExportTheDomainRefusesWritesNothing(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile(filepath.Join(scriptedPondRun(t), "run.json"))
	if err != nil {
		t.Fatal(err)
	}
	tampered := strings.Replace(string(raw), `"fault": "aerator_failure"`, `"fault": "no_such_fault"`, 1)
	if tampered == string(raw) {
		t.Fatal("the run's artifact has no aerator_failure fault to tamper with")
	}
	artifact := writeFile(t, "run.json", tampered)
	out := filepath.Join(t.TempDir(), "s.json")
	got := invoke(t, "scenario", "from-run", "--domains-dir", testsupport.DomainsDir(), "--adapters-dir", testsupport.AdaptersDir(), artifact, "--out", out)
	failure := got.failure(t)
	issues, _ := failure["issues"].([]any)
	if got.code != 1 || failure["code"] != "scenario_invalid" || failure["source"] != artifact || len(issues) != 1 ||
		issues[0].(map[string]any)["pointer"] != "/commands/0/args/fault" {
		t.Fatalf("exit %d failure %v; want scenario_invalid from the artifact at the fault", got.code, failure)
	}
	if _, err := os.Stat(out); !os.IsNotExist(err) {
		t.Fatalf("a scenario the domain refuses was written: %v", err)
	}
}

// tamperedPondArtifact is the scripted pond run's artifact with one text
// replaced, written to a new file.
func tamperedPondArtifact(t *testing.T, old, replacement string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(scriptedPondRun(t), "run.json"))
	if err != nil {
		t.Fatal(err)
	}
	tampered := strings.Replace(string(raw), old, replacement, 1)
	if tampered == string(raw) {
		t.Fatalf("the run's artifact has no %s to tamper with", old)
	}
	return writeFile(t, "run.json", tampered)
}

// A consumer call the world refused at admission is dropped from the export
// and counted, not refused: it changed nothing.
func TestARefusedCallIsDroppedFromTheExportAndCounted(t *testing.T) {
	t.Parallel()
	artifact := tamperedPondArtifact(t, `"command_id": "cli-0"`, `"command_id": ""`)
	out := filepath.Join(t.TempDir(), "s.json")
	exported := invoke(t, "scenario", "from-run", "--domains-dir", testsupport.DomainsDir(), "--adapters-dir", testsupport.AdaptersDir(),
		artifact, "--out", out).mustSucceed(t).document(t, "scenario_export")
	if exported["refused"] != float64(1) || exported["commands"] != float64(3) {
		t.Fatalf("export = %v, want the refused call dropped and counted", exported)
	}
}

// An export is checked against the domain the run recorded; a changed domain
// is a replay mismatch, not a silent export against other rules.
func TestAnExportAgainstAChangedDomainIsAMismatch(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile(filepath.Join(scriptedPondRun(t), "run.json"))
	if err != nil {
		t.Fatal(err)
	}
	first := regexp.MustCompile(`"digest": "sha256:[0-9a-f]`).FindIndex(raw)
	if first == nil {
		t.Fatal("the run's artifact records no domain digest")
	}
	changed := []byte(string(raw))
	changed[first[1]-1] = map[bool]byte{true: '1', false: '0'}[raw[first[1]-1] == '0']
	artifact := writeFile(t, "run.json", string(changed))
	got := invoke(t, "scenario", "from-run", "--domains-dir", testsupport.DomainsDir(), "--adapters-dir", testsupport.AdaptersDir(),
		artifact, "--out", filepath.Join(t.TempDir(), "s.json"))
	if got.code != 1 || got.failure(t)["code"] != "replay_mismatch" {
		t.Fatalf("exit %d failure %v; want replay_mismatch", got.code, got.failure(t))
	}
}
