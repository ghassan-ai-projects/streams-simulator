package app

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/process"
)

func (r *sessionRig) later(name string, args map[string]any) <-chan domain.SessionStatus {
	done := make(chan domain.SessionStatus, 1)
	go func() {
		var status domain.SessionStatus
		if raw, err := r.bridge.RunAction(context.Background(), name, args); err == nil {
			_ = json.Unmarshal(raw, &status)
		}
		done <- status
	}()
	return done
}

func TestStopEndsTheSessionAndReportsTheRunFolderItSaved(t *testing.T) {
	t.Parallel()
	rig := newSessionRig(t)
	rig.exit = process.LastingExit{Stdout: []byte(`{"out":"` + filepath.Join(rig.dir, "runs", "c-1") + `","stopped":true}`)}
	rig.act(t, "session.start", plan())
	stopped := rig.act(t, "session.stop", nil)
	if stopped.State != domain.SessionEnded || string(stopped.Result) != `{"out":"runs/c-1","stopped":true}` || rig.child(0).stops() != 1 ||
		stopped.Domain != "arduino-thermal" {
		t.Fatalf("stop = %+v (result %s), want ended with the run folder relative to the workspace", stopped, stopped.Result)
	}
	if rig.bridge.runIsRunning("c-1") {
		t.Fatal("a stopped session's run is still counted as running")
	}
	if again := rig.act(t, "session.start", plan()); again.Run != "c-2" || rig.children() != 2 {
		t.Fatalf("a start after the stop = %+v, want a new session c-2", again)
	}
}

func TestStatusDoesNotWaitForAStopInProgress(t *testing.T) {
	t.Parallel()
	rig := newSessionRig(t)
	rig.act(t, "session.start", plan())
	child := rig.child(0)
	child.entered, child.release = make(chan struct{}), make(chan struct{})
	done := rig.later("session.stop", nil)
	<-child.entered
	if status := rig.act(t, "session.status", nil); status.State != domain.SessionStopping {
		t.Fatalf("status during the stop = %+v, want stopping at once", status)
	}
	if again := rig.act(t, "session.stop", nil); again.State != domain.SessionStopping {
		t.Fatalf("a second stop = %+v, want stopping without a second interrupt", again)
	}
	close(child.release)
	if stopped := <-done; stopped.State != domain.SessionEnded || child.stops() != 1 {
		t.Fatalf("stop = %+v after %d stops, want ended after one", stopped, child.stops())
	}
}

func TestAStopWhileTheSessionStartsStopsTheChildAsSoonAsItRuns(t *testing.T) {
	t.Parallel()
	rig := newSessionRig(t)
	rig.launched, rig.proceed = make(chan struct{}), make(chan struct{})
	done := rig.later("session.start", plan())
	<-rig.launched
	if status := rig.act(t, "session.status", nil); status.State != domain.SessionStarting || status.Run != "c-1" {
		t.Fatalf("status while starting = %+v, want starting", status)
	}
	if stopping := rig.act(t, "session.stop", nil); stopping.State != domain.SessionStarting {
		t.Fatalf("stop while starting = %+v, want the request taken while the child is not up yet", stopping)
	}
	close(rig.proceed)
	if started := <-done; started.State != domain.SessionEnded || rig.child(0).stops() != 1 {
		t.Fatalf("start = %+v, want the child stopped as soon as it ran", started)
	}
}

func TestASessionThatDiesIsReportedFailedWithTheCommandsError(t *testing.T) {
	t.Parallel()
	rig := newSessionRig(t)
	rig.act(t, "session.start", plan())
	rig.child(0).die(process.LastingExit{Stdout: []byte(`{"error":{"code":"plantwire_failed","message":"host lost"}}`), ExitCode: 1})
	if status := rig.act(t, "session.status", nil); status.State != domain.SessionFailed || status.Failure.Code != "plantwire_failed" {
		t.Fatalf("status = %+v, want failed with plantwire_failed", status)
	}
}

