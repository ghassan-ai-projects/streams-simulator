package app

import (
	"context"
	"encoding/json"
	"errors"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/domain"
)

func TestASessionStartWithAnArgumentItDoesNotTakeIsRefusedBeforeAnyCheckOrChild(t *testing.T) {
	t.Parallel()
	rig := newSessionRig(t)
	args := plan()
	args["bogus"] = 1
	if _, err := rig.bridge.RunAction(context.Background(), "session.start", args); !errors.Is(err, domain.ErrBadArguments) || rig.children() != 0 || len(rig.argv) != 0 {
		t.Fatalf("session.start with an extra argument = %v after %d children and %d commands, want ErrBadArguments and nothing run", err, rig.children(), len(rig.argv))
	}
}

func TestASessionWhoseRunIdIsNotAFolderNameClaimsNothing(t *testing.T) {
	t.Parallel()
	rig := newSessionRig(t)
	rig.bridge.edges.NewRunID = func() string { return "../escape" }
	if _, err := rig.bridge.RunAction(context.Background(), "session.start", plan()); !errors.Is(err, domain.ErrBadPath) || rig.children() != 0 {
		t.Fatalf("session.start with run id ../escape = %v after %d children, want ErrBadPath and no child", err, rig.children())
	}
	if status := rig.act(t, "session.status", nil); status.State != domain.SessionNone {
		t.Fatalf("status after the refused start = %+v, want no session claimed", status)
	}
}

func TestAFailedLaunchEndsItsSessionSoControlSaysNoSessionAndANewStartRuns(t *testing.T) {
	t.Parallel()
	rig := newSessionRig(t)
	launch := rig.bridge.edges.StartLasting
	errNoExecutable := errors.New("no such executable")
	rig.bridge.edges.StartLasting = func([]string, []byte) (Lasting, error) { return nil, errNoExecutable }
	if _, err := rig.bridge.RunAction(context.Background(), "session.start", plan()); !errors.Is(err, errNoExecutable) {
		t.Fatalf("a session whose child cannot launch = %v, want the launch error", err)
	}
	if code := controlCode(t, rig, map[string]any{"op": "state"}); code != "no_session" {
		t.Fatalf("control after a failed launch = %s, want no_session", code)
	}
	rig.bridge.edges.StartLasting = launch
	if started := rig.act(t, "session.start", plan()); started.State != domain.SessionRunning || started.Run != "c-2" {
		t.Fatalf("a start after a failed launch = %+v, want a running c-2", started)
	}
	if rig.control.removed() != 1 {
		t.Fatalf("control directories removed = %d, want the failed session's one released by the next start", rig.control.removed())
	}
}

func TestEncodingAStatusThatIsNotJSONIsReportedNotSent(t *testing.T) {
	t.Parallel()
	_, err := encodeStatus(domain.SessionStatus{State: domain.SessionRunning, Result: json.RawMessage("{")})
	if err == nil || !strings.Contains(err.Error(), "canvasbridge: session status") {
		t.Fatalf("encodeStatus of a result that is not JSON = %v, want an encoding error", err)
	}
}

func TestTheLaunchEdgeStartsALastingChildOrSaysWhy(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if _, err := startLasting(filepath.Join(dir, "absent"), dir)([]string{"plantwire"}, nil); err == nil || !strings.Contains(err.Error(), "canvasbridge: start streamsim") {
		t.Fatalf("launching a missing executable = %v, want the start error", err)
	}
	truthy, err := exec.LookPath("true")
	if err != nil {
		t.Fatal(err)
	}
	lasting, err := startLasting(truthy, dir)([]string{"plantwire"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	_ = lasting.Stop(time.Second)
	if _, exited := lasting.Exited(); !exited {
		t.Fatal("a child stopped by the bridge is still running")
	}
}
