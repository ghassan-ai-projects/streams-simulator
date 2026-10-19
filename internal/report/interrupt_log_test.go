package report_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/report"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func folderWithInterruptLog(t *testing.T) string {
	t.Helper()
	return folderWith(t, func(dir string, r *run.Run) {
		log, err := run.NewInterruptLog(dir)
		if err != nil {
			t.Fatal(err)
		}
		if err := log.Add(map[string]any{"event": "control", "op": "state", "clock": "world", "node_ns": json.Number("1")}); err != nil {
			t.Fatal(err)
		}
		if err := log.Close(); err != nil {
			t.Fatal(err)
		}
		r.SetInterruptLog(log.Digest())
	})
}

func TestTheReportChecksTheInterruptLogAgainstTheDigestTheRunRecorded(t *testing.T) {
	t.Parallel()
	dir := folderWithInterruptLog(t)
	built, err := report.Build(t.Context(), dir)
	if err != nil || statusOfCheck(t, built, "interrupt_log") != "match" || built.Verification.Status != "files_match" {
		t.Fatalf("report = %+v, %v, want the interrupt log to match the recorded digest", built.Verification, err)
	}
	raw, err := os.ReadFile(filepath.Join(dir, model.FileInterruptLog))
	if err != nil {
		t.Fatal(err)
	}
	testsupport.WriteFileIn(t, dir, model.FileInterruptLog, append(raw, []byte(`{"ordinal":1,"record":{"event":"clock"}}`+"\n")...))
	built, err = report.Build(t.Context(), dir)
	if err != nil || statusOfCheck(t, built, "interrupt_log") != "mismatch" || built.Verification.Status != "files_differ" {
		t.Fatalf("report = %+v, %v, want the appended record to make the interrupt log differ", built.Verification, err)
	}
}
