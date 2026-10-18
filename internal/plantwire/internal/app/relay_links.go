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
	l.rec.noted(moves.Notes)
	l.interrupted(moves.Interrupt)
	traced := l.record(moves.Trace)
	upErr, downErr := l.sendUp(moves.Up), l.sendDown(moves.Down)
	l.sim.deliver(moves.ToSim)
	if moves.CloseHost {
		l.closeHost()
	}
	if moves.CloseGateway {
		l.closeGateway()
	}
	return errors.Join(moves.Err, traced, upErr, downErr, l.endIncompleteTrace())
}

func (l *relayLoop) endIncompleteTrace() error {
	if l.gateway == nil || !l.rec.incomplete {
		return nil
	}
	return l.loseGateway()
}

func (l *relayLoop) record(traced []domain.Traced) error {
	if l.rec.empty() || len(traced) == 0 {
		return nil
	}
	var failed error
	for _, line := range traced {
		failed = errors.Join(failed, l.rec.note(line))
	}
	if err := l.rec.flush(); err != nil {
		return fmt.Errorf("plantwire relay: flush trace: %w", err)
	}
	return failed
}

func (l *relayLoop) onSim() error {
	pending := l.drainSim()
	if l.sim.reap() {
		pending = errors.Join(pending, l.applySim(l.relay.SimLost()))
	}
	return pending
}

func (l *relayLoop) drainSim() error {
	for _, record := range l.sim.link.TakeUp() {
		in, err := domain.InboundOf(record)
		if err != nil {
			slog.Warn("plantwire relay: the simulated node sent no record", "error", err)
			continue
		}
		if err := l.applySim(l.relay.FromSim(in)); err != nil {
			return err
		}
	}
	return nil
}

func (l *relayLoop) applySim(moves domain.SimMoves) error {
	l.rec.noted(moves.Notes)
	l.interrupted(moves.Interrupt)
	return errors.Join(l.record(moves.Trace), l.sendUp(moves.Up))
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

func (l *relayLoop) interrupted(entries []domain.InterruptEntry) {
	if len(entries) == 0 {
		return
	}
	l.rec.interrupts(entries)
	view := l.relay.InterruptView()
	l.options.Control.Publish(view)
	l.rec.liveInterrupts(view)
}

func (l *relayLoop) serveControls() error {
	var failed error
	for call, ok := l.options.Control.take(); ok; call, ok = l.options.Control.take() {
		reply, moves := l.relay.Control(call.Entry)
		failed = errors.Join(failed, l.apply(moves))
		call.Reply <- domain.ControlReply(reply)
	}
	return failed
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
