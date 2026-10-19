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
	stopWithin   = 30 * time.Second
	openPoll     = 100 * time.Millisecond
	openAttempts = 100
)

type Lasting interface {
	Exited() (process.LastingExit, bool)
	Stop(within time.Duration) process.LastingExit
}

func (b *Bridge) lasting(ctx context.Context, action domain.Action, args map[string]any) (json.RawMessage, error) {
	if action.Lasting == domain.LastingStart {
		return b.startSession(ctx, action, args)
	}
	if action.Lasting == domain.LastingControl {
		return b.control(ctx, args)
	}
	if _, err := action.Argv(args); err != nil {
		return nil, err
	}
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
	check, err := b.runCommand(ctx, domain.SessionPrecheck(), args)
	if err != nil {
		return nil, err
	}
	session, out, err := b.claimSession(check)
	if err != nil {
		return nil, err
	}
	return b.launch(ctx, session, append(append(argv, b.controlFlags(session)...), "--out", out), stdin)
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

func (b *Bridge) claimSession(check json.RawMessage) (*liveSession, string, error) {
	b.stateMu.Lock()
	defer b.stateMu.Unlock()
	if err := b.vacated(); err != nil {
		return nil, "", err
	}
	run, out, err := b.freshRunFolder()
	if err != nil {
		return nil, "", err
	}
	b.session = &liveSession{view: domain.SessionView{Run: run, Domain: domain.PlanDomain(check), Plan: check}}
	b.sessions[run] = check
	return b.session, out, nil
}

func (b *Bridge) vacated() error {
	if b.session == nil {
		return nil
	}
	if b.session.live() {
		return domain.SessionStillRunning(b.session.view.Run)
	}
	b.session.releaseControl()
	return nil
}

func (b *Bridge) freshRunFolder() (string, string, error) {
	for range domain.RunIDAttempts {
		run := b.edges.NewRunID()
		folder, err := domain.SessionFolder(run)
		if err != nil {
			return "", "", err
		}
		if !b.edges.Workspace.Exists(folder.Path) {
			return run, b.absolute([]domain.ArgvPart{folder})[0], nil
		}
	}
	return "", "", fmt.Errorf("%w: no free run folder after %d tries", domain.ErrConflict, domain.RunIDAttempts)
}

func (b *Bridge) launch(ctx context.Context, session *liveSession, argv []string, stdin []byte) (json.RawMessage, error) {
	lasting, err := b.edges.StartLasting(argv, stdin)
	if err != nil {
		session.fail(err)
		return nil, err
	}
	if session.attach(lasting) {
		lasting.Stop(stopWithin)
	}
	b.awaitOpen(ctx, session.view.Run, lasting)
	return encodeStatus(b.statusOf(session))
}

func (b *Bridge) awaitOpen(ctx context.Context, run string, lasting Lasting) {
	journal, attempts := "runs/"+run+"/journal.jsonl", 0
	_ = b.everyTick(ctx, openPoll, func() (bool, error) {
		attempts++
		_, exited := lasting.Exited()
		return attempts >= openAttempts || exited || b.edges.Workspace.Exists(journal), nil
	})
}

func (b *Bridge) sessionStatus() domain.SessionStatus {
	if current := b.currentSession(); current != nil {
		return b.statusOf(current)
	}
	return domain.SessionStatus{State: domain.SessionNone}
}

func (b *Bridge) stopSession() domain.SessionStatus {
	current := b.currentSession()
	if current == nil {
		return domain.SessionStatus{State: domain.SessionNone}
	}
	if lasting := current.requestStop(); lasting != nil {
		lasting.Stop(stopWithin)
	}
	return b.statusOf(current)
}

func (b *Bridge) stopLiving() {
	if current := b.currentSession(); current != nil {
		if lasting := current.requestStop(); lasting != nil {
			lasting.Stop(stopWithin)
		}
		current.releaseControl()
	}
}

func (b *Bridge) currentSession() *liveSession {
	b.stateMu.Lock()
	defer b.stateMu.Unlock()
	return b.session
}

func (b *Bridge) sessionRuns(runID string) bool {
	return b.session != nil && b.session.view.Run == runID && b.session.live()
}

func (b *Bridge) startedSession(runID string) (json.RawMessage, bool) {
	b.stateMu.Lock()
	defer b.stateMu.Unlock()
	plan, started := b.sessions[runID]
	return plan, started
}

func encodeStatus(status domain.SessionStatus) (json.RawMessage, error) {
	raw, err := json.Marshal(status)
	if err != nil {
		return nil, fmt.Errorf("canvasbridge: session status: %w", err)
	}
	return raw, nil
}
