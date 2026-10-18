package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/uds"
)

type RelayOptions struct {
	GatewaySocket string
	HostSocket    string
	Out           string
	Simulated     *SimulatedOptions
	Ready         func()
	Control       *ControlPort
}

const (
	relayTick = 250 * time.Millisecond
	relayWait = 30 * time.Second
)

func Relay(ctx context.Context, options RelayOptions) (domain.RelayStats, error) {
	ctx, stop := uds.UntilInterrupted(ctx)
	defer stop()
	ticks, stopTicks := uds.Ticks(relayTick)
	defer stopTicks()
	return serveRelay(ctx, options, ticks)
}

func serveRelay(ctx context.Context, options RelayOptions, ticks <-chan time.Time) (domain.RelayStats, error) {
	loop, err := openRelay(ctx, options, ticks)
	if err != nil {
		return domain.RelayStats{}, fmt.Errorf("plantwire relay: %w", err)
	}
	err = loop.serve(ctx)
	stats := loop.relay.Stats()
	stats.TraceIncomplete = loop.rec.incomplete
	return stats, err
}

func openRelay(ctx context.Context, options RelayOptions, ticks <-chan time.Time) (*relayLoop, error) {
	sim, err := openRelaySim(ctx, options)
	if err != nil {
		return nil, err
	}
	relay, err := newRelay(sim, options.Control != nil)
	if err != nil {
		sim.discard()
		return nil, err
	}
	loop := &relayLoop{relay: relay, options: options, sim: sim, ticks: ticks}
	return loop, loop.open(ctx)
}

func openRelaySim(ctx context.Context, options RelayOptions) (*simNode, error) {
	if err := checkRelayOptions(options); err != nil {
		return nil, err
	}
	return openSim(ctx, options)
}

func newRelay(sim *simNode, control bool) (*domain.Relay, error) {
	relay, err := buildRelay(sim)
	if err == nil && control {
		relay.EnableInterrupts()
	}
	return relay, err
}

func buildRelay(sim *simNode) (*domain.Relay, error) {
	if sim == nil {
		return domain.NewRelay(relayTick.Milliseconds()), nil
	}
	relay, err := domain.NewMixedRelay(relayTick.Milliseconds(), sim.birth)
	if err == nil {
		relay.Plan(sim.instances, sim.bound)
	}
	return relay, err
}

func (l *relayLoop) open(ctx context.Context) error {
	rec, err := openRecording(l.options.Out, l.sim, l.options.Control != nil)
	if err != nil {
		l.sim.discard()
		return err
	}
	l.rec = rec
	return l.listen(ctx)
}

func (l *relayLoop) listen(ctx context.Context) error {
	gateways, err := uds.ListenGateways(ctx, l.options.GatewaySocket, relayWait)
	if err != nil {
		l.sim.discard()
		return errors.Join(err, l.rec.close())
	}
	l.gateways = gateways
	if l.options.Ready != nil {
		l.options.Ready()
	}
	return nil
}

func checkRelayOptions(o RelayOptions) error {
	switch {
	case o.GatewaySocket == "" || o.HostSocket == "" || o.GatewaySocket == o.HostSocket:
		return fmt.Errorf("%w: a relay needs two different sockets, the gateway's and the host's", ErrBadOptions)
	case o.Simulated != nil && o.Out == "":
		return fmt.Errorf("%w: a mixed relay needs a run folder for its artifact and its field trace", ErrBadOptions)
	}
	return nil
}
