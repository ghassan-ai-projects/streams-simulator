package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/process"
)

var checkArgv = []string{"kind", "validate", "--stdin"}

func workspaceFiles(t *testing.T, dir string) []string {
	t.Helper()
	var found []string
	err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err == nil && !entry.IsDir() {
			found = append(found, path)
		}
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	return found
}

func TestKindCheckSendsTheContentOnStandardInputAndWritesNothing(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	before := workspaceFiles(t, h.dir)
	result, err := h.bridge.RunAction(context.Background(), "kind.check", map[string]any{"content": fanKindBytes})
	if err != nil || string(result) != `{"valid":true}` {
		t.Fatalf("kind.check = %s, %v", result, err)
	}
	if !slices.Equal(h.argv[0], checkArgv) || string(h.stdin[0]) != fanKindBytes {
		t.Fatalf("argv %v, stdin %q, want %v and the content verbatim", h.argv[0], h.stdin[0], checkArgv)
	}
	if after := workspaceFiles(t, h.dir); !slices.Equal(before, after) {
		t.Fatalf("workspace files %v, want %v: a check writes nothing", after, before)
	}
}

func TestKindCheckAddsTheTwinRequirementsOnlyWhenAsked(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	for _, twin := range []any{true, false} {
		if _, err := h.bridge.RunAction(context.Background(), "kind.check", map[string]any{"content": fanKindBytes, "twin": twin}); err != nil {
			t.Fatal(err)
		}
	}
	if !slices.Equal(h.argv[0], append(slices.Clone(checkArgv), "--twin")) || !slices.Equal(h.argv[1], checkArgv) {
		t.Fatalf("argv = %v, want --twin only for the first call", h.argv)
	}
	if _, err := h.bridge.RunAction(context.Background(), "kind.check", map[string]any{"content": fanKindBytes, "twin": "yes"}); !errors.Is(err, domain.ErrBadArguments) {
		t.Fatalf("a twin flag that is no boolean = %v, want bad arguments", err)
	}
}

func TestKindCheckRefusesContentOverTheLimitWithoutRunningAnything(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	atLimit := strings.Repeat(" ", domain.MaxInputBytes)
	if _, err := h.bridge.RunAction(context.Background(), "kind.check", map[string]any{"content": atLimit}); err != nil {
		t.Fatalf("content of exactly %d bytes = %v, want it passed on", domain.MaxInputBytes, err)
	}
	_, err := h.bridge.RunAction(context.Background(), "kind.check", map[string]any{"content": atLimit + " "})
	if !errors.Is(err, domain.ErrTooLarge) || len(h.argv) != 1 {
		t.Fatalf("content over the limit = %v with %d commands run, want ErrTooLarge and no further command", err, len(h.argv))
	}
}

func TestKindCheckRelaysTheCommandsRefusalOfMalformedContent(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.command = process.CommandResult{ExitCode: 1, Stdout: []byte(`{"error":{"code":"kind_failed","message":"not JSON","source":"<stdin>","issues":[{"message":"unexpected end of JSON input","keyword":"json"}]}}`)}
	_, err := h.bridge.RunAction(context.Background(), "kind.check", map[string]any{"content": "{"})
	var failure *domain.Failure
	if !errors.As(err, &failure) || failure.Code != "kind_failed" || !strings.Contains(string(failure.Issues), `"keyword":"json"`) || string(h.stdin[0]) != "{" {
		t.Fatalf("kind.check of malformed content = %v, want the command's kind_failed with its issues, and the bytes on stdin", err)
	}
}

func TestKindCheckTakesNoPathAndNoArgumentBeyondItsOwn(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	cases := map[string]map[string]any{
		"a path":              {"content": fanKindBytes, "path": fanKind},
		"an out file":         {"content": fanKindBytes, "out": inDomains("x.domain.json")},
		"an etag":             {"content": fanKindBytes, "etag": ""},
		"a flag as a name":    {"content": fanKindBytes, "--stdin": true},
		"no content":          {"twin": true},
		"content that is not": {"content": 7},
		"content as a list":   {"content": []any{fanKindBytes}},
	}
	for name, args := range cases {
		if _, err := h.bridge.RunAction(context.Background(), "kind.check", args); !errors.Is(err, domain.ErrBadArguments) {
			t.Errorf("%s: kind.check = %v, want bad arguments", name, err)
		}
	}
	if len(h.argv) != 0 {
		t.Fatalf("commands run for refused arguments: %v", h.argv)
	}
}

