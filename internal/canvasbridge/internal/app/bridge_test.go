package app

import (
	"context"
	"errors"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/process"
)

var plant = inDomains("plant.domain.json")

func seedOf(seed uint64) *uint64 { return &seed }

func TestPairingAdmitsTheSessionItIssued(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	if _, err := h.bridge.Pair("WRONG"); !errors.Is(err, domain.ErrWrongCode) {
		t.Fatalf("a wrong code: %v", err)
	}
	session, err := h.bridge.Pair("abcd-efgh-jklm")
	if err != nil || session != "s-1" || !h.bridge.Admits("s-1") || h.bridge.Admits("s-2") {
		t.Fatalf("Pair = %q, %v", session, err)
	}
}

func TestTheWorkspaceListsDefinitionsWithTheirIdsAndRunsWithTheirState(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.write(t, plant, `{"id":"plant","title":"Plant","instances":[{"kind":"a","names":["a-1","a-2"]}]}`)
	h.write(t, inDomains("notes.txt"), "x")
	h.write(t, "compositions/broken.composition.json", "{")
	h.write(t, "runs/w-1/journal.jsonl", `{"type":"run.opened","mode":"explore"}`+"\n")
	h.write(t, "runs/w-2/journal.jsonl", `{"type":"run.opened","mode":"evaluate","start_time_ns":100,"domain":{"id":"plant"}}`+"\n"+
		`{"type":"run.closed","at_ns":1800000000100}`+"\n")
	h.write(t, "runs/w-2/run.json", `{"seed":3,"command_log":[{"op":"fault.inject"}]}`)
	h.write(t, "runs/w-3/other.txt", "x")
	h.write(t, "runs/w-4/journal.jsonl", `{"type":"run.opened","mode":"evaluate","start_time_ns":100,"domain":{"id":"plant"}}`+"\n"+
		`{"type":"run.closed","at_ns":1800000000100,"incomplete":true}`+"\n")
	h.write(t, "runs/w-4/run.json", `{"seed":4,"error":"the world was destroyed before its run ended (run.end was not called)"}`)
	listing, err := h.bridge.Workspace(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for i, run := range listing.Runs {
		if _, err := time.Parse(time.RFC3339, run.RecordedAt); err != nil {
			t.Fatalf("run %d recorded at %q", i, run.RecordedAt)
		}
		listing.Runs[i].RecordedAt = ""
	}
	want := []domain.RunEntry{
		{WorldID: "w-1", Explore: true},
		{WorldID: "w-2", Ended: true, Complete: true, Seed: seedOf(3), Plant: "plant", DurationNs: 1800000000000, Faults: 1},
		{WorldID: "w-4", Ended: true, IncompleteReason: "the world was destroyed before its run ended (run.end was not called)", Seed: seedOf(4), Plant: "plant", DurationNs: 1800000000000},
	}
	if len(listing.Domains) != 1 || listing.Domains[0] != (domain.Definition{Path: plant, ID: "plant", Title: "Plant", Instances: 2}) ||
		listing.Compositions[0].ID != "" || !reflect.DeepEqual(listing.Runs, want) || len(listing.Scenarios) != 0 {
		t.Fatalf("listing = %+v", listing)
	}
}

func TestFilesAreReadAndWrittenThroughTheirRules(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.write(t, "runs/w-1/journal.jsonl", `{"type":"run.opened","mode":"evaluate"}`+"\n")
	h.write(t, "runs/w-1/sealed/journal.jsonl", "truth\n")
	if _, err := h.bridge.ReadFile("runs/w-1/sealed/journal.jsonl"); !errors.Is(err, domain.ErrSealed) {
		t.Fatalf("a sealed file of an open Evaluate run: %v", err)
	}
	if document, err := h.bridge.ReadFile("runs/w-1/journal.jsonl"); err != nil || !strings.Contains(document.Content, "run.opened") {
		t.Fatalf("the observer journal: %+v, %v", document, err)
	}
	written, err := h.bridge.WriteFile(plant, "{}", "")
	if err != nil || written.Etag != domain.Version([]byte("{}")) {
		t.Fatalf("WriteFile = %+v, %v", written, err)
	}
	if _, err := h.bridge.WriteFile(plant, "{ }", "stale"); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("a stale write: %v", err)
	}
	if read, err := h.bridge.ReadFile(plant); err != nil || read.Etag != written.Etag {
		t.Fatalf("ReadFile = %+v, %v", read, err)
	}
	for rel, want := range map[string]error{"runs/w-1/run.json": domain.ErrReadOnly, "../x.domain.json": domain.ErrBadPath} {
		if _, err := h.bridge.WriteFile(rel, "{}", ""); !errors.Is(err, want) {
			t.Errorf("writing %s = %v, want %v", rel, err, want)
		}
	}
	if _, err := h.bridge.ReadFile("nowhere"); !errors.Is(err, domain.ErrBadPath) {
		t.Fatalf("a bad path: %v", err)
	}
}

