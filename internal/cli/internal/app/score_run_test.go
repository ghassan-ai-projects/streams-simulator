package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestScoreOfARunWithoutDisturbancesPrintsNoScorecardAndWhatTheConsumerDidAnyway(t *testing.T) {
	t.Parallel()
	dir := runPondDocument(t)
	invoke(t, "refconsumer", "--trace", filepath.Join(dir, "trace.jsonl"), "--out", filepath.Join(dir, "verdict.json")).mustSucceed(t)
	if err := os.WriteFile(filepath.Join(dir, "labels.json"), []byte("[]\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	got := invoke(t, "score", "--run", filepath.Join(dir, "run.json"),
		"--domains-dir", testsupport.DomainsDir(), "--adapters-dir", testsupport.AdaptersDir()).mustSucceed(t).document(t, "score_run")
	if scorecards, _ := got["scorecards"].([]any); len(scorecards) != 0 || got["run_id"] == "" {
		t.Fatalf("score document = %v, want the run id and no scorecard", got)
	}
	if _, ok := got["unexpected"].(map[string]any); !ok {
		t.Fatalf("score document = %v, want an unexpected section", got)
	}
}
