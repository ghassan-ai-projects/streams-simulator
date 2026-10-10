package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

// runPond executes a short scripted run and returns its output directory.
func runPond(t *testing.T, extra ...string) string {
	t.Helper()
	dir := t.TempDir()
	args := append([]string{"run", "--domains-dir", testsupport.DomainsDir(), "--adapters-dir", testsupport.AdaptersDir(), "--domain", pondDomain,
		"--seed", "7", "--duration", "900", "--sink", "file", "--out", dir}, extra...)
	got := invoke(t, args...).mustSucceed(t).json(t)
	if got["run_id"] == "" || got["reproducible"] != true {
		t.Fatalf("run result = %v", got)
	}
	return dir
}

func TestRunPublishesEvidenceThatReplayAndVerifyReproduce(t *testing.T) {
	t.Parallel()
	dir := runPond(t, "--perturb", "drop@1@2")
	for _, name := range []string{"run.json", "trace.jsonl", "ledger.jsonl"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Fatalf("run evidence %s: %v", name, err)
		}
	}
	for _, command := range []string{"replay", "verify"} {
		got := invoke(t, command, "--domains-dir", testsupport.DomainsDir(), "--adapters-dir", testsupport.AdaptersDir(), filepath.Join(dir, "run.json")).mustSucceed(t).json(t)
		if got["matches"] != true || got["version_match"] != true {
			t.Fatalf("%s result = %v", command, got)
		}
	}
}

func TestReplayNamesAnArtifactThatCannotBeRead(t *testing.T) {
	t.Parallel()
	got := invoke(t, "replay", filepath.Join(t.TempDir(), "absent.json"))
	if got.code != 1 || !strings.Contains(got.stderr, "run: read artifact") {
		t.Fatalf("replay of a missing artifact: %+v", got)
	}
}

func TestRunRefusesAnUnknownAdapterAndANonexistentEffector(t *testing.T) {
	t.Parallel()
	base := []string{"run", "--domains-dir", testsupport.DomainsDir(), "--adapters-dir", testsupport.AdaptersDir(), "--domain", pondDomain, "--duration", "60"}
	unknown := invoke(t, append(base, "--adapter", "no-such")...)
	if unknown.code != 1 || !strings.Contains(unknown.stderr, `unknown adapter "no-such"`) {
		t.Fatalf("unknown adapter: %+v", unknown)
	}
	effector := invoke(t, append(base, "--effector", "no_such@no-entity@1")...)
	if effector.code != 1 || !strings.Contains(effector.stderr, "effector no_such:") {
		t.Fatalf("unknown effector: %+v", effector)
	}
}

func TestRefconsumerConsumesATraceAndScoreReadsItsVerdict(t *testing.T) {
	t.Parallel()
	dir := runPond(t)
	verdict := filepath.Join(dir, "verdict.json")
	consumed := invoke(t, "refconsumer", "--trace", filepath.Join(dir, "trace.jsonl"), "--out", verdict).mustSucceed(t).json(t)
	if consumed["verdict_written"] != verdict {
		t.Fatalf("refconsumer result = %v", consumed)
	}
	label := filepath.Join(dir, "label.json")
	record := `{"scenario_id":"s-1","domain":"aquaculture-pond","seed":7,"entity_id":"","label":"nominal",` +
		`"expected_episode":false,"injection_time_ns":0,"first_observable_time_ns":0,"unavoidable_time_ns":0,` +
		`"observability":{},"trivial_baseline_verdict":"non_trivial"}`
	if err := os.WriteFile(label, []byte(record), 0o600); err != nil {
		t.Fatal(err)
	}
	scored := invoke(t, "score", "--run", filepath.Join(dir, "run.json"), "--label", label).mustSucceed(t).json(t)
	if len(scored) == 0 {
		t.Fatal("score printed an empty scorecard")
	}
}

func TestRefconsumerAndScoreNameWhatTheyCannotRead(t *testing.T) {
	t.Parallel()
	missing := filepath.Join(t.TempDir(), "absent")
	cases := []struct {
		name string
		args []string
		want string
	}{
		{"missing trace", []string{"refconsumer", "--trace", missing}, "no such file or directory"},
		{"effector without endpoint", []string{"refconsumer", "--trace", missing, "--effector", "x"}, "--effector requires --mcp"},
		{"missing verdict", []string{"score", "--run", filepath.Join(missing, "run.json")}, "read " + filepath.Join(missing, "verdict.json")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := invoke(t, tc.args...)
			if got.code != 1 || !strings.Contains(got.stderr, tc.want) {
				t.Fatalf("%v: %+v, want %q", tc.args, got, tc.want)
			}
		})
	}
}
