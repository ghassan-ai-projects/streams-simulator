package durable

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestLedgerAppendsFlushesAndClosesItsFile(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "nested", "ledger.jsonl")
	l, err := OpenLedger(path)
	if err != nil {
		t.Fatal(err)
	}
	l.Append(model.LedgerRecord{DeliveryReason: "ok"})
	if err := l.Flush(); err != nil {
		t.Fatal(err)
	}
	if raw, err := os.ReadFile(path); err != nil || !strings.Contains(string(raw), `"delivery_reason":"ok"`) {
		t.Fatalf("flushed ledger = %q (%v)", raw, err)
	}
	if err := l.Finish(); err != nil {
		t.Fatalf("finish: %v", err)
	}
	l.Append(model.LedgerRecord{DeliveryReason: "late"})
	if err := l.Flush(); err == nil || !strings.HasPrefix(err.Error(), "run: flush ledger: ") {
		t.Fatalf("a flush after finish = %v", err)
	}
}

func TestOpenLedgerNamesAPathItCannotCreate(t *testing.T) {
	t.Parallel()
	blocker := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(blocker, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := OpenLedger(filepath.Join(blocker, "sub", "l.jsonl")); err == nil || !strings.Contains(err.Error(), "run: open ledger") {
		t.Fatalf("err = %v", err)
	}
}

func TestPublishEvidenceWritesTheDeclaredFiles(t *testing.T) {
	t.Parallel()
	dir := filepath.Join(t.TempDir(), "out")
	err := PublishEvidence(dir, Evidence{
		Trace:       []byte("t\n"),
		Ledger:      []model.LedgerRecord{{DeliveryReason: "ok"}},
		WriteLedger: true,
		History:     []model.StateSnapshot{{Seq: 1, Entity: "e", States: map[string]float64{"x": 1}}},
		Verdict:     &model.Verdict{},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"trace.jsonl", "ledger.jsonl", "world_state_history.jsonl", "verdict.json"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Errorf("missing %s: %v", name, err)
		}
	}
	other := t.TempDir()
	if err := PublishEvidence(other, Evidence{Trace: []byte("t\n"), WriteLedger: false}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(other, "ledger.jsonl")); !os.IsNotExist(err) {
		t.Fatalf("a durably kept ledger must not be rewritten: %v", err)
	}
	if _, err := os.Stat(filepath.Join(other, "verdict.json")); !os.IsNotExist(err) {
		t.Fatalf("no verdict, no verdict.json: %v", err)
	}
}

func TestWriteRunArtifactAndReadArtifact(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := WriteRunArtifact("", &model.RunArtifact{}); err != nil {
		t.Fatalf("an empty directory writes nothing: %v", err)
	}
	if err := WriteRunArtifact(dir, &model.RunArtifact{RunID: "r-1"}); err != nil {
		t.Fatal(err)
	}
	raw, err := ReadArtifact(filepath.Join(dir, "run.json"))
	if err != nil || !strings.Contains(string(raw), `"run_id": "r-1"`) {
		t.Fatalf("read = %q (%v)", raw, err)
	}
	if _, err := ReadArtifact(filepath.Join(dir, "absent.json")); err == nil || !strings.Contains(err.Error(), "run: read artifact") {
		t.Fatalf("missing artifact: %v", err)
	}
}

func TestPublishEvidenceRefusesAnOutputDirUnderAFile(t *testing.T) {
	t.Parallel()
	blocker := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(blocker, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	err := PublishEvidence(filepath.Join(blocker, "out"), Evidence{Trace: []byte("t\n")})
	if err == nil || !strings.HasPrefix(err.Error(), "End: ") {
		t.Fatalf("publish under a regular file = %v", err)
	}
}

func TestReadJournalReportsPresenceAndFailures(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "journal.jsonl"), []byte("x\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	raw, present, err := ReadJournal(dir, "journal.jsonl")
	if err != nil || !present || string(raw) != "x\n" {
		t.Fatalf("read = %q, %v, %v", raw, present, err)
	}
	if _, present, err := ReadJournal(dir, "sealed/journal.jsonl"); err != nil || present {
		t.Fatalf("an absent journal is not an error: present %v, %v", present, err)
	}
	if _, _, err := ReadJournal(dir, "journal.jsonl/inner"); err == nil || !strings.Contains(err.Error(), "run: read journal") {
		t.Fatalf("a file used as a directory = %v", err)
	}
}
