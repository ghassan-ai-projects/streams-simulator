package run_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func publishedRun(t *testing.T) (string, run.Directories) {
	t.Helper()
	dirs := run.Directories{Domains: testsupport.DomainsDir(), Adapters: testsupport.AdaptersDir()}
	dir := t.TempDir()
	if _, err := run.RunScenario(t.Context(), scenarioFor(model.DefaultStartTimeNS), run.ScenarioRunOptions{DomainsDir: dirs.Domains,
		AdaptersDir: dirs.Adapters, SinkName: model.SinkInproc, OutDir: dir}); err != nil {
		t.Fatal(err)
	}
	return dir, dirs
}

func rewrite(t *testing.T, path string, change func([]byte) []byte) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, change(raw), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestVerifyCatchesATamperedLedgerOrHistory(t *testing.T) {
	t.Parallel()
	for file, change := range map[string]func([]byte) []byte{
		model.FileLedger: func(raw []byte) []byte {
			return bytes.Replace(raw, []byte(`"delivered":true`), []byte(`"delivered":false`), 1)
		},
		model.FileHistory: func(raw []byte) []byte { return raw[:bytes.LastIndexByte(raw[:len(raw)-1], '\n')+1] },
	} {
		dir, dirs := publishedRun(t)
		artifact := filepath.Join(dir, model.FileRunArtifact)
		if res, err := run.ReplayFile(t.Context(), artifact, dirs); err != nil || !res.Matches {
			t.Fatalf("%s untouched: %+v, %v", file, res, err)
		}
		rewrite(t, filepath.Join(dir, file), change)
		res, err := run.ReplayFile(t.Context(), artifact, dirs)
		if err != nil || res.Matches || !strings.Contains(res.Detail, file) {
			t.Fatalf("%s tampered: %+v, %v", file, res, err)
		}
	}
}

func TestARunFolderUsedTwiceKeepsOneLedger(t *testing.T) {
	t.Parallel()
	spec, err := domain.Load(testsupport.Domain("aquaculture-pond"))
	if err != nil {
		t.Fatal(err)
	}
	a, err := adapter.Load(testsupport.Adapter("native-jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	path := filepath.Join(dir, model.FileLedger)
	var sizes []int
	for range 2 {
		r, err := run.New(t.Context(), run.Config{Domain: spec, Adapter: a, Seed: 4, SinkName: model.SinkInproc, TimeMode: model.TimeStepped,
			StartTimeNS: model.DefaultStartTimeNS, LedgerPath: path})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := r.Advance(t.Context(), model.DefaultStartTimeNS+300e9, false); err != nil {
			t.Fatal(err)
		}
		if _, err := r.End(dir); err != nil {
			t.Fatal(err)
		}
		raw, _ := os.ReadFile(path)
		sizes = append(sizes, len(raw))
	}
	if sizes[0] == 0 || sizes[0] != sizes[1] {
		t.Fatalf("ledger sizes %v: the second run must replace the first", sizes)
	}
}
