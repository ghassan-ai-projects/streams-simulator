package run_test

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

func interruptRecords() []map[string]any {
	return []map[string]any{
		{"event": "inventory", "devices": []any{}, "clock": "world", "node_ns": json.Number("1")},
		{"event": "control", "op": "state", "clock": "world", "node_ns": json.Number("1")},
	}
}

func runWithInterruptLog(t *testing.T, records ...map[string]any) recordedRun {
	t.Helper()
	dir, r := openedRun(t)
	log, err := run.NewInterruptLog(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, record := range records {
		if err := log.Add(record); err != nil {
			t.Fatal(err)
		}
	}
	if err := log.Close(); err != nil {
		t.Fatal(err)
	}
	r.SetInterruptLog(log.Digest())
	r.JournalInterrupt(map[string]any{"event": "control", "op": "state", "clock": "world", "node_ns": 1})
	return endedRun(t, dir, r)
}

func TestAnInterruptLogIsEvidenceOfTheRunAndItsDigestIsPartOfTheArtifact(t *testing.T) {
	t.Parallel()
	checkSidecarIsEvidence(t, interruptLogSidecar, runWithInterruptLog(t, interruptRecords()...), func(records []map[string]any) bool {
		return len(records) == 2 && records[1]["op"] == "state"
	})
}

func TestTheJournalCarriesTheInterruptRecordsOfARunAsObserverSessionRecords(t *testing.T) {
	t.Parallel()
	recorded := runWithInterruptLog(t, interruptRecords()...)
	raw, err := os.ReadFile(filepath.Join(recorded.dir, model.FileObserverJournal))
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
		var record map[string]any
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			t.Fatal(err)
		}
		if record["type"] == "interrupt" {
			found = record["stream"] == "observer" && record["event"] == "control"
		}
	}
	if !found {
		t.Fatal("the observer journal holds no interrupt record")
	}
}

func TestReplayingWithADifferentInterruptLogFailsVerification(t *testing.T) {
	t.Parallel()
	checkSidecarTampering(t, interruptLogSidecar, func(t *testing.T) recordedRun { return runWithInterruptLog(t, interruptRecords()...) })
}

func TestARunWithoutAnInterruptLogHasNeitherTheMemberNorTheFile(t *testing.T) {
	t.Parallel()
	checkSidecarAbsent(t, interruptLogSidecar)
}

func TestAnInterruptLogNextToAnArtifactThatDoesNotCoverItIsRefused(t *testing.T) {
	t.Parallel()
	dir, _ := publishedRun(t)
	line := `{"ordinal":0,"record":{"event":"clock"}}` + "\n"
	if err := os.WriteFile(filepath.Join(dir, model.FileInterruptLog), []byte(line), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := run.ReadInterruptLog(dir); !errors.Is(err, run.ErrInterruptLogUncovered) {
		t.Fatalf("ReadInterruptLog = %v, want ErrInterruptLogUncovered: run.json records no digest for this file", err)
	}
}