func TestCommandActionsRunWithAbsoluteWorkspacePathsAndTheirFlags(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	if _, err := h.bridge.RunAction(context.Background(), "run.verify", map[string]any{"run": "w-1"}); err != nil {
		t.Fatal(err)
	}
	want := []string{"verify", h.dir + "/runs/w-1/run.json", "--domains-dir", h.domainsDir(), "--adapters-dir", "/adapter-files"}
	if !slices.Equal(h.argv[0], want) {
		t.Fatalf("argv = %v, want %v", h.argv[0], want)
	}
	if _, err := h.bridge.RunAction(context.Background(), "component.list", nil); err != nil || !slices.Equal(h.argv[1], []string{"component", "list", "--components-dir", "/lib"}) {
		t.Fatalf("component list argv = %v, %v", h.argv[1], err)
	}
	if _, err := h.bridge.RunAction(context.Background(), "component.example", map[string]any{"id": "factory-utilities"}); err != nil ||
		!slices.Equal(h.argv[2], []string{"component", "example", "factory-utilities"}) {
		t.Fatalf("component example argv = %v, %v", h.argv[2], err)
	}
	h.command = process.CommandResult{ExitCode: 1, Stdout: []byte(`{"error":{"code":"domain_invalid","message":"bad","issues":[{"pointer":"/x"}]}}`)}
	var failure *domain.Failure
	if _, err := h.bridge.RunAction(context.Background(), "domain.validate", map[string]any{"path": plant}); !errors.As(err, &failure) ||
		failure.Code != "domain_invalid" || !strings.Contains(string(failure.Issues), "/x") {
		t.Fatalf("a refused command: %v", err)
	}
	h.command = process.CommandResult{Stdout: []byte("not json")}
	if _, err := h.bridge.RunAction(context.Background(), "domain.validate", map[string]any{"path": plant}); !errors.As(err, &failure) {
		t.Fatalf("a command without a JSON document: %v", err)
	}
	if _, err := h.bridge.RunAction(context.Background(), "nope", nil); !errors.Is(err, domain.ErrUnknownAction) {
		t.Fatalf("an unknown action: %v", err)
	}
}

func TestAVerifyThatDoesNotReproduceKeepsItsResultWithoutTheWorkspacePath(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	document := `{"error":{"code":"replay_mismatch","message":"does not reproduce","result":{"matches":false,"detail":"` + h.dir + `/runs/w-1"}}}`
	h.command = process.CommandResult{ExitCode: 1, Stdout: []byte(document)}
	var failure *domain.Failure
	_, err := h.bridge.RunAction(context.Background(), "run.verify", map[string]any{"run": "w-1"})
	if !errors.As(err, &failure) || failure.Code != "replay_mismatch" {
		t.Fatalf("err = %v", err)
	}
	if want := `{"matches":false,"detail":"runs/w-1"}`; string(failure.Result) != want {
		t.Fatalf("result = %s, want %s", failure.Result, want)
	}
}
