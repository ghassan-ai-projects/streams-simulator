package acceptance

// The run journal through the real binary (docs/design/RUN_JOURNAL.md): what
// an Evaluate run leaves where a consumer can read it, and what Explore mode
// does to scoring.

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEvaluateRunLeavesNoInjectedTruthWhereAConsumerCanRead(t *testing.T) {
	t.Parallel()
	b := newBinary(t)
	dir := b.runPond("--fault", "site-a/pond-1=aerator_failure@60", "--perturb", "duplicate_burst@1@2")
	secrets := []string{"aerator_failure", "duplicate_burst", `"seed"`}
	for _, name := range []string{"journal.jsonl", "trace.jsonl"} {
		raw := readRunFile(t, dir, name)
		for _, secret := range secrets {
			if bytes.Contains(raw, []byte(secret)) {
				t.Errorf("%s names %q before the run is sealed", name, secret)
			}
		}
	}
	sealed := readRunFile(t, dir, filepath.Join("sealed", "journal.jsonl"))
	for _, secret := range secrets {
		if !bytes.Contains(sealed, []byte(secret)) {
			t.Errorf("the director stream must name %q", secret)
		}
	}
	verified := b.run(append(append([]string{"verify"}, inputs()...), filepath.Join(dir, "run.json"))...).mustSucceed(t).document(t)
	if verified["matches"] != true {
		t.Fatalf("verify = %v, want both journals reproduced", verified)
	}
}

func TestExploreRunIsRefusedByScore(t *testing.T) {
	t.Parallel()
	b := newBinary(t)
	dir := b.runPond("--explore")
	raw := readRunFile(t, dir, "run.json")
	if !strings.Contains(string(raw), `"mode": "explore"`) || !strings.Contains(string(raw), `"unblinded": true`) {
		t.Fatalf("run.json = %s, want an unblinded explore run", raw)
	}
	label := filepath.Join(dir, "label.json")
	if err := os.WriteFile(label, []byte(nominalPondLabel), 0o600); err != nil {
		t.Fatal(err)
	}
	got := b.run(append([]string{"score", "--run", filepath.Join(dir, "run.json"), "--label", label}, inputs()...)...)
	if got.code != 1 || !strings.Contains(got.stdout, "score_failed") || !strings.Contains(got.stdout, "unblinded") {
		t.Fatalf("score of an explore run: exit %d\nstdout: %s", got.code, got.stdout)
	}
}

func readRunFile(t *testing.T, dir, name string) []byte {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestClearingTheUnblindedFlagOfAnExploreRunDoesNotMakeItScorable(t *testing.T) {
	t.Parallel()
	b := newBinary(t)
	dir := b.runPond("--explore")
	artifact := filepath.Join(dir, "run.json")
	raw := readRunFile(t, dir, "run.json")
	flipped := strings.Replace(string(raw), `"unblinded": true`, `"unblinded": false`, 1)
	if flipped == string(raw) {
		t.Fatal("run.json has no unblinded flag to clear")
	}
	if err := os.WriteFile(artifact, []byte(flipped), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := b.run(append([]string{"verify", artifact}, inputs()...)...); got.code != 1 || !strings.Contains(got.stdout, "replay_mismatch") {
		t.Fatalf("verify: exit %d\nstdout: %s", got.code, got.stdout)
	}
	label := filepath.Join(dir, "label.json")
	if err := os.WriteFile(label, []byte(nominalPondLabel), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := b.run(append([]string{"score", "--run", artifact, "--label", label}, inputs()...)...); got.code != 1 || !strings.Contains(got.stdout, "unblinded") {
		t.Fatalf("score: exit %d\nstdout: %s", got.code, got.stdout)
	}
}
