package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/uds"
)

func (l *relayLoop) ensureHost(ctx context.Context) error {
	if !l.relay.WantsHost() {
		return nil
	}
	link, err := uds.Dial(context.WithoutCancel(ctx), l.options.HostSocket, relayWait)
	switch {
	case errors.Is(err, uds.ErrSocketNotPrivate):
		return errors.Join(l.apply(l.relay.HostUnreachable()), err)
	case err != nil:
		return l.apply(l.relay.HostUnreachable())
	}
	l.host = link
	return l.apply(l.relay.HostAttached())
}

func (l *relayLoop) apply(moves domain.Moves) error {
	traced := l.record(moves.Trace)
	upErr, downErr := l.sendUp(moves.Up), l.sendDown(moves.Down)
	if moves.CloseHost {
		l.closeHost()
	}
	if moves.CloseGateway {
		l.closeGateway()
	}
	return errors.Join(moves.Err, traced, upErr, downErr)
}

func (l *relayLoop) record(traced []domain.Traced) error {
	if l.trace == nil || len(traced) == 0 {
		return nil
	}
	for _, line := range traced {
		l.trace.Wire(line.Direction, line.Record)
	}
	if err := l.trace.Flush(); err != nil {
		return fmt.Errorf("plantwire relay: flush trace: %w", err)
	}
	return nil
}

func (l *relayLoop) sendUp(records []map[string]any) error {
	return deliver(l.host, records, "host", l.loseHost)
}

func (l *relayLoop) sendDown(records []map[string]any) error {
	return deliver(l.gateway, records, "gateway", l.loseGateway)
}

func deliver(link *uds.Link, records []map[string]any, side string, lost func() error) error {
	if len(records) == 0 || link == nil {
		return nil
	}
	if err := link.Send(records); err != nil {
		slog.Warn("plantwire relay: send failed", "side", side, "error", err)
		return lost()
	}
	return nil
}

func (l *relayLoop) loseHost() error {
	moves := l.relay.HostLost()
	l.closeHost()
	return l.apply(moves)
}

func (l *relayLoop) loseGateway() error {
	moves := l.relay.GatewayLost()
	l.closeGateway()
	return l.apply(moves)
}

func (l *relayLoop) closeHost() {
	if l.host != nil {
		_ = l.host.Close()
		l.host = nil
	}
}

func (l *relayLoop) closeGateway() {
	if l.gateway != nil {
		_ = l.gateway.Close()
		l.gateway = nil
	}
}
