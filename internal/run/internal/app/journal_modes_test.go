package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/journal"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

func readJournal(t *testing.T, dir, name string) []journal.Record {
	t.Helper()
	records, err := decodeJournalBytes(readFile(t, filepath.Join(dir, name)))
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	return records
}

func recordsOfType(records []journal.Record, recordType string) []journal.Record {
	var found []journal.Record
	for _, rec := range records {
		if rec.Type == recordType {
			found = append(found, rec)
		}
	}
	return found
}

// A run that has not ended has already flushed its journals at each command
// boundary, and a crash can cut them at any byte: the reader still reads up to
// the last complete line.
func TestJournalOfARunThatHasNotEndedIsReadableUpToItsLastCompleteLine(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	spec, a := testBase(t)
	r, err := New(t.Context(), Config{Domain: spec, Adapter: a, Seed: 3, SinkName: model.SinkInproc, JournalDir: dir})
	if err != nil {
		t.Fatal(err)
	}
	advanceAndReport(t, r, r.World.Clock()+600e9)
	raw := readFile(t, filepath.Join(dir, journal.ObserverFile))
	records, err := decodeJournalBytes(raw)
	if err != nil || len(records) < 3 || records[0].Type != "run.opened" || len(recordsOfType(records, "run.closed")) != 0 {
		t.Fatalf("observer journal before End: %d records, %v", len(records), err)
	}
	cut := raw[:len(raw)-7]
	torn, err := decodeJournalBytes(cut)
	if err != nil || len(torn) != len(records)-1 {
		t.Fatalf("a torn last line: %d records, %v; want %d", len(torn), err, len(records)-1)
	}
	if _, err := r.End(dir); err != nil {
		t.Fatal(err)
	}
}

func TestExploreRunIsUnblindedFromCreationAndStillReplays(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	art, _ := scriptedJournalRun(t, dir, func(cfg *Config) { cfg.Explore = true })
	if !art.Unblinded || art.Mode != journal.ModeExplore || art.UnblindedAt == "" {
		t.Fatalf("artifact unblinded %v mode %q at %q", art.Unblinded, art.Mode, art.UnblindedAt)
	}
	opened := recordsOfType(readJournal(t, dir, journal.ObserverFile), "run.opened")
	if len(opened) != 1 || opened[0].Fields["mode"] != journal.ModeExplore {
		t.Fatalf("run.opened = %+v", opened)
	}
	unblinded := recordsOfType(readJournal(t, dir, journal.DirectorFile), "run.unblinded")
	if len(unblinded) != 1 || unblinded[0].Seq != 1 || unblinded[0].Fields["reason"] != journal.ReasonExplore {
		t.Fatalf("run.unblinded = %+v, want the record right after run.opened", unblinded)
	}
	res, err := ReplayFile(t.Context(), filepath.Join(dir, "run.json"), Directories{Domains: testsupport.DomainsDir(), Adapters: testsupport.AdaptersDir()})
	if err != nil || !res.Matches {
		t.Fatalf("replay of an explore run = %+v, %v", res, err)
	}
}

func TestRunUnblindedWhileOpenRecordsTheRevealAtTheCloseAndStillReplays(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	spec, a := testBase(t)
	r, err := New(t.Context(), Config{Domain: spec, Adapter: a, Seed: 4, SinkName: model.SinkInproc, JournalDir: dir})
	if err != nil {
		t.Fatal(err)
	}
	advanceAndReport(t, r, r.World.Clock()+600e9)
	r.Unblind()
	art, err := r.End(dir)
	if err != nil || !art.Unblinded || art.Mode != journal.ModeEvaluate {
		t.Fatalf("End = %+v, %v", art, err)
	}
	unblinded := recordsOfType(readJournal(t, dir, journal.DirectorFile), "run.unblinded")
	if len(unblinded) != 1 || unblinded[0].Fields["reason"] != journal.ReasonRevealed {
		t.Fatalf("run.unblinded = %+v", unblinded)
	}
	res, err := ReplayArtifact(t.Context(), art, spec, a, "")
	if err != nil || !res.Matches {
		t.Fatalf("replay = %+v, %v", res, err)
	}
}

