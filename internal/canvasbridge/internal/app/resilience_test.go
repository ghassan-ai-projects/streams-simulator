package app

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/process"
)

func TestADeadDirectorIsReplacedAndTheCallRetriedOnce(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	ctx := context.Background()
	h.prepare = func(d *fakeDirector) {
		d.results["sim.world.create"] = process.ToolResult{Structured: []byte(`{"world_id":"w-1"}`)}
	}
	if _, err := h.bridge.RunAction(ctx, "world.create", map[string]any{"domain": "plant"}); err != nil || !h.bridge.worldsOpen() {
		t.Fatalf("world.create: %v", err)
	}
	h.directors[0].failWith = fmt.Errorf("%w: sim.catalog.list: broken pipe", process.ErrGone)
	if _, err := h.bridge.RunAction(ctx, "catalog.list", nil); err != nil {
		t.Fatalf("the call after the director died: %v", err)
	}
	if len(h.directors) != 2 || !h.directors[0].closed || len(h.directors[1].calls) != 1 {
		t.Fatalf("%d directors, first closed %v: the dead director must be replaced and the call retried on the new one", len(h.directors), h.directors[0].closed)
	}
	if h.bridge.worldsOpen() || h.bridge.runIsRunning("w-1") {
		t.Fatal("the worlds of the dead director are still counted as open")
	}
}

func TestADirectorThatKeepsFailingFailsTheCallWithoutLooping(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	gone := fmt.Errorf("%w: sim.clock.state: broken pipe; the director said: boom", process.ErrGone)
	h.prepare = func(d *fakeDirector) { d.failWith = gone }
	_, err := h.bridge.RunAction(context.Background(), "clock.state", map[string]any{"world_id": "w-1"})
	if !errors.Is(err, process.ErrGone) || !strings.Contains(err.Error(), "boom") || len(h.directors) != 2 {
		t.Fatalf("err = %v after %d directors, want the director's last words after one retry", err, len(h.directors))
	}
}

func TestPairingFilesAndAdmissionAreNotBlockedByALongDirectorCall(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	entered, release := make(chan struct{}), make(chan struct{})
	h.prepare = func(d *fakeDirector) { d.entered, d.release = entered, release }
	done := make(chan error, 1)
	go func() {
		_, err := h.bridge.RunAction(context.Background(), "clock.advance", map[string]any{"world_id": "w-1", "by_ns": 1.0})
		done <- err
	}()
	<-entered
	served := make(chan struct{})
	go func() {
		_ = h.bridge.Admits("any")
		_, _ = h.bridge.Pair("WRONG")
		_, _ = h.bridge.WriteFile(plant, "{}", "")
		_, _ = h.bridge.Workspace()
		close(served)
	}()
	select {
	case <-served:
	case <-time.After(10 * time.Second):
		t.Fatal("cookie checks, pairing, file saves and listings wait behind a running director call")
	}
	close(release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}

func TestACommandThatWouldOverwriteAFileNeedsTheVersionItStartedFrom(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	ctx := context.Background()
	h.write(t, "compositions/f.composition.json", "{}")
	h.write(t, inDomains("f.domain.json"), "old")
	args := map[string]any{"composition": "compositions/f.composition.json", "out": inDomains("f.domain.json")}
	var conflict *domain.ConflictError
	_, err := h.bridge.RunAction(ctx, "domain.compose", args)
	if !errors.As(err, &conflict) || !errors.Is(err, domain.ErrConflict) || conflict.Etag != domain.Version([]byte("old")) || len(h.argv) != 0 {
		t.Fatalf("composing over an existing file without its version: %v, ran %d commands", err, len(h.argv))
	}
	args["etag"] = "stale"
	if _, err := h.bridge.RunAction(ctx, "domain.compose", args); !errors.Is(err, domain.ErrConflict) || len(h.argv) != 0 {
		t.Fatalf("composing over a file that changed: %v", err)
	}
	args["etag"] = domain.Version([]byte("old"))
	if _, err := h.bridge.RunAction(ctx, "domain.compose", args); err != nil || len(h.argv) != 1 {
		t.Fatalf("composing with the version it started from: %v, ran %d commands", err, len(h.argv))
	}
	fresh := map[string]any{"composition": "compositions/f.composition.json", "out": inDomains("g.domain.json")}
	if _, err := h.bridge.RunAction(ctx, "domain.compose", fresh); err != nil || len(h.argv) != 2 {
		t.Fatalf("composing a new file: %v", err)
	}
	args["etag"] = 3.0
	if _, err := h.bridge.RunAction(ctx, "domain.compose", args); !errors.Is(err, domain.ErrBadArguments) {
		t.Fatalf("an etag that is not text: %v", err)
	}
}

func TestScenarioExportIsGuardedLikeComposition(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.write(t, "scenarios/s.scenario.json", "old")
	_, err := h.bridge.RunAction(context.Background(), "scenario.export", map[string]any{"run": "w-1", "out": "scenarios/s.scenario.json"})
	if !errors.Is(err, domain.ErrConflict) || len(h.argv) != 0 {
		t.Fatalf("exporting over an existing scenario: %v", err)
	}
}

func TestNoErrorOrResultNamesTheAbsoluteWorkspacePath(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	ctx := context.Background()
	h.command = process.CommandResult{ExitCode: 1, Stderr: []byte("open " + h.dir + "/" + inDomains("x.domain.json") + ": denied")}
	_, err := h.bridge.RunAction(ctx, "domain.validate", map[string]any{"path": plant})
	if err == nil || strings.Contains(err.Error(), h.dir) || !strings.Contains(err.Error(), "open "+inDomains("x.domain.json")+": denied") {
		t.Fatalf("a refused command: %v", err)
	}
	h.command = process.CommandResult{Stdout: []byte(`{"composed":"` + h.dir + "/" + inDomains("x.domain.json") + `"}`)}
	result, err := h.bridge.RunAction(ctx, "domain.validate", map[string]any{"path": plant})
	if err != nil || strings.Contains(string(result), h.dir) || !strings.Contains(string(result), `"composed":"`+inDomains("x.domain.json")+`"`) {
		t.Fatalf("a result: %s, %v", result, err)
	}
	h.prepare = func(d *fakeDirector) {
		d.results["sim.clock.state"] = process.ToolResult{Failure: "domain_invalid: " + h.dir + "/" + inDomains("p.domain.json") + " is bad"}
	}
	_, err = h.bridge.RunAction(ctx, "clock.state", map[string]any{"world_id": "w-1"})
	var failure *domain.Failure
	if !errors.As(err, &failure) || strings.Contains(failure.Message, h.dir) {
		t.Fatalf("a tool failure: %v", err)
	}
}

func TestAFileLargerThanAReadMayBeIsNotWritten(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	if _, err := h.bridge.WriteFile(plant, strings.Repeat("a", domain.MaxFileBytes+1), ""); !errors.Is(err, domain.ErrTooLarge) {
		t.Fatalf("a write over the read limit: %v", err)
	}
}
