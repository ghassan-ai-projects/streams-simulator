package app

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"slices"
	"strconv"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/process"
)

const sessionPlan = `{"schema_version":"0.1","domain":"arduino-thermal","host_socket":"/tmp/h.sock"}`

type fakeLasting struct {
	exit    process.LastingExit
	exited  bool
	stopped int
}

func (f *fakeLasting) Exited() (process.LastingExit, bool) { return f.exit, f.exited }

func (f *fakeLasting) Stop(time.Duration) process.LastingExit {
	f.stopped++
	f.exited = true
	return f.exit
}

type sessionRig struct {
	*harness
	started      []*fakeLasting
	lastingArgv  [][]string
	lastingStdin [][]byte
	exit         process.LastingExit
}

func newSessionRig(t *testing.T) *sessionRig {
	t.Helper()
	rig := &sessionRig{harness: newHarness(t)}
	runs := 0
	rig.bridge.edges.NewRunID = func() string {
		runs++
		return "c-" + strconv.Itoa(runs)
	}
	rig.bridge.edges.StartLasting = func(argv []string, stdin []byte) (Lasting, error) {
		started := &fakeLasting{exit: rig.exit}
		rig.started, rig.lastingArgv, rig.lastingStdin = append(rig.started, started), append(rig.lastingArgv, argv), append(rig.lastingStdin, stdin)
		out, err := filepath.Rel(rig.dir, argv[len(argv)-1])
		if err != nil {
			t.Fatal(err)
		}
		rig.write(t, filepath.ToSlash(out)+"/journal.jsonl", "{}\n")
		return started, nil
	}
	return rig
}

