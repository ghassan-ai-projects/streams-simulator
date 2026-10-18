package run_test

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func gatewayRecords() []map[string]any {
	return []map[string]any{
		{"type": "NBIRTH", "seq": json.Number("0"), "environment": "hardware"},
		{"type": "DDATA", "seq": json.Number("1"), "metrics": []any{map[string]any{"alias": json.Number("1"), "value": json.Number("22.5")}}},
	}
}

type recordedRun struct {
	dir      string
	dirs     run.Directories
	artifact *model.RunArtifact
}

func runWithFieldTrace(t *testing.T, records ...map[string]any) recordedRun {
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
	for _, record := range records {
		if err := trace.Add(record); err != nil {
			t.Fatal(err)
		}
	}
	if err := trace.Close(); err != nil {
		t.Fatal(err)
	}
	r.SetFieldTrace(trace.Digest())
	if _, err := r.Advance(t.Context(), model.DefaultStartTimeNS+300e9, false); err != nil {
		t.Fatal(err)
	}
	art, err := r.End(dir)
	if err != nil {
		t.Fatal(err)
	}
	return recordedRun{dir: dir, dirs: run.Directories{Domains: testsupport.DomainsDir(), Adapters: testsupport.AdaptersDir()}, artifact: art}
}

func TestAFieldTraceIsARunInputAndItsDigestIsPartOfTheArtifact(t *testing.T) {
	t.Parallel()
	recorded := runWithFieldTrace(t, gatewayRecords()...)
	raw, err := os.ReadFile(filepath.Join(recorded.dir, model.FileFieldTrace))
	if err != nil {
		t.Fatal(err)
	}
	if want := canonical.DigestBytes(raw); recorded.artifact.FieldTraceDigest != want {
		t.Fatalf("artifact field_trace_digest = %q, want the digest %q of %s", recorded.artifact.FieldTraceDigest, want, model.FileFieldTrace)
	}
	document, err := os.ReadFile(filepath.Join(recorded.dir, model.FileRunArtifact))
	if err != nil || !strings.Contains(string(document), `"field_trace_digest": "`+recorded.artifact.FieldTraceDigest+`"`) {
		t.Fatalf("run.json (%v) does not record the field trace digest", err)
	}
	if res, err := run.ReplayFile(t.Context(), filepath.Join(recorded.dir, model.FileRunArtifact), recorded.dirs); err != nil || !res.Matches {
		t.Fatalf("replay of an untouched run with a field trace: %+v, %v, want a match", res, err)
	}
	records, digest, err := run.ReadFieldTrace(recorded.dir)
	if err != nil || len(records) != 2 || digest != recorded.artifact.FieldTraceDigest || records[1]["type"] != "DDATA" {
		t.Fatalf("ReadFieldTrace = %v, %q, %v, want both records in order under the recorded digest", records, digest, err)
	}
}

func TestReplayingWithADifferentFieldTraceFailsVerification(t *testing.T) {
	t.Parallel()
	cases := map[string]struct {
		change func(raw []byte) []byte
		reason string
		read   error
	}{
		"a value changed": {func(raw []byte) []byte { return []byte(strings.Replace(string(raw), "22.5", "99", 1)) }, "file_differs", run.ErrFieldTraceDigest},
		"a record appended": {func(raw []byte) []byte {
			return append(raw, []byte(`{"ordinal":2,"record":{"type":"DDATA"}}`+"\n")...)
		}, "file_differs", run.ErrFieldTraceDigest},
		"a record removed": {func(raw []byte) []byte { return raw[:strings.Index(string(raw), "\n")+1] }, "file_differs", run.ErrFieldTraceDigest},
		"the file removed": {nil, "file_missing", run.ErrFieldTraceMissing},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			recorded := runWithFieldTrace(t, gatewayRecords()...)
			path := filepath.Join(recorded.dir, model.FileFieldTrace)
			if tc.change == nil {
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
			} else {
				rewrite(t, path, tc.change)
			}
			res, err := run.ReplayFile(t.Context(), filepath.Join(recorded.dir, model.FileRunArtifact), recorded.dirs)
			if err != nil || res.Matches || res.Reason != tc.reason || res.File != model.FileFieldTrace {
				t.Fatalf("replay = %+v, %v, want a verdict of %s on %s", res, err, tc.reason, model.FileFieldTrace)
			}
			if _, _, err := run.ReadFieldTrace(recorded.dir); !errors.Is(err, tc.read) {
				t.Fatalf("ReadFieldTrace = %v, want %v", err, tc.read)
			}
		})
	}
}

