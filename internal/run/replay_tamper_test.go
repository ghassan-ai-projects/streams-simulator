package run_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func editArtifact(t *testing.T, artifact string, edit func(doc map[string]any)) {
	t.Helper()
	raw, err := os.ReadFile(artifact)
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	edit(doc)
	out, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	place(t, artifact, out)
}

func editDomainSpec(t *testing.T, artifact string, redigest bool) {
	t.Helper()
	editArtifact(t, artifact, func(doc map[string]any) {
		raw, _ := json.Marshal(doc["domain_spec"])
		edited := bytes.Replace(raw, []byte(`"initial":6.5,`), []byte(`"initial":6.4,`), 1)
		if bytes.Equal(raw, edited) {
			t.Fatal("the embedded domain spec has no initial 6.5 to edit")
		}
		var spec map[string]any
		if err := json.Unmarshal(edited, &spec); err != nil {
			t.Fatal(err)
		}
		doc["domain_spec"] = spec
		if redigest {
			compiled, err := domain.Parse(edited, "tampered")
			if err != nil {
				t.Fatal(err)
			}
			doc["domain"].(map[string]any)["digest"] = compiled.Digest
		}
	})
}

func TestReplayFileReportsABrokenRecordedDigestAsAVerdictNotAFailure(t *testing.T) {
	t.Parallel()
	cases := map[string]struct {
		tamper func(t *testing.T, artifact string)
		detail string
		reason string
		file   string
	}{
		"the embedded spec edited": {
			func(t *testing.T, artifact string) { editDomainSpec(t, artifact, false) },
			"the recorded domain spec does not match its digest",
			"recorded_spec", "",
		},
		"the embedded spec and its digest rewritten together": {
			func(t *testing.T, artifact string) { editDomainSpec(t, artifact, true) },
			"the recorded specs replay to a different world than the recorded world digest",
			"world_digest", "",
		},
		"the trace file with an appended newline": {
			func(t *testing.T, artifact string) {
				rewrite(t, filepath.Join(filepath.Dir(artifact), model.FileTrace), func(raw []byte) []byte { return append(raw, '\n') })
			},
			model.FileTrace + " differs from the digest run.json recorded",
			"file_differs", model.FileTrace,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			artifact, _, dirs := runInEditableDomains(t)
			tc.tamper(t, artifact)
			res, err := run.ReplayFile(t.Context(), artifact, dirs)
			if err != nil || res.Matches || !strings.Contains(res.Detail, tc.detail) || res.WantDigest == "" || res.Reason != tc.reason || res.File != tc.file {
				t.Fatalf("%+v, %v; want matches:false, reason %s, file %q, with detail %q", res, err, tc.reason, tc.file, tc.detail)
			}
		})
	}
}

func TestReplayFileStaysAFailureWhenTheRunCannotBeReplayed(t *testing.T) {
	t.Parallel()
	artifact, domainFile, dirs := runInEditableDomains(t)
	if res, err := run.ReplayFile(t.Context(), filepath.Join(filepath.Dir(artifact), "absent.json"), dirs); res != nil || !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("a missing artifact: %+v, %v", res, err)
	}
	editArtifact(t, artifact, func(doc map[string]any) { delete(doc, "domain_spec"); delete(doc, "adapter_spec") })
	rewrite(t, domainFile, func(raw []byte) []byte {
		return bytes.Replace(raw, []byte(`"initial": 6.5,`), []byte(`"initial": 6.4,`), 1)
	})
	if res, err := run.ReplayFile(t.Context(), artifact, dirs); res != nil || !errors.Is(err, run.ErrDigestMismatch) {
		t.Fatalf("an artifact without recorded specs against a changed domain: %+v, %v; want ErrDigestMismatch and no result", res, err)
	}
}

func TestExportScenarioFileUsesTheRecordedSpecsAfterTheDomainChanges(t *testing.T) {
	t.Parallel()
	artifact, domainFile, dirs := runInEditableDomains(t)
	before, err := run.ExportScenarioFile(t.Context(), artifact, dirs)
	if err != nil {
		t.Fatal(err)
	}
	rewrite(t, domainFile, func(raw []byte) []byte {
		return bytes.Replace(raw, []byte(`"initial": 6.5,`), []byte(`"initial": 6.4,`), 1)
	})
	after, err := run.ExportScenarioFile(t.Context(), artifact, dirs)
	if err != nil || !bytes.Equal(before.Raw, after.Raw) {
		t.Fatalf("after the domain edit: %v; the export must equal the one made before", err)
	}
	editDomainSpec(t, artifact, false)
	if _, err := run.ExportScenarioFile(t.Context(), artifact, dirs); !errors.Is(err, run.ErrDigestMismatch) {
		t.Fatalf("a recorded spec that no longer matches its digest: %v, want ErrDigestMismatch", err)
	}
}

func TestReplayFileOmitsDomainChangedWhenTheCurrentDomainIsUnknown(t *testing.T) {
	t.Parallel()
	artifact, domainFile, dirs := runInEditableDomains(t)
	if err := os.Remove(domainFile); err != nil {
		t.Fatal(err)
	}
	other, err := os.ReadFile(testsupport.Domain("rotating-machinery"))
	if err != nil {
		t.Fatal(err)
	}
	withoutTheRunsDomain := t.TempDir()
	place(t, filepath.Join(withoutTheRunsDomain, "rotating-machinery.domain.json"), other)
	for name, domains := range map[string]string{"the domain was removed": dirs.Domains, "another domain only": withoutTheRunsDomain,
		"the directory is absent": filepath.Join(t.TempDir(), "absent")} {
		res, err := run.ReplayFile(t.Context(), artifact, run.Directories{Domains: domains, Adapters: dirs.Adapters})
		if err != nil || !res.Matches || res.DomainChanged || res.CurrentDomainDigest != "" {
			t.Fatalf("%s: %+v, %v; want a match with domain_changed omitted (unknown)", name, res, err)
		}
	}
}
