package app

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/process"
)

func (b *Bridge) controlFlags(session *liveSession) []string {
	path, remove, err := b.edges.NewControlSocket()
	if err != nil {
		slog.Warn("canvasbridge: the session starts without a control socket", "error", err)
		return nil
	}
	session.offerControl(path, remove)
	return []string{"--control-socket", path}
}

func (b *Bridge) statusOf(session *liveSession) domain.SessionStatus {
	status := session.status()
	if !domain.Controlling(status.State) {
		session.releaseControl()
		return status
	}
	path := session.controlPath()
	status.Controls = path != "" && b.edges.ControlReady(path)
	return status
}

func (b *Bridge) control(ctx context.Context, args map[string]any) (json.RawMessage, error) {
	session, err := b.controlled()
	if err != nil {
		return nil, err
	}
	line, err := domain.ControlLine(args)
	if err != nil {
		return nil, err
	}
	reply, err := b.edges.SendControl(ctx, session.controlPath(), line)
	return answered(session.view.Run, reply, err)
}

func (b *Bridge) controlled() (*liveSession, error) {
	current := b.currentSession()
	switch {
	case current == nil || !current.live():
		return nil, domain.NoLiveSession()
	case current.controlPath() == "":
		return nil, domain.ControlNotOffered(current.view.Run)
	}
	return current, nil
}

func answered(run string, reply []byte, err error) (json.RawMessage, error) {
	switch {
	case errors.Is(err, process.ErrControlSilent):
		return nil, domain.ControlSilent(run)
	case err != nil:
		return nil, domain.ControlGone(run)
	}
	return domain.ControlAnswer(run, reply)
}