// The observer journal records exactly the acknowledgement the operator got,
// which is the same for every acknowledging failure mode; the director journal
// keeps the real outcome.
func TestAcknowledgementInTheObserverJournalDoesNotRevealTheFailureMode(t *testing.T) {
	t.Parallel()
	responses, outcomes := map[world.FailureMode]string{}, map[world.FailureMode]string{}
	for _, mode := range []world.FailureMode{world.ModeOK, world.ModeSilentNoEffect, world.ModeConfirmedNoEffect} {
		dir := t.TempDir()
		scriptedJournalRun(t, dir, func(cfg *Config) { cfg.ForceFailureMode = mode })
		acks := recordsOfType(readJournal(t, dir, journal.ObserverFile), "effector.acknowledged")
		sealed := recordsOfType(readJournal(t, dir, journal.DirectorFile), "effector.outcome")
		if len(acks) != 1 || len(sealed) != 1 {
			t.Fatalf("%s: %d acknowledgements, %d outcomes, want one each", mode, len(acks), len(sealed))
		}
		responses[mode], outcomes[mode] = marshal(t, acks[0].Fields), marshal(t, sealed[0].Fields)
	}
	if responses[world.ModeOK] != responses[world.ModeSilentNoEffect] || responses[world.ModeOK] != responses[world.ModeConfirmedNoEffect] {
		t.Fatalf("acknowledgements differ by failure mode: %v", responses)
	}
	if outcomes[world.ModeOK] == outcomes[world.ModeSilentNoEffect] || !strings.Contains(outcomes[world.ModeSilentNoEffect], "silent_no_effect") {
		t.Fatalf("the director outcome must carry the real mode: %v", outcomes)
	}
}

func marshal(t *testing.T, value any) string {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func TestDeclinedEffectorCallIsInTheDirectorJournalOnly(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	spec, a := testBase(t)
	r, err := New(t.Context(), Config{Domain: spec, Adapter: a, Seed: 4, SinkName: model.SinkInproc, JournalDir: dir})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.InvokeEffector("start_aerator", "no-such-pond", "cmd-9", map[string]any{}, r.World.Clock()); !errors.Is(err, world.ErrInvalidInvocation) {
		t.Fatalf("InvokeEffector on an unknown entity = %v, want %v", err, world.ErrInvalidInvocation)
	}
	if _, err := r.End(dir); err != nil {
		t.Fatal(err)
	}
	observer, sealed := readFile(t, filepath.Join(dir, journal.ObserverFile)), readFile(t, filepath.Join(dir, journal.DirectorFile))
	if bytes.Contains(observer, []byte("cmd-9")) || !bytes.Contains(sealed, []byte(`"refused"`)) {
		t.Fatalf("a declined call is not acknowledged: observer names it %v, director names the refusal %v",
			bytes.Contains(observer, []byte("cmd-9")), bytes.Contains(sealed, []byte(`"refused"`)))
	}
}

func TestRecordThatCannotBeEncodedFailsTheRun(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	spec, a := testBase(t)
	r, err := New(t.Context(), Config{Domain: spec, Adapter: a, Seed: 4, SinkName: model.SinkInproc, JournalDir: dir})
	if err != nil {
		t.Fatal(err)
	}
	pond := r.World.EntityIDs()[0]
	_, _ = r.InvokeEffector("start_aerator", pond, "cmd-1", map[string]any{"pond_id": pond, "level": math.NaN()}, r.World.Clock())
	if _, err := r.Advance(t.Context(), r.World.Clock()+60e9, false); err == nil || !strings.Contains(err.Error(), "encode command record") {
		t.Fatalf("Advance after an unencodable record = %v, want the encode failure", err)
	}
	art, err := r.End(dir)
	if err == nil || art == nil || !art.Incomplete {
		t.Fatalf("End = %+v, %v; want an incomplete artifact and the failure", art, err)
	}
}

func TestJournalDirectoryThatCannotBeCreatedFailsRunCreation(t *testing.T) {
	t.Parallel()
	blocker := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(blocker, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	spec, a := testBase(t)
	_, err := New(t.Context(), Config{Domain: spec, Adapter: a, SinkName: model.SinkInproc, JournalDir: filepath.Join(blocker, "run")})
	if err == nil || !strings.Contains(err.Error(), "run: journal: open") {
		t.Fatalf("New = %v", err)
	}
}

func TestRunWithoutAJournalDirectoryKeepsTheDigestsOnly(t *testing.T) {
	t.Parallel()
	spec, a := testBase(t)
	r, err := New(t.Context(), Config{Domain: spec, Adapter: a, Seed: 4, SinkName: model.SinkInproc})
	if err != nil {
		t.Fatal(err)
	}
	art, err := r.End("")
	if err != nil || art.JournalDigest == "" || art.SealedJournalDigest == "" || art.Mode != journal.ModeEvaluate {
		t.Fatalf("End = %+v, %v", art, err)
	}
	r.ReportQuiesced(1) // a late report after the close is ignored, not an error
}
