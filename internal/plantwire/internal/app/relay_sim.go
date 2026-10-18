package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	simdomain "github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/uds"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

type SimulatedOptions struct {
	DomainsDir  string
	AdaptersDir string
	Domain      string
	Adapter     string
	Seed        uint64
	BarrierCap  time.Duration
	Node        Options
}

type simNode struct {
	run       *run.Run
	link      *uds.SimLink
	birth     domain.Inbound
	instances map[string]string
	bound     []domain.Bound
	begin     func()
	done      <-chan error
	cancel    context.CancelFunc
	started   bool
	failure   error
}

func openSim(ctx context.Context, options RelayOptions) (*simNode, error) {
	if options.Simulated == nil {
		return nil, nil
	}
	r, err := newSimRun(ctx, *options.Simulated, options.Out)
	if err != nil {
		return nil, err
	}
	node, err := newSimNode(ctx, r, *options.Simulated)
	if err != nil {
		discardRun(r)
		return nil, err
	}
	return node, nil
}

func discardRun(r *run.Run) {
	r.Abandon("plantwire relay: the simulated node could not be built")
	_, _ = r.End("")
}

func newSimRun(ctx context.Context, o SimulatedOptions, out string) (*run.Run, error) {
	spec, adapterSpec, err := loadSimInputs(o)
	if err != nil {
		return nil, fmt.Errorf("plantwire relay: %w", err)
	}
	r, err := run.New(context.WithoutCancel(ctx), run.Config{Domain: spec, Adapter: adapterSpec, Seed: o.Seed, SinkName: model.SinkInproc,
		TimeMode: model.TimeStepped, StartTimeNS: model.DefaultStartTimeNS, JournalDir: out})
	if err != nil {
		return nil, fmt.Errorf("plantwire relay: %w", err)
	}
	return r, nil
}

func loadSimInputs(o SimulatedOptions) (*simdomain.Compiled, *model.Adapter, error) {
	spec, adapterSpec, err := run.LoadInputs(o.DomainsDir, o.AdaptersDir, o.Domain, o.Adapter)
	if err != nil {
		return nil, nil, fmt.Errorf("load simulated inputs: %w", err)
	}
	if _, err := CheckImplements(spec.Spec, o.Node.KindsDirs); err != nil {
		return nil, nil, fmt.Errorf("%w", model.WithSource(err, spec.Spec.ID))
	}
	return spec, adapterSpec, nil
}

func newSimNode(ctx context.Context, r *run.Run, o SimulatedOptions) (*simNode, error) {
	options := o.Node
	options.NoWireJournal = true
	link := uds.NewSimLink(o.BarrierCap)
	session := NewSession(r, link, options)
	birth, err := simBirth(session, options)
	if err != nil {
		return nil, fmt.Errorf("plantwire relay: %w", err)
	}
	ctx, cancel := context.WithCancel(ctx)
	node := &simNode{run: r, link: link, birth: birth, instances: session.devices.instances(), bound: options.Bound, cancel: cancel}
	node.arm(ctx, session)
	return node, nil
}

func (s *simNode) arm(ctx context.Context, session *Session) {
	s.begin = func() {
		s.done, s.started = uds.Spawn(func() error { return session.Run(ctx) }, s.link.Wake), true
	}
}

func simBirth(session *Session, options Options) (domain.Inbound, error) {
	if err := errors.Join(validate(options), session.devices.safeStopsUnguarded()); err != nil {
		return domain.Inbound{}, err
	}
	record, err := session.nodeBirth()
	if err != nil {
		return domain.Inbound{}, err
	}
	return domain.BirthInbound(record, options.SessionID)
}

func (s *simNode) discard() {
	if s == nil {
		return
	}
	s.cancel()
	s.link.Close()
	discardRun(s.run)
}

func (s *simNode) signal() <-chan struct{} {
	if s == nil {
		return nil
	}
	return s.link.UpReady()
}

func (s *simNode) reap() bool {
	select {
	case err := <-s.done:
		s.complete(err)
		return true
	default:
		return false
	}
}

func (s *simNode) deliver(records []map[string]any) {
	if s == nil || len(records) == 0 {
		return
	}
	for _, record := range records {
		in, err := domain.InboundOf(record)
		if err != nil || !s.link.Deliver(in) {
			slog.Warn("plantwire relay: a record for the simulated node was lost", "error", err)
		}
	}
	if !s.started {
		s.begin()
	}
}

func (s *simNode) stop() error {
	if s == nil {
		return nil
	}
	s.cancel()
	s.link.Close()
	if s.done != nil {
		s.complete(<-s.done)
	}
	return s.failure
}

func (s *simNode) complete(err error) {
	s.done = nil
	if err != nil && !endedByRelay(err) && s.failure == nil {
		slog.Warn("plantwire relay: the simulated node ended", "error", err)
		s.failure = err
	}
}

func endedByRelay(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, uds.ErrSimClosed)
}

func (s *simNode) end(out, fieldTrace string) error {
	if s == nil {
		return nil
	}
	s.run.SetFieldTrace(fieldTrace)
	if _, err := s.run.End(out); err != nil {
		return fmt.Errorf("plantwire relay: end the run: %w", err)
	}
	return nil
}
