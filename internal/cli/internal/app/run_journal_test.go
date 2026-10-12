package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

// runInto runs the pond domain into a fresh directory and returns it.
func runInto(t *testing.T, extra ...string) string {
	t.Helper()
	out := t.TempDir()
	args := append([]string{"run", "--domain", pondDomain, "--seed", "5", "--duration", "600", "--sink", "file", "--out", out,
		"--domains-dir", testsupport.DomainsDir(), "--adapters-dir", testsupport.AdaptersDir()}, extra...)
	invoke(t, args...).mustSucceed(t)
	return out
}

func TestRunWithAnOutDirectoryWritesBothJournalStreams(t *testing.T) {
	t.Parallel()
	out := runInto(t)
	for _, name := range []string{"journal.jsonl", filepath.Join("sealed", "journal.jsonl")} {
		if raw, err := os.ReadFile(filepath.Join(out, name)); err != nil || !strings.Contains(string(raw), `"type":"run.closed"`) {
			t.Fatalf("%s = %q (%v), want a closed journal", name, raw, err)
		}
	}
}

func TestVerifyFailsNamingTheJournalStreamThatChanged(t *testing.T) {
	t.Parallel()
	out := runInto(t)
	path := filepath.Join(out, "journal.jsonl")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	// #nosec G703 -- the path is a file in the test's own temporary run directory
	if err := os.WriteFile(path, []byte(strings.Replace(string(raw), `"mode":"evaluate"`, `"mode":"explore"`, 1)), 0o600); err != nil {
		t.Fatal(err)
	}
	got := invoke(t, "verify", filepath.Join(out, "run.json"), "--domains-dir", testsupport.DomainsDir(), "--adapters-dir", testsupport.AdaptersDir())
	failure := got.json(t)["error"].(map[string]any)
	if got.code != 1 || failure["code"] != "replay_mismatch" || !strings.Contains(failure["message"].(string), "journal.jsonl differs from the replay at record 0") {
		t.Fatalf("exit %d, %v; want replay_mismatch naming journal.jsonl", got.code, failure)
	}
}

func TestExploreRunIsStampedUnblindedAndScoreRefusesIt(t *testing.T) {
	t.Parallel()
	out := runInto(t, "--explore")
	artifact := filepath.Join(out, "run.json")
	raw, err := os.ReadFile(artifact)
	if err != nil || !strings.Contains(string(raw), `"mode": "explore"`) || !strings.Contains(string(raw), `"unblinded": true`) {
		t.Fatalf("run.json = %s (%v), want an unblinded explore run", raw, err)
	}
	got := invoke(t, "score", "--run", artifact, "--domains-dir", testsupport.DomainsDir(), "--adapters-dir", testsupport.AdaptersDir())
	failure := got.json(t)["error"].(map[string]any)
	if got.code != 1 || failure["code"] != "score_failed" || !strings.Contains(failure["message"].(string), "unblinded") {
		t.Fatalf("score of an explore run: exit %d, %v", got.code, failure)
	}
}
