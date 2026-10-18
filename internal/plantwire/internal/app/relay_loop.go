package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/uds"
)

type relayLoop struct {
	relay    *domain.Relay
	options  RelayOptions
	rec      recording
	sim      *simNode
	gateways *uds.Gateways
	gateway  *uds.Link
	host     *uds.Link
	ticks    <-chan time.Time
	lastTick time.Time
}

func (l *relayLoop) serve(ctx context.Context) error {
	for ctx.Err() == nil {
		if err := l.step(ctx); err != nil {
			return l.finish(err)
		}
	}
	return l.finish(nil)
}

func (l *relayLoop) step(ctx context.Context) error {
	if l.gateway == nil {
		return l.awaitGateway(ctx)
	}
	return l.exchange(ctx)
}

func (l *relayLoop) awaitGateway(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return nil
	case link := <-l.gateways.Links():
		return l.attach(link)
	case in, open := <-l.hostIn():
		return l.onHost(in, open)
	case <-l.sim.signal():
		return l.onSim()
	case at := <-l.ticks:
		return l.onTick(ctx, at)
	}
}

func (l *relayLoop) exchange(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return nil
	case in, open := <-l.gateway.Incoming():
		return l.onGateway(ctx, in, open)
	case in, open := <-l.hostIn():
		return l.onHost(in, open)
	case <-l.sim.signal():
		return l.onSim()
	case at := <-l.ticks:
		return l.onTick(ctx, at)
	}
}

func (l *relayLoop) finish(cause error) error {
	shutdown := l.apply(l.relay.Shutdown())
	l.closeHost()
	l.closeGateway()
	stopped := l.sim.stop()
	closed := l.rec.close()
	ended := l.sim.end(l.options.Out, l.rec.digest(), l.rec.interruptDigest())
	if err := errors.Join(cause, shutdown, l.gateways.Close(), closed, stopped, ended); err != nil {
		return fmt.Errorf("plantwire relay: %w", err)
	}
	return nil
}

func (l *relayLoop) hostIn() <-chan domain.Inbound {
	if l.host == nil {
		return nil
	}
	return l.host.Incoming()
}

func (l *relayLoop) attach(link *uds.Link) error {
	if l.rec.incomplete {
		return link.Close()
	}
	l.gateway = link
	l.relay.GatewayAttached()
	return link.Expect(uds.HandshakeWait)
}

func (l *relayLoop) onGateway(ctx context.Context, in domain.Inbound, open bool) error {
	if !open {
		return l.loseGateway()
	}
	if err := l.apply(l.relay.FromGateway(in)); err != nil {
		return err
	}
	l.settle()
	if in.Type != "NBIRTH" {
		return nil
	}
	return l.ensureHost(ctx)
}

func (l *relayLoop) onHost(in domain.Inbound, open bool) error {
	if !open {
		return l.loseHost()
	}
	return l.apply(l.relay.FromHost(in))
}

func (l *relayLoop) onTick(ctx context.Context, at time.Time) error {
	elapsed := l.elapsedSince(at)
	l.rec.tickLive(elapsed)
	l.relay.SetWall(at.UnixNano())
	if err := errors.Join(l.apply(l.relay.TickBy(elapsed)), l.serveControls()); err != nil {
		return err
	}
	return l.ensureHost(ctx)
}

func (l *relayLoop) elapsedSince(at time.Time) int64 {
	elapsed := relayTick.Milliseconds()
	if !l.lastTick.IsZero() {
		elapsed = max(1, at.Sub(l.lastTick).Milliseconds())
	}
	l.lastTick = at
	return elapsed
}

func (l *relayLoop) settle() {
	if l.gateway != nil && l.relay.GatewayAdmitted() {
		_ = l.gateway.Settle()
	}
}
