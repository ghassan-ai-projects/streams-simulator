package run_test

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

// sidecarCase describes an optional evidence file of a run: the field trace or the interrupt log.
type sidecarCase struct {
	name, file, from, to, appended string
	missing, mismatch              error
	digest                         func(*model.RunArtifact) string
	read                           func(dir string) ([]map[string]any, string, error)
}

var fieldTraceSidecar = sidecarCase{name: "field_trace", file: model.FileFieldTrace, from: "22.5", to: "99",
	appended: `{"ordinal":2,"record":{"type":"DDATA"}}`, missing: run.ErrFieldTraceMissing, mismatch: run.ErrFieldTraceDigest,
	digest: func(a *model.RunArtifact) string { return a.FieldTraceDigest }, read: run.ReadFieldTrace}

var interruptLogSidecar = sidecarCase{name: "interrupt_log", file: model.FileInterruptLog, from: `"op":"state"`, to: `"op":"hold"`,
	appended: `{"ordinal":2,"record":{"event":"clock"}}`, missing: run.ErrInterruptLogMissing, mismatch: run.ErrInterruptLogDigest,
	digest: func(a *model.RunArtifact) string { return a.InterruptLogDigest }, read: run.ReadInterruptLog}

func (sc sidecarCase) digestKey() string { return sc.name + "_digest" }

func (sc sidecarCase) edit(raw []byte) []byte {
	return []byte(strings.Replace(string(raw), sc.from, sc.to, 1))
}

func checkSidecarIsEvidence(t *testing.T, sc sidecarCase, recorded recordedRun, check func([]map[string]any) bool) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(recorded.dir, sc.file))
	if err != nil {
		t.Fatal(err)
	}
	if got, want := sc.digest(recorded.artifact), canonical.DigestBytes(raw); got != want {
		t.Fatalf("artifact %s = %q, want the digest %q of %s", sc.digestKey(), got, want, sc.file)
	}
	document, err := os.ReadFile(filepath.Join(recorded.dir, model.FileRunArtifact))
	if err != nil || !strings.Contains(string(document), `"`+sc.digestKey()+`": "`+sc.digest(recorded.artifact)+`"`) {
		t.Fatalf("run.json (%v) does not record the %s", err, sc.digestKey())
	}
	if res, err := run.ReplayFile(t.Context(), filepath.Join(recorded.dir, model.FileRunArtifact), recorded.dirs); err != nil || !res.Matches {
		t.Fatalf("replay of an untouched run with a %s: %+v, %v, want a match", sc.name, res, err)
	}
	records, digest, err := sc.read(recorded.dir)
	if err != nil || !check(records) || digest != sc.digest(recorded.artifact) {
		t.Fatalf("read %s = %v, %q, %v, want its records in order under the recorded digest", sc.name, records, digest, err)
	}
}

func checkSidecarTampering(t *testing.T, sc sidecarCase, record func(*testing.T) recordedRun) {
	t.Helper()
	cases := map[string]struct {
		change func(raw []byte) []byte
		reason string
		read   error
	}{
		"a value changed":   {sc.edit, "file_differs", sc.mismatch},
		"a record appended": {func(raw []byte) []byte { return append(raw, []byte(sc.appended+"\n")...) }, "file_differs", sc.mismatch},
		"a record removed":  {func(raw []byte) []byte { return raw[:strings.Index(string(raw), "\n")+1] }, "file_differs", sc.mismatch},
		"the file removed":  {nil, "file_missing", sc.missing},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			recorded := record(t)
			path := filepath.Join(recorded.dir, sc.file)
			if tc.change == nil {
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
			} else {
				rewrite(t, path, tc.change)
			}
			res, err := run.ReplayFile(t.Context(), filepath.Join(recorded.dir, model.FileRunArtifact), recorded.dirs)
			if err != nil || res.Matches || res.Reason != tc.reason || res.File != sc.file {
				t.Fatalf("replay = %+v, %v, want a verdict of %s on %s", res, err, tc.reason, sc.file)
			}
			if _, _, err := sc.read(recorded.dir); !errors.Is(err, tc.read) {
				t.Fatalf("read %s = %v, want %v", sc.name, err, tc.read)
			}
		})
	}
}

func checkSidecarAbsent(t *testing.T, sc sidecarCase) {
	t.Helper()
	dir, dirs := publishedRun(t)
	document, err := os.ReadFile(filepath.Join(dir, model.FileRunArtifact))
	if err != nil || strings.Contains(string(document), sc.name) {
		t.Fatalf("run.json (%v) names a %s although the run took none", err, sc.name)
	}
	if _, err := os.Stat(filepath.Join(dir, sc.file)); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("stat %s = %v, want it absent", sc.file, err)
	}
	if res, err := run.ReplayFile(t.Context(), filepath.Join(dir, model.FileRunArtifact), dirs); err != nil || !res.Matches {
		t.Fatalf("replay = %+v, %v, want a match: a run without a %s is verified as before", res, err, sc.name)
	}
	if _, _, err := sc.read(dir); !errors.Is(err, sc.missing) {
		t.Fatalf("read %s = %v, want %v", sc.name, err, sc.missing)
	}
}
