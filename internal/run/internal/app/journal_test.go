package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/journal"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

// canarySeed is distinctive, so the canary can look for its decimal value.
const canarySeed = 7424242

// scriptedJournalRun runs a fixed session into dir: a perturbation, a fault, an
// effector call, a consumer report and two advances. It returns the artifact
// and the fault id the world assigned.
func scriptedJournalRun(t *testing.T, dir string, edit func(*Config)) (*model.RunArtifact, string) {
	t.Helper()
	spec, a := testBase(t)
	cfg := Config{Domain: spec, Adapter: a, Seed: canarySeed, SinkName: model.SinkFile, SinkTarget: filepath.Join(dir, "trace.jsonl"),
		TimeMode: model.TimeStepped, StartTimeNS: model.DefaultStartTimeNS, JournalDir: dir}
	if edit != nil {
		edit(&cfg)
	}
	r, err := New(t.Context(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	pond, now := r.World.EntityIDs()[0], r.World.Clock()
	if _, err := r.ApplyPerturb("duplicate_burst", nil, 0, 0); err != nil {
		t.Fatal(err)
	}
	fault, err := r.InjectFault(pond, spec.Spec.Kinds[0].Faults[0].ID, now+1200e9, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.InvokeEffector("start_aerator", pond, "cmd-1", map[string]any{"pond_id": pond, "level": 1.0}, now); err != nil {
		t.Fatal(err)
	}
	advanceAndReport(t, r, now+1800e9)
	art, err := r.End(dir)
	if err != nil {
		t.Fatal(err)
	}
	return art, fault
}

func advanceAndReport(t *testing.T, r *Run, to int64) {
	t.Helper()
	if _, err := r.Advance(t.Context(), to, false); err != nil {
		t.Fatal(err)
	}
	r.ReportQuiesced(to)
}

func readFile(t *testing.T, path string) []byte {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

// The canary: nothing a consumer may read before the run is sealed names the
// injected fault or the perturbation; the director stream does.
func TestEvaluateRunKeepsInjectedTruthOutOfConsumerReadableFiles(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	spec, _ := testBase(t)
	_, fault := scriptedJournalRun(t, dir, nil)
	secrets := []string{`"` + fault + `"`, spec.Spec.Kinds[0].Faults[0].ID, "duplicate_burst", `"seed"`, strconv.Itoa(canarySeed)}
	for _, name := range []string{journal.ObserverFile, "trace.jsonl"} {
		raw := readFile(t, filepath.Join(dir, name))
		for _, secret := range secrets {
			if bytes.Contains(raw, []byte(secret)) {
				t.Errorf("%s names %q before sealing", name, secret)
			}
		}
	}
	sealed := readFile(t, filepath.Join(dir, journal.DirectorFile))
	for _, secret := range append(secrets[1:3:3], `"seed":`+strconv.Itoa(canarySeed)) {
		if !bytes.Contains(sealed, []byte(secret)) {
			t.Errorf("the director stream must name %q", secret)
		}
	}
}

func TestBothJournalsAreByteIdenticalAcrossRunsAndReplays(t *testing.T) {
	t.Parallel()
	first, second := t.TempDir(), t.TempDir()
	artifact, _ := scriptedJournalRun(t, first, nil)
	other, _ := scriptedJournalRun(t, second, nil)
	for _, name := range []string{journal.ObserverFile, journal.DirectorFile} {
		if !bytes.Equal(readFile(t, filepath.Join(first, name)), readFile(t, filepath.Join(second, name))) {
			t.Errorf("%s differs between two runs of the same inputs", name)
		}
	}
	if artifact.JournalDigest != other.JournalDigest || artifact.SealedJournalDigest != other.SealedJournalDigest || artifact.JournalDigest == "" {
		t.Fatalf("artifact digests: %+v vs %+v", artifact, other)
	}
	res, err := ReplayFile(t.Context(), filepath.Join(first, "run.json"), Directories{Domains: testsupport.DomainsDir(), Adapters: testsupport.AdaptersDir()})
	if err != nil || !res.Matches || len(res.JournalDiverged) != 0 {
		t.Fatalf("replay = %+v, %v; want a match", res, err)
	}
}

// A consumer report is a session record: it is in the file but outside the
// digest, because replay cannot reproduce it.
func TestConsumerReportsAreWrittenButNotPartOfTheReplayIdentity(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	scriptedJournalRun(t, dir, nil)
	records, err := decodeJournalBytes(readFile(t, filepath.Join(dir, journal.ObserverFile)))
	if err != nil {
		t.Fatal(err)
	}
	reported := 0
	for _, rec := range records {
		if rec.Type == "consumer.reported" {
			reported++
		}
	}
	if reported != 1 {
		t.Fatalf("consumer.reported records = %d, want 1", reported)
	}
}

func TestVerifyDetectsAChangedRecordInEitherJournalStream(t *testing.T) {
	t.Parallel()
	dirs := Directories{Domains: testsupport.DomainsDir(), Adapters: testsupport.AdaptersDir()}
	for _, name := range []string{journal.ObserverFile, journal.DirectorFile} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			scriptedJournalRun(t, dir, nil)
			path := filepath.Join(dir, name)
			changed := bytes.Replace(readFile(t, path), []byte(`"seq":`), []byte(`"seq" :`), 3)
			if err := os.WriteFile(path, changed, 0o600); err != nil {
				t.Fatal(err)
			}
			res, err := ReplayFile(t.Context(), filepath.Join(dir, "run.json"), dirs)
			if err != nil || res.Matches || len(res.JournalDiverged) != 1 || !strings.Contains(res.Detail, name+" differs from the replay at record ") {
				t.Fatalf("replay = %+v, %v; want a mismatch naming %s", res, err, name)
			}
		})
	}
}

func TestVerifyDetectsAChangedJournalDigestInTheArtifact(t *testing.T) {
	t.Parallel()
	spec, a := testBase(t)
	for stream, edit := range map[string]func(*model.RunArtifact){
		"observer": func(art *model.RunArtifact) { art.JournalDigest = "sha256:" + strings.Repeat("0", 64) },
		"director": func(art *model.RunArtifact) { art.SealedJournalDigest = "sha256:" + strings.Repeat("0", 64) },
	} {
		t.Run(stream, func(t *testing.T) {
			t.Parallel()
			art, _ := scriptedJournalRun(t, t.TempDir(), nil)
			edit(art)
			res, err := ReplayArtifact(t.Context(), art, spec, a, "")
			if err != nil || res.Matches || len(res.JournalDiverged) != 1 || res.JournalDiverged[0] != stream ||
				!strings.Contains(res.Detail, "the "+stream+" journal stream differs from the replay") {
				t.Fatalf("replay = %+v, %v; want a mismatch naming the %s stream", res, err, stream)
			}
		})
	}
}
