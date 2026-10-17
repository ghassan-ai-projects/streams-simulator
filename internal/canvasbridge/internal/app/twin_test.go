package app

import (
	"context"
	"errors"
	"path/filepath"
	"slices"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/domain"
)

const (
	fanChoices = "twins/fan.twin.json"
	twinResult = `{"id":"fan-twin","implements":["fan"]}`
)

func TestKindTwinRunsTheKindCommandOnAWorkspaceChoicesFileAndWritesADomain(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.write(t, fanKind, fanKindBytes)
	h.write(t, fanChoices, "{}")
	h.command.Stdout = []byte(twinResult)
	args := map[string]any{"choices": fanChoices, "out": inDomains("fan.domain.json")}
	result, err := h.bridge.RunAction(context.Background(), "kind.twin", args)
	want := []string{"kind", "twin", "--choices", h.dir + "/" + fanChoices, "--out", h.dir + "/" + inDomains("fan.domain.json"),
		"--components-dir", "/lib", "--kinds-dir", filepath.Join(h.dir, "kinds")}
	if err != nil || string(result) != twinResult || !slices.Equal(h.argv[0], want) {
		t.Fatalf("kind.twin = %s, %v with argv %v, want %v", result, err, h.argv, want)
	}
}

func TestKindTwinWithoutAKindsFolderPassesNoKindsDirectory(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	if _, err := h.bridge.RunAction(context.Background(), "kind.twin", map[string]any{"choices": fanChoices, "out": inDomains("fan.domain.json")}); err != nil {
		t.Fatal(err)
	}
	if slices.Contains(h.argv[0], "--kinds-dir") {
		t.Fatalf("argv = %v, want no kinds directory before the workspace has one", h.argv[0])
	}
}

func TestKindTwinRefusesToReplaceADomainWhoseVersionTheCallerDoesNotHold(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.write(t, inDomains("fan.domain.json"), `{"id":"hand-edited"}`)
	_, err := h.bridge.RunAction(context.Background(), "kind.twin", map[string]any{"choices": fanChoices, "out": inDomains("fan.domain.json"), "etag": ""})
	var conflict *domain.ConflictError
	if !errors.As(err, &conflict) || conflict.Path != inDomains("fan.domain.json") || len(h.argv) != 0 {
		t.Fatalf("kind.twin over an existing domain = %v with argv %v, want a version conflict before the command runs", err, h.argv)
	}
}

func TestTheDirectorIsStartedWithTheWorkspaceKindsFolderOnceItExists(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	if args := h.bridge.DirectorArguments(); slices.Contains(args, "--kinds-dir") {
		t.Fatalf("director arguments = %v, want no kinds directory before the workspace has one", args)
	}
	h.write(t, fanKind, fanKindBytes)
	args := h.bridge.DirectorArguments()
	index := slices.Index(args, "--kinds-dir")
	if index < 0 || args[index+1] != filepath.Join(h.dir, "kinds") {
		t.Fatalf("director arguments = %v, want the workspace kinds folder", args)
	}
}