func TestAStartWaitsUntilTheSessionOpensItsRunOrDies(t *testing.T) {
	t.Parallel()
	rig := newSessionRig(t)
	rig.bridge.edges.StartLasting = func([]string, []byte) (Lasting, error) {
		return &fakeLasting{exited: true, exit: process.LastingExit{Stderr: "no domain arduino-thermal", ExitCode: 1}}, nil
	}
	if started := rig.act(t, "session.start", plan()); started.State != domain.SessionFailed || started.Failure.Message != "no domain arduino-thermal" {
		t.Fatalf("a start whose child died at once = %+v, want failed with its last words", started)
	}
	rig.bridge.edges.StartLasting = func([]string, []byte) (Lasting, error) { return &fakeLasting{}, nil }
	ticks := 0
	rig.bridge.edges.Tick = func(time.Duration) (<-chan time.Time, func()) {
		always := make(chan time.Time, openAttempts)
		for range openAttempts {
			always <- time.Time{}
		}
		ticks++
		return always, func() {}
	}
	if started := rig.act(t, "session.start", plan()); started.State != domain.SessionRunning || ticks != 1 {
		t.Fatalf("a start whose run never opened = %+v after %d waits, want it reported running once the bounded wait is over", started, ticks)
	}
}

func TestAStartThatCannotLaunchTheChildSaysWhy(t *testing.T) {
	t.Parallel()
	rig := newSessionRig(t)
	rig.bridge.edges.StartLasting = func([]string, []byte) (Lasting, error) { return nil, errors.New("no such executable") }
	if _, err := rig.bridge.RunAction(context.Background(), "session.start", plan()); err == nil || err.Error() != "no such executable" {
		t.Fatalf("a start whose child could not launch = %v, want the launch error", err)
	}
	if status := rig.act(t, "session.status", nil); status.State != domain.SessionFailed || status.Failure.Message != "no such executable" {
		t.Fatalf("status after a failed launch = %+v, want failed with the launch error", status)
	}
}

func TestARunIDTakenOnDiskIsSkippedAndOnlyEndlessCollisionsAreRefused(t *testing.T) {
	t.Parallel()
	rig := newSessionRig(t)
	rig.write(t, "runs/c-1/journal.jsonl", "{}\n")
	if started := rig.act(t, "session.start", plan()); started.Run != "c-2" {
		t.Fatalf("start next to an existing c-1 = %+v, want a fresh id c-2", started)
	}
	stuck := newSessionRig(t)
	stuck.bridge.edges.NewRunID = func() string { return "c-9" }
	stuck.write(t, "runs/c-9/journal.jsonl", "{}\n")
	if _, err := stuck.bridge.RunAction(context.Background(), "session.start", plan()); !errors.Is(err, domain.ErrConflict) || stuck.children() != 0 {
		t.Fatalf("start when every id is taken = %v, want ErrConflict and no child", err)
	}
	if _, err := rig.bridge.RunAction(context.Background(), "session.start", map[string]any{}); !errors.Is(err, domain.ErrBadArguments) {
		t.Fatalf("start without a plan = %v, want ErrBadArguments", err)
	}
}

func TestTheTruthGateHoldsWhileASessionRuns(t *testing.T) {
	t.Parallel()
	rig := newSessionRig(t)
	rig.act(t, "session.start", plan())
	rig.write(t, "runs/c-1/sealed/journal.jsonl", "{}\n")
	rig.write(t, "runs/c-1/field-trace.jsonl", "{}\n")
	if _, err := rig.bridge.ReadFile("runs/c-1/sealed/journal.jsonl"); !errors.Is(err, domain.ErrSealed) {
		t.Fatalf("the sealed journal of a running session = %v, want ErrSealed", err)
	}
	if document, err := rig.bridge.ReadFile("runs/c-1/field-trace.jsonl"); err != nil || document.Content != "{}\n" {
		t.Fatalf("the field trace of a running session = %+v, %v, want it readable", document, err)
	}
}

func TestClosingTheBridgeStopsALivingSession(t *testing.T) {
	t.Parallel()
	rig := newSessionRig(t)
	rig.act(t, "session.start", plan())
	if err := rig.bridge.Close(); err != nil || rig.child(0).stops() != 1 {
		t.Fatalf("Close = %v, stopped %d times, want the session stopped once", err, rig.child(0).stops())
	}
}
