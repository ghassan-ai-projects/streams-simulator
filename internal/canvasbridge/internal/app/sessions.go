package app

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/process"
)

const (
	stopWithin   = 20 * time.Second
	openPoll     = 100 * time.Millisecond
	openAttempts = 100
)

type Lasting interface {
	Exited() (process.LastingExit, bool)
	Stop(within time.Duration) process.LastingExit
}

type liveSession struct {
	run     string
	lasting Lasting
}

func (s *liveSession) running() bool {
	_, exited := s.lasting.Exited()
	return !exited
}

func (s *liveSession) status() domain.SessionStatus {
	exit, exited := s.lasting.Exited()
	if !exited {
		return domain.SessionStatus{Run: s.run, State: domain.SessionRunning}
	}
	return domain.SessionOutcome(s.run, exit.Stdout, exit.ExitCode, exit.Stderr)
}

func (b *Bridge) lasting(ctx context.Context, action domain.Action, args map[string]any) (json.RawMessage, error) {
	if action.Lasting == domain.LastingStart {
		return b.startSession(ctx, action, args)
	}
	if _, err := action.Argv(args); err != nil {
		return nil, err
	}
	b.sessionMu.Lock()
	defer b.sessionMu.Unlock()
	if action.Lasting == domain.LastingStop {
		return encodeStatus(b.stopSession())
	}
	return encodeStatus(b.sessionStatus())
}

func (b *Bridge) startSession(ctx context.Context, action domain.Action, args map[string]any) (json.RawMessage, error) {
	argv, stdin, err := b.sessionCommand(action, args)
	if err != nil {
		return nil, err
	}
	b.sessionMu.Lock()
	defer b.sessionMu.Unlock()
	if current := b.currentSession(); current != nil && current.running() {
		return nil, domain.SessionStillRunning(current.run)
	}
	return b.launch(ctx, argv, stdin)
}

func (b *Bridge) sessionCommand(action domain.Action, args map[string]any) ([]string, []byte, error) {
	parts, err := action.Argv(args)
	if err != nil {
		return nil, nil, err
	}
	stdin, err := action.Stdin(args)
	if err != nil {
		return nil, nil, err
	}
	return append(b.absolute(parts), b.flags(action)...), stdin, nil
}

func (b *Bridge) launch(ctx context.Context, argv []string, stdin []byte) (json.RawMessage, error) {
	run, out, err := b.freshRunFolder()
	if err != nil {
		return nil, err
	}
	lasting, err := b.edges.StartLasting(append(argv, "--out", out), stdin)
	if err != nil {
		return nil, err
	}
	session := &liveSession{run: run, lasting: lasting}
	b.setSession(session)
	b.awaitOpen(ctx, session)
	return encodeStatus(session.status())
}

func (b *Bridge) awaitOpen(ctx context.Context, session *liveSession) {
	journal, attempts := "runs/"+session.run+"/journal.jsonl", 0
	_ = b.everyTick(ctx, openPoll, func() (bool, error) {
		attempts++
		return attempts >= openAttempts || !session.running() || b.edges.Workspace.Exists(journal), nil
	})
}

func (b *Bridge) freshRunFolder() (string, string, error) {
	run := b.edges.NewRunID()
	folder, err := domain.SessionFolder(run)
	if err != nil {
		return "", "", err
	}
	if b.edges.Workspace.Exists(folder.Path) {
		return "", "", fmt.Errorf("%w: run folder %s exists", domain.ErrConflict, folder.Path)
	}
	return run, b.absolute([]domain.ArgvPart{folder})[0], nil
}

func (b *Bridge) sessionStatus() domain.SessionStatus {
	if current := b.currentSession(); current != nil {
		return current.status()
	}
	return domain.SessionStatus{State: domain.SessionNone}
}

func (b *Bridge) stopSession() domain.SessionStatus {
	current := b.currentSession()
	if current == nil {
		return domain.SessionStatus{State: domain.SessionNone}
	}
	exit := current.lasting.Stop(stopWithin)
	return domain.SessionOutcome(current.run, exit.Stdout, exit.ExitCode, exit.Stderr)
}

func (b *Bridge) stopLiving() {
	b.sessionMu.Lock()
	defer b.sessionMu.Unlock()
	if current := b.currentSession(); current != nil && current.running() {
		current.lasting.Stop(stopWithin)
	}
}

func (b *Bridge) currentSession() *liveSession {
	b.stateMu.Lock()
	defer b.stateMu.Unlock()
	return b.session
}

func (b *Bridge) setSession(session *liveSession) {
	b.stateMu.Lock()
	defer b.stateMu.Unlock()
	b.session = session
}

func (b *Bridge) sessionRuns(runID string) bool {
	return b.session != nil && b.session.run == runID && b.session.running()
}

func encodeStatus(status domain.SessionStatus) (json.RawMessage, error) {
	raw, err := json.Marshal(status)
	if err != nil {
		return nil, fmt.Errorf("canvasbridge: session status: %w", err)
	}
	return raw, nil
}
