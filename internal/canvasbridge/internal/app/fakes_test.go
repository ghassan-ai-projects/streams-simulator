package app

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/files"
	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/process"
)

type fakeDirector struct {
	calls    []string
	results  map[string]process.ToolResult
	closed   bool
	lastArgs map[string]any
	failWith error
	entered  chan struct{}
	release  chan struct{}
}

func (d *fakeDirector) Call(_ context.Context, tool string, args map[string]any) (process.ToolResult, error) {
	d.calls = append(d.calls, tool)
	d.lastArgs = args
	if d.release != nil {
		d.entered <- struct{}{}
		<-d.release
	}
	if failure := d.failWith; failure != nil {
		d.failWith = nil
		return process.ToolResult{}, failure
	}
	if result, ok := d.results[tool]; ok {
		return result, nil
	}
	return process.ToolResult{Structured: json.RawMessage(`{"ok":true}`)}, nil
}

func (d *fakeDirector) Close() error {
	d.closed = true
	return nil
}

type harness struct {
	bridge    *Bridge
	dir       string
	directors []*fakeDirector
	argv      [][]string
	stdin     [][]byte
	command   process.CommandResult
	ticks     chan time.Time
	prepare   func(*fakeDirector)
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	h := &harness{dir: t.TempDir(), ticks: make(chan time.Time), command: process.CommandResult{Stdout: []byte(`{"valid":true}`)}}
	workspace, err := files.Open(h.dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = workspace.Close() })
	if err := workspace.Ensure(workspaceFolders); err != nil {
		t.Fatal(err)
	}
	h.bridge = New(Edges{Workspace: workspace, NewSession: func() string { return "s-1" }, Tick: h.tick,
		StartDirector: h.startDirector, RunCommand: h.runCommand}, Settings{WorkspaceDir: h.dir, AdaptersDir: "/adapter-files",
		ComponentDirs: []string{"/lib"}, PairingCode: "ABCDEFGHJKLM"})
	return h
}

func inDomains(name string) string {
	return domain.DefinitionFolders()[0] + "/" + name
}

func (h *harness) domainsDir() string {
	return filepath.Join(h.dir, domain.DefinitionFolders()[0])
}

func (h *harness) tick(time.Duration) (<-chan time.Time, func()) { return h.ticks, func() {} }

func (h *harness) startDirector(context.Context) (Director, error) {
	director := &fakeDirector{results: map[string]process.ToolResult{}}
	if h.prepare != nil {
		h.prepare(director)
	}
	h.directors = append(h.directors, director)
	return director, nil
}

func (h *harness) runCommand(_ context.Context, argv []string, stdin []byte) (process.CommandResult, error) {
	h.argv = append(h.argv, argv)
	h.stdin = append(h.stdin, stdin)
	return h.command, nil
}

func (h *harness) replace(t *testing.T, rel, content string) {
	t.Helper()
	staged := filepath.Join(t.TempDir(), "staged")
	if err := os.WriteFile(staged, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(h.dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(staged, path); err != nil {
		t.Fatal(err)
	}
}

func (h *harness) write(t *testing.T, rel, content string) {
	t.Helper()
	path := filepath.Join(h.dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}
