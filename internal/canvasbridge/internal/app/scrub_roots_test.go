package app

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/process"
)

func TestASymlinkedWorkspaceIsScrubbedUnderBothItsNames(t *testing.T) {
	t.Parallel()
	real := t.TempDir()
	link := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(real, link); err != nil {
		t.Fatal(err)
	}
	settings, err := settingsOf(Config{Workspace: link, AdaptersDir: "adapters"})
	if err != nil {
		t.Fatal(err)
	}
	resolved, err := filepath.EvalSymlinks(real)
	if err != nil {
		t.Fatal(err)
	}
	if roots := (&Bridge{settings: settings}).workspaceRoots(); len(roots) != 2 || roots[0] != link || roots[1] != resolved {
		t.Fatalf("workspaceRoots = %q, want [%q %q]", roots, link, resolved)
	}
}

func TestAnErrorNamingTheResolvedWorkspaceShowsOnlyRelativePaths(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.bridge.settings.WorkspaceAliases = []string{h.dir + "-resolved"}
	h.command = process.CommandResult{ExitCode: 1, Stderr: []byte("dial unix " + h.dir + "-resolved/sockets/absent.sock: no such file")}
	_, err := h.bridge.RunAction(context.Background(), "domain.validate", map[string]any{"path": plant})
	if err == nil || !strings.Contains(err.Error(), "dial unix sockets/absent.sock: no such file") || strings.Contains(err.Error(), h.dir) {
		t.Fatalf("a refused command naming the resolved workspace: %v", err)
	}
}
