// Package app runs one plant wire node session: handshake, births, slices with their barrier, and host records in order.
package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/host"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/uds"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

type Link interface {
	Send(records []map[string]any) error
	Await(ctx context.Context, until func(domain.Message) bool) (inbound []domain.Inbound, met bool, err error)
	AwaitHandshake(ctx context.Context, until func(domain.Message) bool) (inbound []domain.Inbound, met bool, err error)
	WaitedMS() int64
}

type Options struct {
	Environment  string
	ProtocolMax  int
	SessionID    string
	SliceNS      int64
	UntilNS      int64
	BDSeq        uint64
	PriorLastSeq int64
	Policies     []string
	KindsDirs    []string
	WireFaults   []domain.WireFault
}

var ErrBadOptions = errors.New("plantwire: invalid node options")

func validate(o Options) error {
	switch {
	case o.SliceNS <= 0 || o.UntilNS <= 0:
		return fmt.Errorf("%w: a slice and an end time are required", ErrBadOptions)
	case o.ProtocolMax < domain.Version:
		return fmt.Errorf("%w: protocol_max %d", ErrBadOptions, o.ProtocolMax)
	case o.Environment != "simulated" && o.Environment != "hardware", domain.WireID(o.SessionID) != o.SessionID:
		return fmt.Errorf("%w: environment %q, session %q", ErrBadOptions, o.Environment, o.SessionID)
	}
	return nil
}

var (
	ErrHostLost = errors.New("plantwire: host lost")
	ErrHostGap  = errors.New("plantwire: host records out of sequence")
)

type Session struct {
	run       *run.Run
	link      Link
	options   Options
	out       *domain.Transmitter
	in        domain.Receiver
	devices   *devices
	epoch     int64
	bdSeq     uint64
	missed    int
	events    []model.SimEvent
	finished  bool
	ledger    domain.Ledger
	queued    []domain.Command
	version   int
	safeStops int
}

func NewSession(r *run.Run, link Link, options Options) *Session {
	s := &Session{run: r, link: link, options: options, out: domain.NewTransmitter(options.SessionID), devices: newDevices(r), bdSeq: options.BDSeq, ledger: domain.Ledger{}}
	r.SetDeliveryObserver(func(event model.SimEvent) { s.events = append(s.events, event) })
	return s
}

func (s *Session) Run(ctx context.Context) error {
	if err := errors.Join(validate(s.options), s.devices.safeStopsUnguarded()); err != nil {
		return fmt.Errorf("plantwire: %w", err)
	}
	if err := s.handshake(ctx); err != nil {
		return err
	}
	for at := s.run.Status().ClockNS + s.options.SliceNS; at <= s.options.UntilNS && !s.finished; at += s.options.SliceNS {
		if err := s.slice(ctx, at); err != nil {
			return err
		}
	}
	return s.close()
}

func (s *Session) handshake(ctx context.Context) error {
	if err := s.sendNodeBirth(); err != nil {
		return err
	}
	state, err := s.awaitState(ctx)
	if err != nil {
		return err
	}
	if err := s.adopt(state); err != nil {
		return err
	}
	return s.send(s.devices.births(s.run.Status().ClockNS)...)
}

func (s *Session) adopt(state domain.Message) error {
	version, err := domain.Negotiate(state, s.options.ProtocolMax)
	if err != nil {
		return fmt.Errorf("handshake: %w", err)
	}
	if s.in.Accept(state.Seq) != domain.InOrder {
		return fmt.Errorf("%w: the first host record has seq %d", ErrHostGap, state.Seq)
	}
	s.out.UseVersion(version)
	s.version, s.epoch = version, state.AuthorityEpoch
	return nil
}

func (s *Session) awaitState(ctx context.Context) (domain.Message, error) {
	inbound, met, err := s.link.AwaitHandshake(ctx, isType("STATE"))
	for _, arrival := range inbound {
		s.run.JournalWire("host", arrival.Record)
	}
	if err != nil || !met || len(inbound) != 1 || inbound[0].Invalid != nil {
		return domain.Message{}, fmt.Errorf("plantwire: no valid STATE first after NBIRTH: %w", errors.Join(err, ErrHostLost))
	}
	return inbound[0].Message, nil
}

func (s *Session) send(records ...domain.Record) error {
	for _, record := range records {
		s.out.Queue(record)
	}
	wireRecords := domain.ApplyWireFaults(s.out.Drain(), s.options.WireFaults)
	for _, record := range wireRecords {
		s.run.JournalWire("node", record)
	}
	if err := s.link.Send(wireRecords); err != nil {
		return fmt.Errorf("plantwire: send: %w", err)
	}
	return nil
}

func (s *Session) close() error {
	return s.send(domain.NodeDeath(s.bdSeq, "shutdown"))
}

func isType(kind string) func(domain.Message) bool {
	return func(m domain.Message) bool { return m.Type == kind }
}

func Serve(ctx context.Context, r *run.Run, socketPath string, barrierCap time.Duration, options Options) error {
	if _, err := CheckImplements(r.Domain().Spec, options.KindsDirs); err != nil {
		return fmt.Errorf("plantwire: %w", model.WithSource(err, r.Domain().Spec.ID))
	}
	return serveSession(ctx, r, socketPath, barrierCap, options)
}

func serveSession(ctx context.Context, r *run.Run, socketPath string, barrierCap time.Duration, options Options) error {
	ctx, stop := uds.UntilInterrupted(ctx)
	defer stop()
	link, err := uds.Dial(ctx, socketPath, barrierCap)
	if err != nil {
		return fmt.Errorf("plantwire: %w", err)
	}
	served := NewSession(r, link, options).Run(ctx)
	if err := errors.Join(served, link.Close()); err != nil {
		return fmt.Errorf("plantwire: %w", err)
	}
	return nil
}

func ServeHost(ctx context.Context, socketPath string, delay time.Duration, options host.Options) (host.Report, error) {
	reference := host.New(options)
	if err := uds.ServeHost(ctx, socketPath, reference, uds.Pacing{Delay: delay, Handshake: uds.HandshakeWait}, options.Ready); err != nil {
		return reference.Report(), fmt.Errorf("plantwire: %w", err)
	}
	return reference.Report(), nil
}