func TestAStartWaitsUntilTheSessionOpensItsRunOrDies(t *testing.T) {
	t.Parallel()
	rig := newSessionRig(t)
	rig.bridge.edges.StartLasting = func([]string, []byte) (Lasting, error) {
		return &fakeLasting{exited: true, exit: process.LastingExit{Stderr: "no domain arduino-thermal", ExitCode: 1}}, nil
	}
	if started := rig.act(t, "session.start", map[string]any{"content": sessionPlan}); started.State != domain.SessionFailed ||
		started.Failure.Message != "no domain arduino-thermal" {
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
	if started := rig.act(t, "session.start", map[string]any{"content": sessionPlan}); started.State != domain.SessionRunning || ticks != 1 {
		t.Fatalf("a start whose run never opened = %+v after %d waits, want it reported running once the bounded wait is over", started, ticks)
	}
}

func TestAStartThatCannotLaunchTheChildSaysWhy(t *testing.T) {
	t.Parallel()
	rig := newSessionRig(t)
	rig.bridge.edges.StartLasting = func([]string, []byte) (Lasting, error) { return nil, errors.New("no such executable") }
	if _, err := rig.bridge.RunAction(context.Background(), "session.start", map[string]any{"content": sessionPlan}); err == nil ||
		err.Error() != "no such executable" {
		t.Fatalf("a start whose child could not launch = %v, want the launch error", err)
	}
	if status := rig.act(t, "session.status", nil); status.State != domain.SessionNone {
		t.Fatalf("status after a failed launch = %+v, want none", status)
	}
}

func (r *sessionRig) act(t *testing.T, name string, args map[string]any) domain.SessionStatus {
	t.Helper()
	raw, err := r.bridge.RunAction(context.Background(), name, args)
	if err != nil {
		t.Fatalf("%s = %v", name, err)
	}
	var status domain.SessionStatus
	if err := json.Unmarshal(raw, &status); err != nil {
		t.Fatal(err)
	}
	return status
}

func TestAStartedSessionIsAChildTheBridgeOwnsWithItsOwnRunFolder(t *testing.T) {
	t.Parallel()
	rig := newSessionRig(t)
	started := rig.act(t, "session.start", map[string]any{"content": sessionPlan})
	want := []string{"plantwire", "connect", "--stdin", "--domains-dir", rig.domainsDir(), "--adapters-dir", "/adapter-files",
		"--out", filepath.Join(rig.dir, "runs", "c-1")}
	if started.Run != "c-1" || started.State != domain.SessionRunning || !slices.Equal(rig.lastingArgv[0], want) || string(rig.lastingStdin[0]) != sessionPlan {
		t.Fatalf("start = %+v argv %v stdin %q, want run c-1 running as %v with the plan on stdin", started, rig.lastingArgv[0], rig.lastingStdin[0], want)
	}
	if status := rig.act(t, "session.status", nil); status.Run != "c-1" || status.State != domain.SessionRunning || !rig.bridge.runIsRunning("c-1") {
		t.Fatalf("status = %+v, want c-1 running and counted as a running run", status)
	}
	_, err := rig.bridge.RunAction(context.Background(), "session.start", map[string]any{"content": sessionPlan})
	var failure *domain.Failure
	if !errors.As(err, &failure) || failure.Code != "session_running" || len(rig.started) != 1 {
		t.Fatalf("a second start = %v, want session_running and no second child", err)
	}
}

func TestStopEndsTheSessionAndReportsTheRunFolderItSaved(t *testing.T) {
	t.Parallel()
	rig := newSessionRig(t)
	rig.exit = process.LastingExit{Stdout: []byte(`{"out":"` + filepath.Join(rig.dir, "runs", "c-1") + `","stopped":true}`)}
	rig.act(t, "session.start", map[string]any{"content": sessionPlan})
	stopped := rig.act(t, "session.stop", nil)
	if stopped.State != domain.SessionEnded || string(stopped.Result) != `{"out":"runs/c-1","stopped":true}` || rig.started[0].stopped != 1 {
		t.Fatalf("stop = %+v (result %s), want ended with the run folder relative to the workspace", stopped, stopped.Result)
	}
	if rig.bridge.runIsRunning("c-1") {
		t.Fatal("a stopped session's run is still counted as running")
	}
	if again := rig.act(t, "session.start", map[string]any{"content": sessionPlan}); again.Run != "c-2" || len(rig.started) != 2 {
		t.Fatalf("a start after the stop = %+v, want a new session c-2", again)
	}
}

func TestASessionThatDiesIsReportedFailedWithTheCommandsError(t *testing.T) {
	t.Parallel()
	rig := newSessionRig(t)
	rig.act(t, "session.start", map[string]any{"content": sessionPlan})
	rig.started[0].exit = process.LastingExit{Stdout: []byte(`{"error":{"code":"plantwire_failed","message":"host lost"}}`), ExitCode: 1}
	rig.started[0].exited = true
	if status := rig.act(t, "session.status", nil); status.State != domain.SessionFailed || status.Failure.Code != "plantwire_failed" {
		t.Fatalf("status = %+v, want failed with plantwire_failed", status)
	}
}

func TestStatusAndStopWithoutASessionSayThereIsNone(t *testing.T) {
	t.Parallel()
	rig := newSessionRig(t)
	for _, name := range []string{"session.status", "session.stop"} {
		if status := rig.act(t, name, nil); status.State != domain.SessionNone || status.Run != "" {
			t.Fatalf("%s = %+v, want none", name, status)
		}
	}
	if _, err := rig.bridge.RunAction(context.Background(), "session.status", map[string]any{"run": "c-1"}); !errors.Is(err, domain.ErrBadArguments) {
		t.Fatalf("status with an argument = %v, want ErrBadArguments", err)
	}
}

func TestASessionNeverReusesARunFolderAndNeedsAPlan(t *testing.T) {
	t.Parallel()
	rig := newSessionRig(t)
	rig.write(t, "runs/c-1/journal.jsonl", "{}\n")
	if _, err := rig.bridge.RunAction(context.Background(), "session.start", map[string]any{"content": sessionPlan}); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("start onto an existing run folder = %v, want ErrConflict", err)
	}
	if _, err := rig.bridge.RunAction(context.Background(), "session.start", map[string]any{}); !errors.Is(err, domain.ErrBadArguments) || len(rig.started) != 0 {
		t.Fatalf("start without a plan = %v, want ErrBadArguments and no child", err)
	}
}

func TestClosingTheBridgeStopsALivingSession(t *testing.T) {
	t.Parallel()
	rig := newSessionRig(t)
	rig.act(t, "session.start", map[string]any{"content": sessionPlan})
	if err := rig.bridge.Close(); err != nil || rig.started[0].stopped != 1 {
		t.Fatalf("Close = %v, stopped %d times, want the session stopped once", err, rig.started[0].stopped)
	}
}

func TestACheckRunsConnectOnceWithThePlanAndTheWorkspaceKinds(t *testing.T) {
	t.Parallel()
	rig := newSessionRig(t)
	rig.write(t, fanKind, fanKindBytes)
	if _, err := rig.bridge.RunAction(context.Background(), "session.check", map[string]any{"content": sessionPlan}); err != nil {
		t.Fatal(err)
	}
	want := []string{"plantwire", "connect", "--check", "--stdin", "--domains-dir", rig.domainsDir(), "--adapters-dir", "/adapter-files",
		"--kinds-dir", filepath.Join(rig.dir, "kinds")}
	if !slices.Equal(rig.argv[0], want) || string(rig.stdin[0]) != sessionPlan || len(rig.started) != 0 {
		t.Fatalf("check argv %v stdin %q, want %v and no lasting child", rig.argv[0], rig.stdin[0], want)
	}
}
