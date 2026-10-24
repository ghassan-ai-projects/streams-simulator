package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/files"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/uds"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

const openEndedNS int64 = 100 * 365 * 24 * 3600 * 1_000_000_000

func Connect(ctx context.Context, options ConnectOptions) (ConnectResult, error) {
	if options.Out == "" || options.SliceNS <= 0 {
		return ConnectResult{}, fmt.Errorf("%w: a connected session needs a run folder and a slice", ErrBadOptions)
	}
	if err := freshRunFolder(options.Out); err != nil {
		return ConnectResult{}, err
	}
	planned, err := planConnect(options)
	if err != nil {
		return ConnectResult{}, err
	}
	ctx, stop := uds.UntilInterrupted(ctx)
	defer stop()
	return planned.serve(ctx, options)
}

func freshRunFolder(out string) error {
	present, err := files.RunFolderFiles(out)
	if err == nil {
		err = domain.CheckFreshRunFolder(out, present)
	}
	if err != nil {
		return fmt.Errorf("plantwire connect: %w", err)
	}
	return nil
}

func (p *connectPlan) serve(ctx context.Context, o ConnectOptions) (ConnectResult, error) {
	if p.session.PacedByWall() {
		p.pacer = uds.StartWallPacer()
	}
	control, err := openControl(ctx, o.ControlSocket)
	if err != nil {
		return ConnectResult{}, err
	}
	defer control.close()
	if p.plan.GatewaySocket == "" {
		return connectNode(ctx, o, p, control.port)
	}
	return connectRelay(ctx, o, p, control.port)
}

type controlEndpoint struct {
	port   *ControlPort
	server *uds.ControlServer
}

func openControl(ctx context.Context, path string) (*controlEndpoint, error) {
	if path == "" {
		return &controlEndpoint{}, nil
	}
	port := NewControlPort()
	server, err := uds.ListenControl(ctx, path, port)
	if err != nil {
		return nil, fmt.Errorf("plantwire connect: %w", err)
	}
	return &controlEndpoint{port: port, server: server}, nil
}

func (c *controlEndpoint) close() {
	if c.server != nil {
		_ = c.server.Close()
	}
}

func (p *connectPlan) newRun(ctx context.Context, out string) (*run.Run, error) {
	r, err := run.New(context.WithoutCancel(ctx), run.Config{Domain: p.spec, Adapter: p.adapter, Seed: p.session.WorldSeed(),
		SinkName: model.SinkInproc, TimeMode: model.TimeStepped, StartTimeNS: p.startNS(), JournalDir: out, SessionRecord: sessionNote(p.result()),
		BoardDigests: boardDigests(p.checked)})
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
	if check.Policies == nil {
		check.Policies = []string{}
	}
	return map[string]any{"mode": "connected", "plan": check}
}

func connectNode(ctx context.Context, o ConnectOptions, p *connectPlan, port *ControlPort) (ConnectResult, error) {
	r, err := p.newRun(ctx, o.Out)
	if err != nil {
		return ConnectResult{}, err
	}
	node := p.nodeOptions(o, "node-1")
	node.JournalOrigins = true
	file, err := controlLog(&node, o.Out, port)
	if err != nil {
		discardRun(r)
		return ConnectResult{}, err
	}
	return serveNode(ctx, o, p, r, node, file)
}

func controlLog(node *Options, dir string, port *ControlPort) (*interruptFile, error) {
	if port == nil {
		return nil, nil
	}
	node.Control = port
	file, err := openInterruptFile(dir)
	if err == nil {
		node.InterruptLog = file
	}
	return file, err
}

func serveNode(ctx context.Context, o ConnectOptions, p *connectPlan, r *run.Run, node Options, file *interruptFile) (ConnectResult, error) {
	served := serveSession(ctx, r, p.plan.HostSocket, o.BarrierCap, node)
	stopped := ctx.Err() != nil
	failure := withoutInterrupt(stopped, served)
	sealed := file.finish(o.Out)
	r.SetInterruptLog(file.digest())
	abandonFailed(r, "plantwire connect", failure)
	_, ended := r.End(o.Out)
	return ConnectResult{Out: o.Out, Stopped: stopped}, joined(failure, errors.Join(sealed, ended))
}

func abandonFailed(r *run.Run, session string, failure error) {
	if failure != nil {
		r.Abandon(fmt.Sprintf("%s: %v", session, failure))
	}
}

func withoutInterrupt(stopped bool, served error) error {
	if stopped && uds.Interrupted(served) {
		return nil
	}
	return served
}

func connectRelay(ctx context.Context, o ConnectOptions, p *connectPlan, port *ControlPort) (ConnectResult, error) {
	ticks, stopTicks := uds.Ticks(relayTick)
	defer stopTicks()
	stats, err := serveRelay(ctx, p.relayOptions(o, port), ticks)
	if err != nil {
		return ConnectResult{}, err
	}
	return ConnectResult{Out: o.Out, Stopped: ctx.Err() != nil, Relay: &stats}, nil
}

func (p *connectPlan) relayOptions(o ConnectOptions, port *ControlPort) RelayOptions {
	return RelayOptions{GatewaySocket: p.plan.GatewaySocket, HostSocket: p.plan.HostSocket, Out: o.Out, Ready: o.Ready, Control: port,
		Simulated: &SimulatedOptions{DomainsDir: o.DomainsDir, AdaptersDir: o.AdaptersDir, Domain: p.session.Domain, Adapter: connectAdapter,
			Seed: p.session.WorldSeed(), BarrierCap: o.BarrierCap, Node: p.nodeOptions(o, "sim-1"), Session: p.result(), StartNS: p.startNS()}}
}

func (p *connectPlan) nodeOptions(o ConnectOptions, session string) Options {
	options := Options{Environment: domain.EnvironmentSimulated, ProtocolMax: domain.Version, SessionID: session, SliceNS: o.SliceNS,
		UntilNS: p.startNS() + openEndedNS, Policies: p.plan.Policies.Digests(), KindsDirs: o.KindsDirs, Bound: p.plan.Bound, Interfaces: p.faces}
	if p.pacer != nil {
		options.Pace = p.pacer
	}
	options.Disturbances = domain.ScheduleDisturbances(p.session.Disturbances, p.startNS())
	return options
}

func (p *connectPlan) startNS() int64 {
	if p.pacer == nil {
		return model.DefaultStartTimeNS
	}
	return p.pacer.StartNS()
}

func joined(served, ended error) error {
	if err := errors.Join(served, ended); err != nil {
		return fmt.Errorf("plantwire connect: %w", err)
	}
	return nil
}