func TestKindCheckNeverReadsContentThatLooksLikeAPath(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.write(t, fanKind, fanKindBytes)
	for _, content := range []string{fanKind, "../../etc/passwd", "/etc/passwd", "-rf", h.dir + "/" + fanKind} {
		if _, err := h.bridge.RunAction(context.Background(), "kind.check", map[string]any{"content": content}); err != nil {
			t.Fatalf("content %q = %v", content, err)
		}
	}
	for i, argv := range h.argv {
		if !slices.Equal(argv, checkArgv) || string(h.stdin[i]) == "" {
			t.Errorf("call %d: argv %v stdin %q, want only %v and the text on stdin", i, argv, h.stdin[i], checkArgv)
		}
	}
}

func TestKindCheckKeepsConcurrentContentsApart(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	var mu sync.Mutex
	h.bridge.edges.RunCommand = func(_ context.Context, _ []string, stdin []byte) (process.CommandResult, error) {
		mu.Lock()
		defer mu.Unlock()
		echoed, err := json.Marshal(map[string]string{"echo": string(stdin)})
		if err != nil {
			return process.CommandResult{}, fmt.Errorf("marshal echo: %w", err)
		}
		return process.CommandResult{Stdout: echoed}, nil
	}
	var group sync.WaitGroup
	for i := range 32 {
		group.Add(1)
		go func() {
			defer group.Done()
			content := `{"kind":"k` + strconv.Itoa(i) + `"}`
			result, err := h.bridge.RunAction(context.Background(), "kind.check", map[string]any{"content": content})
			if want := `{"echo":` + strconv.Quote(content) + `}`; err != nil || string(result) != want {
				t.Errorf("call %d = %s, %v, want %s", i, result, err, want)
			}
		}()
	}
	group.Wait()
}

func TestKindSuggestRunsTheCommandOnTheWorkspaceKindsAndTheLibraryAndTakesNoArgument(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.write(t, fanKind, fanKindBytes)
	if _, err := h.bridge.RunAction(context.Background(), "kind.suggest", nil); err != nil {
		t.Fatal(err)
	}
	want := []string{"kind", "twin", "--suggest", "--components-dir", "/lib", "--kinds-dir", filepath.Join(h.dir, "kinds")}
	if !slices.Equal(h.argv[0], want) || h.stdin[0] != nil {
		t.Fatalf("argv %v stdin %q, want %v and no input", h.argv[0], h.stdin[0], want)
	}
	for _, args := range []map[string]any{{"path": fanKind}, {"out": inDomains("x.domain.json")}, {"content": "{}"}} {
		if _, err := h.bridge.RunAction(context.Background(), "kind.suggest", args); !errors.Is(err, domain.ErrBadArguments) {
			t.Errorf("kind.suggest %v = %v, want bad arguments", args, err)
		}
	}
}

func TestComponentListFiltersByAUnitAsATextArgument(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	if _, err := h.bridge.RunAction(context.Background(), "component.list", map[string]any{"unit": "Cel"}); err != nil {
		t.Fatal(err)
	}
	want := []string{"component", "list", "--unit", "Cel", "--components-dir", "/lib"}
	if !slices.Equal(h.argv[0], want) {
		t.Fatalf("argv = %v, want %v", h.argv[0], want)
	}
	for _, unit := range []any{"--components-dir", "", 3} {
		if _, err := h.bridge.RunAction(context.Background(), "component.list", map[string]any{"unit": unit}); !errors.Is(err, domain.ErrBadArguments) {
			t.Errorf("unit %v = %v, want bad arguments", unit, err)
		}
	}
}
