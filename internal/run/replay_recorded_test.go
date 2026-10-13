package run_test

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func place(t *testing.T, path string, content []byte) {
	t.Helper()
	// #nosec G703 -- the path is a file in the test's own temporary directory
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}
}

func runInEditableDomains(t *testing.T) (artifact, domainFile string, dirs run.Directories) {
	t.Helper()
	domainsDir := t.TempDir()
	domainFile = filepath.Join(domainsDir, "aquaculture-pond.domain.json")
	raw, err := os.ReadFile(testsupport.Domain("aquaculture-pond"))
	if err != nil {
		t.Fatal(err)
	}
	place(t, domainFile, raw)
	dirs = run.Directories{Domains: domainsDir, Adapters: testsupport.AdaptersDir()}
	out := t.TempDir()
	if _, err := run.RunScenario(t.Context(), scenarioFor(model.DefaultStartTimeNS), run.ScenarioRunOptions{DomainsDir: dirs.Domains,
		AdaptersDir: dirs.Adapters, SinkName: model.SinkInproc, OutDir: out}); err != nil {
		t.Fatal(err)
	}
	return filepath.Join(out, model.FileRunArtifact), domainFile, dirs
}

func TestReplayFileReplaysTheRecordedDomainAndReportsAnEditedOneAsInformation(t *testing.T) {
	t.Parallel()
	edits := map[string]func([]byte) []byte{
		"a description only": func(raw []byte) []byte {
			return bytes.Replace(raw, []byte("Eight shrimp ponds."), []byte("Eight ponds."), 1)
		},
		"a dynamics change": func(raw []byte) []byte {
			return bytes.Replace(raw, []byte(`"initial": 6.5,`), []byte(`"initial": 6.4,`), 1)
		},
	}
	for name, edit := range edits {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			artifact, domainFile, dirs := runInEditableDomains(t)
			before, err := run.ReplayFile(t.Context(), artifact, dirs)
			if err != nil || !before.Matches || before.DomainChanged || before.CurrentDomainDigest != "" {
				t.Fatalf("untouched: %+v, %v", before, err)
			}
			rewrite(t, domainFile, edit)
			after, err := run.ReplayFile(t.Context(), artifact, dirs)
			if err != nil || !after.Matches || !after.DomainChanged || after.CurrentDomainDigest == "" || after.GotDigest != before.GotDigest {
				t.Fatalf("edited: %+v, %v", after, err)
			}
		})
	}
}

func TestReplayFileStillFailsATamperedArtifactAndNeedsNoDomainsDirectory(t *testing.T) {
	t.Parallel()
	artifact, _, dirs := runInEditableDomains(t)
	alone, err := run.ReplayFile(t.Context(), artifact, run.Directories{Domains: filepath.Join(t.TempDir(), "absent")})
	if err != nil || !alone.Matches || alone.DomainChanged {
		t.Fatalf("no domains directory: %+v, %v", alone, err)
	}
	rewrite(t, artifact, func(raw []byte) []byte {
		return regexp.MustCompile(`("expected_trace_digest": "sha256:)[0-9a-f]{64}`).ReplaceAll(raw, []byte("${1}"+strings.Repeat("0", 64)))
	})
	res, err := run.ReplayFile(t.Context(), artifact, dirs)
	if err != nil || res.Matches {
		t.Fatalf("tampered expectation: %+v, %v", res, err)
	}
}
