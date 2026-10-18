package app

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
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
	taken, _ := plain["taken"].(map[string]any)
	if _, digested := draft["kind_digest"]; plain["environment"] != "simulated" || plain["bd_seq"] != float64(0) || digested || pump["comparison"] != nil ||
		plain["births_seen"] != float64(1) || taken["bd_seq"] != float64(0) || taken["line"].(float64) < 1 || plain["later_births"] != nil {
		t.Fatalf("kind capture = %v, want the draft of pump without kind_digest or comparison from the one simulated birth, with its provenance", plain)
	}
	boards, _ := plain["boards"].([]any)
	for _, board := range boards {
		if fields, _ := board.(map[string]any); fields["board"] == "" || fields["environment"] != "simulated" {
			t.Fatalf("board = %v, want it echoed with the environment it claims, simulated", board)
		}
	}
	if len(boards) == 0 {
		t.Fatalf("capture = %v, want the boards of the simulated node echoed", plain)
	}
	if claims, _ := pump["claims"].([]any); len(claims) == 0 {
		t.Fatalf("pump = %v, want the members the board claimed listed", pump)
	}
	kinds := pumpKindsDirectory(t, asWritten)
	same := invoke(t, "kind", "capture", journal, "--kinds-dir", kinds).mustSucceed(t).document(t, "kind_capture")
	if comparison, _ := capturedKind(t, same, "pump")["comparison"].(map[string]any); comparison["status"] != "same" || comparison["file_digest"] != pump["kind_digest"] ||
		comparison["target"] != filepath.Join(kinds, "pump.kind.json") || comparison["file"] != comparison["target"] {
		t.Fatalf("comparison = %v, want pump the same as its kind file, aimed at it", comparison)
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
	fromFile := invoke(t, "kind", "capture", journal).mustSucceed(t).document(t, "kind_capture")
	fileLine := fromFile["taken"].(map[string]any)["line"].(float64)
	if fileLine < 2 {
		t.Fatalf("the birth stands at line %v of the journal, want a line after the first", fileLine)
	}
	cases := []struct {
		name, input string
		line        float64
	}{{"the birth", birthInJournal(t, journal), 1}, {"the journal", string(raw), fileLine}}
	for _, tc := range cases {
		got := invokeWithInput(t, tc.input, "kind", "capture", "--stdin").mustSucceed(t).document(t, "kind_capture")
		taken, _ := got["taken"].(map[string]any)
		if taken["line"] != tc.line {
			t.Errorf("capture of %s from standard input took the birth at line %v, want %v", tc.name, taken["line"], tc.line)
		}
		delete(got, "taken")
		wantDoc := map[string]any{}
		for key, value := range fromFile {
			wantDoc[key] = value
		}
		delete(wantDoc, "taken")
		if !reflect.DeepEqual(got, wantDoc) {
			t.Errorf("capture of %s from standard input differs from the capture of the file beyond where the birth stood:\n%v\n%v", tc.name, got, wantDoc)
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
	dir := t.TempDir()
	testsupport.WriteFileIn(t, dir, "birth.json", []byte(birth))
	link := filepath.Join(dir, "link.json")
	if err := os.Symlink(filepath.Join(dir, "birth.json"), link); err != nil {
		t.Fatal(err)
	}
	linked := invoke(t, "kind", "capture", link)
	if body := linked.failure(t); linked.code != 1 || body["code"] != "kind_failed" || !strings.Contains(body["message"].(string), "not a regular file") {
		t.Fatalf("exit %d with %v, want a symbolic link refused as no regular file", linked.code, body)
	}
	missing := invoke(t, "kind", "capture", filepath.Join(t.TempDir(), "absent.jsonl"))
	if body := missing.failure(t); missing.code != 1 || body["code"] != "file_not_found" {
		t.Fatalf("exit %d with %v, want file_not_found", missing.code, body)
	}
}

func TestKindCaptureTakesAMissingKindsDirectoryForNoSavedKindsAndCreatesNothing(t *testing.T) {
	t.Parallel()
	missing := filepath.Join(t.TempDir(), "kinds")
	doc := invoke(t, "kind", "capture", pumpNodeJournal(t), "--kinds-dir", missing).mustSucceed(t).document(t, "kind_capture")
	comparison, _ := capturedKind(t, doc, "pump")["comparison"].(map[string]any)
	if comparison["status"] != "new" || comparison["target"] != filepath.Join(missing, "pump.kind.json") || comparison["file"] != nil {
		t.Fatalf("comparison = %v, want pump new and aimed at a file of the missing directory", comparison)
	}
	if _, err := os.Stat(missing); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("stat = %v, want nothing created", err)
	}
}