func TestAMalformedFieldTraceDigestIsNotTolerated(t *testing.T) {
	t.Parallel()
	recorded := runWithFieldTrace(t, gatewayRecords()...)
	artifact := filepath.Join(recorded.dir, model.FileRunArtifact)
	editArtifact(t, artifact, func(doc map[string]any) { doc["field_trace_digest"] = "sha256:00" })
	var invalid *model.ValidationError
	if res, err := run.ReplayFile(t.Context(), artifact, recorded.dirs); res != nil || !errors.As(err, &invalid) {
		t.Fatalf("replay = %+v, %v, want the artifact refused for its field_trace_digest", res, err)
	}
}

func TestARunWithoutAFieldTraceHasNeitherTheMemberNorTheFile(t *testing.T) {
	t.Parallel()
	dir, dirs := publishedRun(t)
	document, err := os.ReadFile(filepath.Join(dir, model.FileRunArtifact))
	if err != nil || strings.Contains(string(document), "field_trace") {
		t.Fatalf("run.json (%v) names a field trace although the run took none", err)
	}
	if _, err := os.Stat(filepath.Join(dir, model.FileFieldTrace)); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("stat %s = %v, want it absent", model.FileFieldTrace, err)
	}
	if res, err := run.ReplayFile(t.Context(), filepath.Join(dir, model.FileRunArtifact), dirs); err != nil || !res.Matches {
		t.Fatalf("replay = %+v, %v, want a match: a run without a trace is verified as before", res, err)
	}
	if _, _, err := run.ReadFieldTrace(dir); !errors.Is(err, run.ErrFieldTraceMissing) {
		t.Fatalf("ReadFieldTrace = %v, want ErrFieldTraceMissing", err)
	}
}

func TestAFieldTraceNextToAnArtifactThatDoesNotCoverItIsRefused(t *testing.T) {
	t.Parallel()
	dir, _ := publishedRun(t)
	raw, err := json.Marshal(map[string]any{"ordinal": 0, "record": map[string]any{"type": "NBIRTH"}})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, model.FileFieldTrace), append(raw, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := run.ReadFieldTrace(dir); !errors.Is(err, run.ErrFieldTraceUncovered) {
		t.Fatalf("ReadFieldTrace = %v, want ErrFieldTraceUncovered: run.json records no digest for this file", err)
	}
}

func TestAFieldTraceWithoutAnArtifactIsReadAsLongAsItIsCanonical(t *testing.T) {
	t.Parallel()
	valid := `{"ordinal":0,"record":{"type":"NBIRTH"}}` + "\n"
	cases := map[string]struct {
		content string
		bad     bool
	}{
		"an empty trace":                 {"", false},
		"one canonical record":           {valid, false},
		"no final newline":               {strings.TrimSuffix(valid, "\n"), true},
		"not JSON":                       {"garbage\n", true},
		"an ordinal out of order":        {`{"ordinal":1,"record":{"type":"NBIRTH"}}` + "\n", true},
		"an extra member":                {`{"extra":1,"ordinal":0,"record":{}}` + "\n", true},
		"a record that is no object":     {`{"ordinal":0,"record":[]}` + "\n", true},
		"members out of canonical order": {`{"record":{"type":"NBIRTH"},"ordinal":0}` + "\n", true},
		"spaces":                         {`{"ordinal": 0, "record": {"type":"NBIRTH"}}` + "\n", true},
		"a record over a mebibyte":       {`{"ordinal":0,"record":{"padding":"` + strings.Repeat("x", 1<<20) + `"}}` + "\n", true},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, model.FileFieldTrace), []byte(tc.content), 0o600); err != nil {
				t.Fatal(err)
			}
			_, _, err := run.ReadFieldTrace(dir)
			if tc.bad != errors.Is(err, run.ErrFieldTrace) || (!tc.bad && err != nil) {
				t.Fatalf("ReadFieldTrace = %v, want ErrFieldTrace: %v", err, tc.bad)
			}
		})
	}
}

func TestAFieldTraceRefusesWhatItCannotRecordCanonically(t *testing.T) {
	t.Parallel()
	trace, err := run.NewFieldTrace(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = trace.Close() })
	if err := trace.Add(map[string]any{"samples": []map[string]any{{"alias": 1}}}); !errors.Is(err, run.ErrFieldTrace) {
		t.Fatalf("Add of a value canonical JSON does not take = %v, want ErrFieldTrace", err)
	}
	if err := trace.Add(map[string]any{"padding": strings.Repeat("x", 1<<20)}); !errors.Is(err, run.ErrFieldTrace) {
		t.Fatalf("Add of a record over a mebibyte = %v, want ErrFieldTrace", err)
	}
	if err := trace.Add(map[string]any{"type": "NBIRTH"}); err != nil {
		t.Fatalf("Add after two refusals = %v, want the ordinal not to have moved", err)
	}
}
