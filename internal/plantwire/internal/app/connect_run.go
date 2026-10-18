package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/uds"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

const openEndedNS int64 = 100 * 365 * 24 * 3600 * 1_000_000_000

func Connect(ctx context.Context, options ConnectOptions) (ConnectResult, error) {
	if options.Out == "" || options.SliceNS <= 0 {
		return ConnectResult{}, fmt.Errorf("%w: a connected session needs a run folder and a slice", ErrBadOptions)
	}
	planned, err := planConnect(options)
	if err != nil {
		return ConnectResult{}, err
	}
	ctx, stop := uds.UntilInterrupted(ctx)
	defer stop()
	if planned.plan.GatewaySocket == "" {
		return connectNode(ctx, options, &planned)
	}
	return connectRelay(ctx, options, &planned)
}

func (p *connectPlan) newRun(ctx context.Context, out string) (*run.Run, error) {
	r, err := run.New(context.WithoutCancel(ctx), run.Config{Domain: p.spec, Adapter: p.adapter, Seed: p.session.WorldSeed(),
		SinkName: model.SinkInproc, TimeMode: model.TimeStepped, StartTimeNS: model.DefaultStartTimeNS, JournalDir: out, SessionRecord: sessionNote(p.result())})
	if err != nil {
		return nil, fmt.Errorf("plantwire connect: %w", err)
	}
	return r, nil
}

func sessionNote(check ConnectCheck) map[string]any {
	if check.Instances == nil {
		check.Instances = []domain.PlannedInstance{}
	}
	if check.ImplementsChecked == nil {
		check.ImplementsChecked = []model.ImplementsChecked{}
	}
	return map[string]any{"mode": "connected", "plan": check}
}

func connectNode(ctx context.Context, o ConnectOptions, p *connectPlan) (ConnectResult, error) {
	r, err := p.newRun(ctx, o.Out)
	if err != nil {
		return ConnectResult{}, err
	}
	node := p.nodeOptions(o, "node-1")
	node.JournalOrigins = true
	served := serveSession(ctx, r, p.plan.HostSocket, o.BarrierCap, node)
	stopped := ctx.Err() != nil
	_, ended := r.End(o.Out)
	return ConnectResult{Out: o.Out, Stopped: stopped}, joined(withoutInterrupt(stopped, served), ended)
}

func withoutInterrupt(stopped bool, served error) error {
	if stopped && uds.Interrupted(served) {
		return nil
	}
	return served
}

func connectRelay(ctx context.Context, o ConnectOptions, p *connectPlan) (ConnectResult, error) {
	ticks, stopTicks := uds.Ticks(relayTick)
	defer stopTicks()
	stats, err := serveRelay(ctx, p.relayOptions(o), ticks)
	if err != nil {
		return ConnectResult{}, err
	}
	return ConnectResult{Out: o.Out, Stopped: ctx.Err() != nil, Relay: &stats}, nil
}

func (p *connectPlan) relayOptions(o ConnectOptions) RelayOptions {
	return RelayOptions{GatewaySocket: p.plan.GatewaySocket, HostSocket: p.plan.HostSocket, Out: o.Out, Ready: o.Ready,
		Simulated: &SimulatedOptions{DomainsDir: o.DomainsDir, AdaptersDir: o.AdaptersDir, Domain: p.session.Domain, Adapter: connectAdapter,
			Seed: p.session.WorldSeed(), BarrierCap: o.BarrierCap, Node: p.nodeOptions(o, "sim-1"), Session: p.result()}}
}

func (p *connectPlan) nodeOptions(o ConnectOptions, session string) Options {
	return Options{Environment: domain.EnvironmentSimulated, ProtocolMax: domain.Version, SessionID: session, SliceNS: o.SliceNS,
		UntilNS: model.DefaultStartTimeNS + openEndedNS, Policies: o.Policies, KindsDirs: o.KindsDirs, Bound: p.plan.Bound}
}

func joined(served, ended error) error {
	if err := errors.Join(served, ended); err != nil {
		return fmt.Errorf("plantwire connect: %w", err)
	}
	return nil
}
