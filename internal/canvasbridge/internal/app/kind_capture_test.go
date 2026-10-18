package app

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/process"
)

const (
	birthText     = `{"type":"NBIRTH","seq":0}`
	runJournal    = "runs/w-1/journal.jsonl"
	sealedJournal = "runs/w-1/sealed/journal.jsonl"
)

func capture(h *harness, args map[string]any) (json.RawMessage, error) {
	return h.bridge.RunAction(context.Background(), "kind.capture", args)
}

func TestKindCaptureSendsTheTextOnStandardInputWithTheWorkspaceKindsAndWritesNothing(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.write(t, fanKind, fanKindBytes)
	before := workspaceFiles(t, h.dir)
	result, err := capture(h, map[string]any{"content": birthText})
	if err != nil || string(result) != `{"valid":true}` {
		t.Fatalf("kind.capture = %s, %v", result, err)
	}
	want := []string{"kind", "capture", "--stdin", "--kinds-dir", filepath.Join(h.dir, "kinds")}
	if !slices.Equal(h.argv[0], want) || string(h.stdin[0]) != birthText {
		t.Fatalf("argv %v, stdin %q, want %v and the text verbatim", h.argv[0], h.stdin[0], want)
	}
	if after := workspaceFiles(t, h.dir); !slices.Equal(before, after) {
		t.Fatalf("workspace files %v, want %v: a capture writes nothing", after, before)
	}
}

func TestKindCaptureOfARunReadsItsObserverJournalNeverAFreePathAndSendsNoInput(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.write(t, fanKind, fanKindBytes)
	h.write(t, runJournal, birthText)
	h.write(t, sealedJournal, birthText)
	if _, err := capture(h, map[string]any{"run": "w-1"}); err != nil {
		t.Fatal(err)
	}
	want := []string{"kind", "capture", filepath.Join(h.dir, runJournal), "--kinds-dir", filepath.Join(h.dir, "kinds")}
	if !slices.Equal(h.argv[0], want) || h.stdin[0] != nil {
		t.Fatalf("argv %v stdin %q, want %v and no input", h.argv[0], h.stdin[0], want)
	}
	for _, word := range h.argv[0] {
		if strings.Contains(word, "sealed") {
			t.Fatalf("argv %v names the sealed journal", h.argv[0])
		}
	}
}

func TestKindCaptureComparesNothingWhenTheWorkspaceHasNoKindsFolder(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	if _, err := capture(h, map[string]any{"content": birthText}); err != nil {
		t.Fatal(err)
	}
	if want := []string{"kind", "capture", "--stdin"}; !slices.Equal(h.argv[0], want) {
		t.Fatalf("argv %v, want %v without a kinds directory", h.argv[0], want)
	}
}

func TestKindCaptureRefusesContentOverTheLimitWithoutRunningAnything(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	atLimit := strings.Repeat(" ", domain.MaxInputBytes)
	if _, err := capture(h, map[string]any{"content": atLimit}); err != nil {
		t.Fatalf("content of exactly %d bytes = %v, want it passed on", domain.MaxInputBytes, err)
	}
	_, err := capture(h, map[string]any{"content": atLimit + " "})
	if !errors.Is(err, domain.ErrTooLarge) || len(h.argv) != 1 {
		t.Fatalf("content over the limit = %v with %d commands run, want ErrTooLarge and no further command", err, len(h.argv))
	}
}

func TestKindCaptureRelaysTheCommandsRefusalOfAMalformedOrDishonestBirth(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.command = process.CommandResult{ExitCode: 1, Stdout: []byte(`{"error":{"code":"kind_failed","message":"the birth is refused","source":"<stdin>","issues":[{"pointer":"/kinds/0/kind_digest","message":"claim","keyword":"kind_digest_mismatch","member":"kind_digest"}]}}`)}
	_, err := capture(h, map[string]any{"content": "{"})
	var failure *domain.Failure
	if !errors.As(err, &failure) || failure.Code != "kind_failed" || !strings.Contains(string(failure.Issues), `"keyword":"kind_digest_mismatch"`) || string(h.stdin[0]) != "{" {
		t.Fatalf("kind.capture of a malformed birth = %v, want the command's kind_failed with its issues, and the bytes on stdin", err)
	}
}

func TestKindCaptureNeverTakesARunIdThatLooksLikeAPathAndNeverTheSealedJournal(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.write(t, sealedJournal, birthText)
	for _, run := range []string{"../w-1", "w-1/sealed/journal.jsonl", "w-1/sealed", "/etc/passwd", "w-1/../w-2", "..", ".", "", "-w", "w-1\x00"} {
		if _, err := capture(h, map[string]any{"run": run}); !errors.Is(err, domain.ErrBadArguments) {
			t.Errorf("run %q = %v, want bad arguments", run, err)
		}
	}
	for name, args := range map[string]map[string]any{
		"the sealed journal as a path": {"path": sealedJournal},
		"a path beside a run":          {"run": "w-1", "path": runJournal},
		"a run as a number":            {"run": 1.0},
		"both sources":                 {"run": "w-1", "content": birthText},
		"no source":                    {},
		"an etag":                      {"run": "w-1", "etag": ""},
	} {
		if _, err := capture(h, args); !errors.Is(err, domain.ErrBadArguments) {
			t.Errorf("%s: kind.capture = %v, want bad arguments", name, err)
		}
	}
	if len(h.argv) != 0 {
		t.Fatalf("commands run for refused arguments: %v", h.argv)
	}
}

func TestKindCaptureKeepsConcurrentContentsApart(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.echoStdin()
	var group sync.WaitGroup
	for i := range 32 {
		group.Add(1)
		go func() {
			defer group.Done()
			content := `{"type":"NBIRTH","bd_seq":` + strconv.Itoa(i) + `}`
			result, err := capture(h, map[string]any{"content": content})
			if want := `{"echo":` + strconv.Quote(content) + `}`; err != nil || string(result) != want {
				t.Errorf("call %d = %s, %v, want %s", i, result, err, want)
			}
		}()
	}
	group.Wait()
}
