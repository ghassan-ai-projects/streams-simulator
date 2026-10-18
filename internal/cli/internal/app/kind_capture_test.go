package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func pumpNodeJournal(t *testing.T) string {
	t.Helper()
	socket := filepath.Join(filepath.Dir(testsupport.SocketPath(t)), "h.sock")
	reports := referenceHost(t, socket)
	out := t.TempDir()
	invoke(t, "plantwire", "node", "--host-socket", socket, "--domain", "pump-board", "--domains-dir", filepath.Dir(testsupport.PumpBoardDomain()),
		"--adapters-dir", testsupport.AdaptersDir(), "--out", out, "--duration", "180").mustSucceed(t)
	if report := <-reports; report.Refused != "" {
		t.Fatalf("the host refused the node: %q", report.Refused)
	}
	return filepath.Join(out, "journal.jsonl")
}

func birthInJournal(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(string(raw), "\n") {
		var wire struct {
			Type      string          `json:"type"`
			Direction string          `json:"direction"`
			Record    json.RawMessage `json:"record"`
		}
		if json.Unmarshal([]byte(line), &wire) == nil && wire.Type == "wire" && wire.Direction == "node" && strings.Contains(string(wire.Record), `"NBIRTH"`) {
			return string(wire.Record)
		}
	}
	t.Fatalf("journal %s holds no NBIRTH wire record of the node", path)
	return ""
}

func capturedKind(t *testing.T, doc map[string]any, name string) map[string]any {
	t.Helper()
	kinds, _ := doc["kinds"].([]any)
	for _, kind := range kinds {
		if entry, _ := kind.(map[string]any); entry["kind"] == name {
			return entry
		}
	}
	t.Fatalf("capture %v holds no kind %s", doc, name)
	return nil
}

func TestKindCaptureDraftsTheKindsOfARecordedNodeAndComparesThemWithTheKindFiles(t *testing.T) {
	t.Parallel()
	journal := pumpNodeJournal(t)
	plain := invoke(t, "kind", "capture", journal).mustSucceed(t).document(t, "kind_capture")
	pump := capturedKind(t, plain, "pump")
	draft, _ := pump["draft"].(map[string]any)
	if _, digested := draft["kind_digest"]; plain["environment"] != "simulated" || plain["bd_seq"] != float64(0) || digested || pump["comparison"] != nil {
		t.Fatalf("kind capture = %v, want the draft of pump without kind_digest or comparison from a simulated birth", plain)
	}
	same := invoke(t, "kind", "capture", journal, "--kinds-dir", pumpKindsDirectory(t, asWritten)).mustSucceed(t).document(t, "kind_capture")
	if comparison, _ := capturedKind(t, same, "pump")["comparison"].(map[string]any); comparison["status"] != "same" || comparison["file_digest"] != pump["kind_digest"] {
		t.Fatalf("comparison = %v, want pump the same as its kind file", comparison)
	}
	drifted := pumpKindsDirectory(t, func(raw string) string { return strings.Replace(raw, `"m3/h"`, `"l/s"`, 1) })
	differs := invoke(t, "kind", "capture", "--kinds-dir", drifted, journal).mustSucceed(t).document(t, "kind_capture")
	if comparison, _ := capturedKind(t, differs, "pump")["comparison"].(map[string]any); comparison["status"] != "differs" || comparison["differs_at"] != "/metrics/0/unit" {
		t.Fatalf("comparison = %v, want a difference at the unit", comparison)
	}
}

func TestKindCaptureReadsTheBirthOrTheWholeJournalFromStandardInput(t *testing.T) {
	t.Parallel()
	journal := pumpNodeJournal(t)
	raw, err := os.ReadFile(journal)
	if err != nil {
		t.Fatal(err)
	}
	fromFile := invoke(t, "kind", "capture", journal).mustSucceed(t).stdout
	for name, input := range map[string]string{"the birth": birthInJournal(t, journal), "the journal": string(raw)} {
		if got := invokeWithInput(t, input, "kind", "capture", "--stdin").mustSucceed(t).stdout; got != fromFile {
			t.Errorf("capture of %s from standard input differs from the capture of the file:\n%s\n%s", name, got, fromFile)
		}
	}
}

func TestKindCaptureNeedsExactlyOneSourceAndKnowsItsFlags(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		args []string
	}{
		{"no source", []string{"kind", "capture"}},
		{"stdin and a file", []string{"kind", "capture", "--stdin", "birth.json"}},
		{"two files", []string{"kind", "capture", "a.json", "b.json"}},
		{"an unknown flag", []string{"kind", "capture", "--stdin", "--lenient"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := invokeWithInput(t, "{}", tc.args...)
			if body := got.failure(t); got.code != 2 || body["code"] != "usage_invalid" {
				t.Fatalf("exit %d with %v, want a usage refusal", got.code, body)
			}
		})
	}
}

func TestKindCaptureRefusesWhatIsNoBirthOrNoHonestBirthAndWritesNothing(t *testing.T) {
	t.Parallel()
	birth := birthInJournal(t, pumpNodeJournal(t))
	sealed := `{"stream":"director","seq":2,"type":"wire","at_ns":5,"direction":"node","record":` + birth + `}`
	cases := []struct {
		name, input, keyword string
	}{
		{"input without a birth", `{"type":"DDATA"}`, "no_birth"},
		{"a birth of the sealed stream", sealed, "sealed_stream"},
		{"a birth whose kind changed after its digest", strings.Replace(birth, `"m3/h"`, `"gal/h"`, 1), "kind_digest_mismatch"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := invokeWithInput(t, tc.input, "kind", "capture", "--stdin")
			body := got.failure(t)
			issues, _ := body["issues"].([]any)
			if got.code != 1 || body["code"] != "kind_failed" || body["source"] != "<stdin>" || len(issues) == 0 || issues[0].(map[string]any)["keyword"] != tc.keyword {
				t.Fatalf("exit %d with %v, want kind_failed from <stdin>", got.code, body)
			}
		})
	}
	missing := invoke(t, "kind", "capture", filepath.Join(t.TempDir(), "absent.jsonl"))
	if body := missing.failure(t); missing.code != 1 || body["code"] != "file_not_found" {
		t.Fatalf("exit %d with %v, want file_not_found", missing.code, body)
	}
}
