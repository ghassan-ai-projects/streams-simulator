package app

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/journal"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestARunSealsItsKPITruthIntoTheDirectorJournalOnceBeforeItCloses(t *testing.T) {
	t.Parallel()
	spec, err := domain.Load(filepath.Join(testsupport.RepositoryRoot(), "examples", "factory-utilities.domain.json"))
	if err != nil {
		t.Fatal(err)
	}
	native, err := adapter.Load(testsupport.Adapter("native-jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	dir, start := t.TempDir(), model.DefaultStartTimeNS
	r, err := New(context.Background(), Config{Domain: spec, Adapter: native, Seed: 5, SinkName: model.SinkInproc,
		TimeMode: model.TimeStepped, StartTimeNS: start, JournalDir: dir})
	if err != nil {
		t.Fatal(err)
	}
	advanceAndReport(t, r, start+600e9)
	if _, err := r.End(""); err != nil {
		t.Fatal(err)
	}
	director, err := journal.Read(readFile(t, filepath.Join(dir, journal.DirectorFile)))
	if err != nil {
		t.Fatal(err)
	}
	var sealed []int
	for i, record := range director {
		if record.Type == journal.TypeKPITruth {
			sealed = append(sealed, i)
		}
	}
	if len(sealed) != 1 || sealed[0] != len(director)-2 || director[len(director)-1].Type != "run.closed" {
		t.Fatalf("kpi.truth records at %v of %d, want exactly one just before run.closed", sealed, len(director))
	}
	observer, err := journal.Read(readFile(t, filepath.Join(dir, journal.ObserverFile)))
	if err != nil {
		t.Fatal(err)
	}
	for _, record := range observer {
		if record.Type == journal.TypeKPITruth {
			t.Fatal("the observer journal carries the KPI truth; it is sealed")
		}
	}
}

func TestTheArtifactRecordsTheDigestOfEachJournalFileAsWritten(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cfg := pondConfig(t)
	cfg.JournalDir = dir
	r, err := New(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	advanceAndReport(t, r, cfg.StartTimeNS+900e9)
	art, err := r.End("")
	if err != nil {
		t.Fatal(err)
	}
	for file, recorded := range map[string]string{journal.ObserverFile: art.JournalFileDigest, journal.DirectorFile: art.SealedFileDigest} {
		if got := canonical.DigestBytes(readFile(t, filepath.Join(dir, file))); got != recorded || recorded == "" {
			t.Errorf("%s: the file digests to %s, the artifact records %q", file, got, recorded)
		}
	}
}
