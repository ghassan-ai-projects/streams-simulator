package report_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/report"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func folderWithFieldTrace(t *testing.T) string {
	t.Helper()
	spec, err := domain.Load(testsupport.Domain("aquaculture-pond"))
	if err != nil {
		t.Fatal(err)
	}
	a, err := adapter.Load(testsupport.Adapter("native-jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	r, err := run.New(t.Context(), run.Config{Domain: spec, Adapter: a, Seed: 4, SinkName: model.SinkInproc, TimeMode: model.TimeStepped,
		StartTimeNS: model.DefaultStartTimeNS, JournalDir: dir})
	if err != nil {
		t.Fatal(err)
	}
	trace, err := run.NewFieldTrace(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := trace.Add(map[string]any{"type": "NBIRTH", "seq": json.Number("0")}); err != nil {
		t.Fatal(err)
	}
	if err := trace.Close(); err != nil {
		t.Fatal(err)
	}
	r.SetFieldTrace(trace.Digest())
	if _, err := r.Advance(t.Context(), model.DefaultStartTimeNS+120e9, false); err != nil {
		t.Fatal(err)
	}
	if _, err := r.End(dir); err != nil {
		t.Fatal(err)
	}
	return dir
}

func statusOfCheck(t *testing.T, built *report.Report, name string) string {
	t.Helper()
	for _, check := range built.Verification.Checks {
		if check.Name == name {
			return check.Status
		}
	}
	t.Fatalf("verification checks %+v hold no %s", built.Verification.Checks, name)
	return ""
}

func TestTheReportChecksTheFieldTraceAgainstTheDigestTheRunRecorded(t *testing.T) {
	t.Parallel()
	dir := folderWithFieldTrace(t)
	built, err := report.Build(t.Context(), dir)
	if err != nil || statusOfCheck(t, built, "field_trace") != "match" || built.Verification.Status != "files_match" {
		t.Fatalf("report = %+v, %v, want the field trace to match the recorded digest", built.Verification, err)
	}
	path := filepath.Join(dir, model.FileFieldTrace)
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	testsupport.WriteFileIn(t, dir, model.FileFieldTrace, append(raw, []byte(`{"ordinal":1,"record":{"type":"DDATA"}}`+"\n")...))
	built, err = report.Build(t.Context(), dir)
	if err != nil || statusOfCheck(t, built, "field_trace") != "mismatch" || built.Verification.Status != "files_differ" {
		t.Fatalf("report = %+v, %v, want the appended record to make the field trace differ", built.Verification, err)
	}
}
