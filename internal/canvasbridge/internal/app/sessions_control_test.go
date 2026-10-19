package app

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/process"
)

const (
	fakeControlSocket = "/control/node.sock"
	nodeState         = `{"holds":[],"approve":[],"pending":[],"devices":[]}`
)

type controlFake struct {
	mu       sync.Mutex
	ready    bool
	made     error
	err      error
	reply    []byte
	paths    []string
	lines    [][]byte
	removals int
}

func newControlFake(bridge *Bridge) *controlFake {
	c := &controlFake{ready: true, reply: []byte(`{"result":"applied","op":"state","state":` + nodeState + "}\n")}
	bridge.edges.NewControlSocket = c.newSocket
	bridge.edges.ControlReady = c.isReady
	bridge.edges.SendControl = c.send
	return c
}

func (c *controlFake) newSocket() (string, func(), error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.made != nil {
		return "", nil, c.made
	}
	return fakeControlSocket, func() {
		c.mu.Lock()
		defer c.mu.Unlock()
		c.removals++
	}, nil
}

func (c *controlFake) isReady(string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.ready
}

func (c *controlFake) send(_ context.Context, path string, line []byte) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.paths, c.lines = append(c.paths, path), append(c.lines, slices.Clone(line))
	return c.reply, c.err
}

func (c *controlFake) removed() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.removals
}

func controlCode(t *testing.T, rig *sessionRig, args map[string]any) string {
	t.Helper()
	_, err := rig.bridge.RunAction(context.Background(), "session.control", args)
	var failure *domain.Failure
	if !errors.As(err, &failure) {
		t.Fatalf("session.control = %v, want a failure with a code", err)
	}
	return failure.Code
}

func TestASessionStartsWithAControlSocketAndSaysSoUntilItEnds(t *testing.T) {
	t.Parallel()
	rig := newSessionRig(t)
	started := rig.act(t, "session.start", plan())
	if !started.Controls || !slices.Contains(rig.lastingArgv[0], "--control-socket") || rig.control.removed() != 0 {
		t.Fatalf("start = %+v with argv %v, want the control socket passed and offered", started, rig.lastingArgv[0])
	}
	rig.control.ready = false
	if status := rig.act(t, "session.status", nil); status.Controls || status.State != domain.SessionRunning {
		t.Fatalf("status before the node listens = %+v, want running without controls", status)
	}
	rig.control.ready = true
	if stopped := rig.act(t, "session.stop", nil); stopped.State != domain.SessionEnded || stopped.Controls {
		t.Fatalf("stop = %+v, want ended without controls", stopped)
	}
	rig.act(t, "session.status", nil)
	if rig.control.removed() != 1 {
		t.Fatalf("control directories removed = %d, want exactly one after the session ended", rig.control.removed())
	}
}

func TestAControlRequestReachesTheNodeUnchangedAndItsAnswerComesBackUnchanged(t *testing.T) {
	t.Parallel()
	rig := newSessionRig(t)
	rig.act(t, "session.start", plan())
	raw, err := rig.bridge.RunAction(context.Background(), "session.control",
		map[string]any{"op": "approve_mode", "scope": map[string]any{"kind": "board", "id": "mega-01"}, "enabled": true})
	want := `{"op":"approve_mode","scope":{"id":"mega-01","kind":"board"},"enabled":true}`
	if err != nil || len(rig.control.lines) != 1 || rig.control.paths[0] != fakeControlSocket {
		t.Fatalf("session.control = %v, sent %d lines to %v", err, len(rig.control.lines), rig.control.paths)
	}
	var sent, expected map[string]any
	if json.Unmarshal(rig.control.lines[0], &sent) != nil || json.Unmarshal([]byte(want), &expected) != nil || len(sent) != len(expected) {
		t.Fatalf("the node received %s, want the members of %s", rig.control.lines[0], want)
	}
	if !strings.Contains(string(raw), `"result":"applied"`) || !strings.Contains(string(raw), nodeState) {
		t.Fatalf("answer = %s, want the node's document unchanged", raw)
	}
}

func TestAControlWithoutALivingSessionOrWithoutASocketSaysWhich(t *testing.T) {
	t.Parallel()
	rig := newSessionRig(t)
	if code := controlCode(t, rig, map[string]any{"op": "state"}); code != "no_session" {
		t.Fatalf("control before any session = %s, want no_session", code)
	}
	rig.act(t, "session.start", plan())
	rig.act(t, "session.stop", nil)
	if code := controlCode(t, rig, map[string]any{"op": "state"}); code != "no_session" || len(rig.control.lines) != 0 {
		t.Fatalf("control after the session ended = %s with %d lines sent, want no_session and nothing sent", code, len(rig.control.lines))
	}
	bare := newSessionRig(t)
	bare.control.made = errors.New("no temporary directory")
	if started := bare.act(t, "session.start", plan()); started.State != domain.SessionRunning || started.Controls {
		t.Fatalf("start without a control directory = %+v, want a running session that says it has no controls", started)
	}
	if slices.Contains(bare.lastingArgv[0], "--control-socket") || controlCode(t, bare, map[string]any{"op": "state"}) != "control_unavailable" {
		t.Fatalf("a session without a socket: argv %v, want no flag and control_unavailable", bare.lastingArgv[0])
	}
}

func TestAControlThatFailsOnTheWayReportsHowItFailed(t *testing.T) {
	t.Parallel()
	for name, tc := range map[string]struct {
		err   error
		reply []byte
		code  string
	}{
		"silent":     {err: process.ErrControlSilent, code: "control_timeout"},
		"gone":       {err: errors.New("connection refused"), code: "control_gone"},
		"unreadable": {reply: []byte("not json\n"), code: "control_unreadable"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			rig := newSessionRig(t)
			rig.act(t, "session.start", plan())
			rig.control.err, rig.control.reply = tc.err, tc.reply
			if code := controlCode(t, rig, map[string]any{"op": "state"}); code != tc.code {
				t.Fatalf("control = %s, want %s", code, tc.code)
			}
		})
	}
}

func TestAControlOverTheLineLimitIsNeverSent(t *testing.T) {
	t.Parallel()
	rig := newSessionRig(t)
	rig.act(t, "session.start", plan())
	_, err := rig.bridge.RunAction(context.Background(), "session.control", map[string]any{"op": strings.Repeat("a", domain.MaxControlLine)})
	if !errors.Is(err, domain.ErrTooLarge) || len(rig.control.lines) != 0 {
		t.Fatalf("an oversized control = %v with %d lines sent, want ErrTooLarge and nothing sent", err, len(rig.control.lines))
	}
}

func TestClosingTheBridgeRemovesTheControlDirectoryOfALivingSession(t *testing.T) {
	t.Parallel()
	rig := newSessionRig(t)
	rig.act(t, "session.start", plan())
	if err := rig.bridge.Close(); err != nil || rig.control.removed() != 1 {
		t.Fatalf("Close = %v with %d control directories removed, want one", err, rig.control.removed())
	}
}
