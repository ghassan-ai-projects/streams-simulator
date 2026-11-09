package app

import (
	"context"
	"errors"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/process"
)

func TestInstallingAnExamplesKindFilesNamesTheWorkspaceKindsFolderEvenBeforeItExists(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.command = process.CommandResult{Stdout: []byte(`{"example":"belt","written":["a.kind.json"],"unchanged":[]}`)}
	result, err := h.bridge.RunAction(context.Background(), "component.install_kinds", map[string]any{"example": "belt"})
	if err != nil || !strings.Contains(string(result), `"written":["a.kind.json"]`) {
		t.Fatalf("install = %s, %v", result, err)
	}
	want := []string{"component", "install-kinds", "belt", "--kinds-dir", filepath.Join(h.dir, "kinds")}
	if !slices.Equal(h.argv[0], want) {
		t.Fatalf("argv = %v, want %v", h.argv[0], want)
	}
}

func TestARefusedKindFileInstallKeepsItsIssuesForTheCanvas(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.command = process.CommandResult{ExitCode: 1, Stdout: []byte(`{"error":{"code":"kind_failed","message":"not written",` +
		`"issues":[{"pointer":"/bulk_feeder.kind.json","keyword":"kind_file_differs","message":"differs"}]}}`)}
	var failure *domain.Failure
	_, err := h.bridge.RunAction(context.Background(), "component.install_kinds", map[string]any{"example": "belt"})
	if !errors.As(err, &failure) || failure.Code != "kind_failed" || !strings.Contains(string(failure.Issues), "kind_file_differs") {
		t.Fatalf("a refused install = %v", err)
	}
	if _, err := h.bridge.RunAction(context.Background(), "component.install_kinds", map[string]any{"example": "--force"}); !errors.Is(err, domain.ErrBadArguments) {
		t.Fatalf("a flag as example = %v", err)
	}
}
