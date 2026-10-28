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
	Drain(ctx context.Context, wait time.Duration) ([]domain.Inbound, error)
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
	Bound        []domain.Bound

	NoWireJournal  bool
	JournalOrigins bool

	Control      *ControlPort
	InterruptLog InterruptLog
	Pace         Pacer
	Disturbances []domain.ScheduledFault
	Interfaces   domain.Interfaces
}

type Pacer interface {
	WaitNS(worldNS int64) int64
	WorldNowNS() int64
}

var (
	ErrBadOptions = errors.New("plantwire: invalid node options")
	ErrHostPolicy = errors.New("plantwire: a host that issues commands carries a policy digest")
)

func validate(o Options) error {
	switch {
	case o.SliceNS <= 0 || o.UntilNS <= 0:
		return fmt.Errorf("%w: a slice and an end time are required", ErrBadOptions)
	case o.ProtocolMax < domain.Version:
		return fmt.Errorf("%w: protocol_max %d", ErrBadOptions, o.ProtocolMax)
	case o.Environment != domain.EnvironmentSimulated:
		return fmt.Errorf("%w: environment %q, a simulated world announces only %q", ErrBadOptions, o.Environment, domain.EnvironmentSimulated)
	case domain.WireID(o.SessionID) != o.SessionID:
		return fmt.Errorf("%w: session %q is no wire id", ErrBadOptions, o.SessionID)
	}
	return policyOptions(o.Policies)
}

func policyOptions(digests []string) error {
	if _, issues := domain.NewAllowList(digests, "/policy"); len(issues) > 0 {
		return fmt.Errorf("%w: %s", ErrBadOptions, issues[0].Message)
	}
	return nil
}

func allowListOf(digests []string) domain.AllowList {
	allowed, _ := domain.NewAllowList(digests, "/policy")
	return allowed
}

var (
	ErrHostLost = errors.New("plantwire: host lost")
	ErrHostGap  = errors.New("plantwire: host records out of sequence")
)

type Session struct {
	run         *run.Run
	link        Link
	options     Options
	out         *domain.Transmitter
	in          domain.Receiver
	devices     *devices
	epoch       int64
	bdSeq       uint64
	events      []model.SimEvent
	finished    bool
	ledger      domain.Ledger
	queued      []domain.Command
	version     int
	safeStops   int
	live        *domain.Live
	liveChanged bool
	ints        *interruptSide
	allowed     domain.AllowList
	held        []domain.Inbound
}

func NewSession(r *run.Run, link Link, options Options) *Session {
	s := &Session{run: r, link: link, options: options, out: domain.NewTransmitter(options.SessionID), devices: newDevices(r, boundInstances(options.Bound), options.Interfaces), bdSeq: options.BDSeq, ledger: domain.Ledger{}, allowed: allowListOf(options.Policies)}
	if options.JournalOrigins && !options.NoWireJournal {
		s.live = domain.NewLive(false, s.devices.instances(), options.Bound)
	}
	if options.Control != nil {
		s.ints = &interruptSide{gate: domain.NewInterruptGate(), port: options.Control, log: options.InterruptLog}
	}
	r.SetDeliveryObserver(func(event model.SimEvent) { s.events = append(s.events, event) })
	return s
}

func (s *Session) Run(ctx context.Context) error {
	defer s.endLive()
	if err := s.prepare(); err != nil {
		return err
	}
	if err := s.handshake(ctx); err != nil {
		return err
	}
	for at := s.run.Status().ClockNS + s.options.SliceNS; at <= s.options.UntilNS && !s.finished; at += s.options.SliceNS {
		if err := s.step(ctx, at); err != nil {
			return err
		}
	}
	return s.close()
}

func (s *Session) handshake(ctx context.Context) error {
	if err := s.sendNodeBirth(); err != nil {
		return err
	}
	if err := s.attach(ctx); err != nil {
		return err
	}
	return s.send(s.devices.births(s.run.Status().ClockNS)...)
}

func (s *Session) attach(ctx context.Context) error {
	state, err := s.awaitState(ctx)
	if err != nil {
		return err
	}
	if err := s.adopt(state); err != nil {
		return err
	}
	return s.anchor(ctx)
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
		s.journalWire("host", arrival.Record)
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
		s.journalWire("node", record)
	}
	if err := s.link.Send(wireRecords); err != nil {
		return fmt.Errorf("plantwire: send: %w", err)
	}
	return nil
}

func (s *Session) journalWire(direction string, record map[string]any) {
	if !s.options.NoWireJournal {
		s.run.JournalWire(direction, record)
	}
	s.journalLive(direction, record)
	if s.options.JournalOrigins && direction == "node" && record["type"] == "DBIRTH" {
		device, _ := record["device"].(string)
		s.run.JournalOrigin(domain.OriginNote(record, domain.EnvironmentSimulated, s.devices.ids[device]).Fields)
	}
}

func (s *Session) close() error {
	if s.ints != nil {
		s.recordInterrupts(s.ints.gate.Flush(domain.ClockWorld, s.now()).Entries)
	}
	return s.send(domain.NodeDeath(s.bdSeq, "shutdown"))
}

func isType(kind string) func(domain.Message) bool {
	return func(m domain.Message) bool { return m.Type == kind }
}

func Serve(ctx context.Context, r *run.Run, socketPath string, barrierCap time.Duration, options Options) error {
	if err := policyOptions(options.Policies); err != nil {
		return fmt.Errorf("plantwire: %w", err)
	}
	_, faces, err := loadInterfaces(r.Domain().Spec, options.KindsDirs)
	if err != nil {
		return fmt.Errorf("plantwire: %w", model.WithSource(err, r.Domain().Spec.ID))
	}
	options.Interfaces = faces
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

func CheckHost(options host.Options) error {
	if len(options.Script) == 0 {
		return nil
	}
	if _, issues := domain.NewAllowList([]string{options.Policy}, "/policy"); len(issues) > 0 {
		return fmt.Errorf("%w: %s", ErrHostPolicy, issues[0].Message)
	}
	return nil
}

func ServeHost(ctx context.Context, socketPath string, delay time.Duration, options host.Options) (host.Report, error) {
	if err := CheckHost(options); err != nil {
		return host.Report{}, fmt.Errorf("plantwire: host: %w", err)
	}
	if err := domain.CheckEnvironment(options.Environment); err != nil {
		return host.Report{}, fmt.Errorf("plantwire: host: %w", err)
	}
	reference := host.New(options)
	err := uds.ServeHost(ctx, socketPath, reference, hostPacing(delay, options), options.Ready)
	if err != nil && !endsHostCleanly(err) {
		return reference.Report(), fmt.Errorf("plantwire: %w", err)
	}
	return reference.Report(), nil
}

func hostPacing(delay time.Duration, options host.Options) uds.Pacing {
	return uds.Pacing{Delay: delay, Handshake: uds.HandshakeWait, Heartbeat: host.HeartbeatInterval(options)}
}

func endsHostCleanly(err error) bool {
	return uds.Interrupted(err) || uds.PeerGone(err)
}
