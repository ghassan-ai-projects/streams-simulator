package app

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"slices"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/process"
)

const (
	sessionPlan  = `{"schema_version":"0.1","domain":"arduino-thermal","host_socket":"/tmp/h.sock"}`
	sessionCheck = `{"domain":"arduino-thermal","environment":"simulated","instances":[],"implements_checked":[]}`
)

type fakeLasting struct {
	mu      sync.Mutex
	exit    process.LastingExit
	exited  bool
	stopped int
	entered chan struct{}
	release chan struct{}
}

func (f *fakeLasting) Exited() (process.LastingExit, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.exit, f.exited
}

func (f *fakeLasting) Stop(time.Duration) process.LastingExit {
	if f.entered != nil {
		f.entered <- struct{}{}
		<-f.release
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.stopped++
	f.exited = true
	return f.exit
}

func (f *fakeLasting) die(exit process.LastingExit) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.exit, f.exited = exit, true
}

func (f *fakeLasting) stops() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.stopped
}

type sessionRig struct {
	*harness
	mu           sync.Mutex
	started      []*fakeLasting
	lastingArgv  [][]string
	lastingStdin [][]byte
	exit         process.LastingExit
	control      *controlFake
	launched     chan struct{}
	proceed      chan struct{}
}

func newSessionRig(t *testing.T) *sessionRig {
	t.Helper()
	rig := &sessionRig{harness: newHarness(t), exit: process.LastingExit{Stdout: []byte(`{"out":"runs/c-1","stopped":true}`)}}
	rig.command = process.CommandResult{Stdout: []byte(sessionCheck)}
	rig.control = newControlFake(rig.bridge)
	runs := 0
	rig.bridge.edges.NewRunID = func() string {
		runs++
		return "c-" + strconv.Itoa(runs)
	}
	rig.bridge.edges.StartLasting = func(argv []string, stdin []byte) (Lasting, error) {
		if rig.launched != nil {
			rig.launched <- struct{}{}
			<-rig.proceed
		}
		return rig.record(t, argv, stdin), nil
	}
	return rig
}

func (r *sessionRig) record(t *testing.T, argv []string, stdin []byte) *fakeLasting {
	t.Helper()
	r.mu.Lock()
	defer r.mu.Unlock()
	started := &fakeLasting{exit: r.exit}
	r.started, r.lastingArgv, r.lastingStdin = append(r.started, started), append(r.lastingArgv, argv), append(r.lastingStdin, stdin)
	out, err := filepath.Rel(r.dir, argv[len(argv)-1])
	if err != nil {
		t.Error(err)
		return started
	}
	r.write(t, filepath.ToSlash(out)+"/journal.jsonl", "{}\n")
	return started
}

func (r *sessionRig) child(i int) *fakeLasting {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.started[i]
}

func (r *sessionRig) children() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.started)
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

func plan() map[string]any { return map[string]any{"content": sessionPlan} }

func TestAStartChecksThePlanThenRunsAChildTheBridgeOwnsWithItsOwnRunFolder(t *testing.T) {
	t.Parallel()
	rig := newSessionRig(t)
	started := rig.act(t, "session.start", plan())
	check := []string{"plantwire", "connect", "--check", "--stdin", "--domains-dir", rig.domainsDir(), "--adapters-dir", "/adapter-files"}
	want := []string{"plantwire", "connect", "--stdin", "--barrier-cap-ms", domain.SessionBarrierCapMS, "--domains-dir", rig.domainsDir(),
		"--adapters-dir", "/adapter-files", "--control-socket", fakeControlSocket, "--out", filepath.Join(rig.dir, "runs", "c-1")}
	if !slices.Equal(rig.argv[0], check) || !slices.Equal(rig.lastingArgv[0], want) || string(rig.lastingStdin[0]) != sessionPlan {
		t.Fatalf("check argv %v, child argv %v stdin %q, want the check %v and then the child %v", rig.argv[0], rig.lastingArgv[0], rig.lastingStdin[0], check, want)
	}
	if started.Run != "c-1" || started.State != domain.SessionRunning || started.Domain != "arduino-thermal" || string(started.Plan) != sessionCheck {
		t.Fatalf("start = %+v, want c-1 running for arduino-thermal with the binary's check", started)
	}
	status := rig.act(t, "session.status", nil)
	if status.Domain != "arduino-thermal" || status.State != domain.SessionRunning || !rig.bridge.runIsRunning("c-1") {
		t.Fatalf("status = %+v, want the session's own domain and plan, counted as a running run", status)
	}
	listing, err := rig.bridge.Workspace()
	if err != nil || len(listing.Runs) != 1 || !listing.Runs[0].Connected || !listing.Runs[0].Running ||
		string(listing.Runs[0].Plan) != sessionCheck {
		t.Fatalf("runs = %+v, %v, want c-1 listed as a running connected session with its plan check", listing.Runs, err)
	}
}

func TestASecondStartIsRefusedAndTwoAtOnceStartOneChild(t *testing.T) {
	t.Parallel()
	rig := newSessionRig(t)
	rig.bridge.edges.RunCommand = func(context.Context, []string, []byte) (process.CommandResult, error) {
		return process.CommandResult{Stdout: []byte(sessionCheck)}, nil
	}
	var wg sync.WaitGroup
	refusals := make(chan error, 2)
	for range 2 {
		wg.Go(func() {
			_, err := rig.bridge.RunAction(context.Background(), "session.start", plan())
			refusals <- err
		})
	}
	wg.Wait()
	close(refusals)
	refused := 0
	for err := range refusals {
		var failure *domain.Failure
		if errors.As(err, &failure) && failure.Code == "session_running" {
			refused++
		}
	}
	if refused != 1 || rig.children() != 1 {
		t.Fatalf("two starts at once: %d refused, %d children, want one of each", refused, rig.children())
	}
}

func TestAPlanTheBinaryRefusesStartsNothing(t *testing.T) {
	t.Parallel()
	rig := newSessionRig(t)
	rig.command = process.CommandResult{ExitCode: 1,
		Stdout: []byte(`{"error":{"code":"plantwire_failed","message":"the plan does not hold","issues":[{"pointer":"/real","message":"x"}]}}`)}
	_, err := rig.bridge.RunAction(context.Background(), "session.start", plan())
	var failure *domain.Failure
	if !errors.As(err, &failure) || failure.Code != "plantwire_failed" || len(failure.Issues) == 0 || rig.children() != 0 {
		t.Fatalf("start of a refused plan = %v, want the binary's refusal and no child", err)
	}
	if status := rig.act(t, "session.status", nil); status.State != domain.SessionNone {
		t.Fatalf("status = %+v, want none", status)
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

func TestACheckRunsConnectOnceWithThePlanAndTheWorkspaceKinds(t *testing.T) {
	t.Parallel()
	rig := newSessionRig(t)
	rig.write(t, fanKind, fanKindBytes)
	if _, err := rig.bridge.RunAction(context.Background(), "session.check", plan()); err != nil {
		t.Fatal(err)
	}
	want := []string{"plantwire", "connect", "--check", "--stdin", "--domains-dir", rig.domainsDir(), "--adapters-dir", "/adapter-files",
		"--kinds-dir", filepath.Join(rig.dir, "kinds")}
	if !slices.Equal(rig.argv[0], want) || string(rig.stdin[0]) != sessionPlan || rig.children() != 0 {
		t.Fatalf("check argv %v stdin %q, want %v and no lasting child", rig.argv[0], rig.stdin[0], want)
	}
}
