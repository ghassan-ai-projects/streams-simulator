// Package app runs one plant wire node session: handshake, births, slices with their barrier, and host records in order.
package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/uds"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

type Link interface {
	Send(records []map[string]any) error
	Await(ctx context.Context, until func(record map[string]any) bool) (inbound []domain.Inbound, met bool, err error)
}

type Options struct {
	Environment     string
	ProtocolMax     int
	ContractVersion string
	SessionID       string
	SliceNS         int64
	UntilNS         int64
	BDSeq           uint64
	PriorLastSeq    int64
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
	run      *run.Run
	link     Link
	options  Options
	out      *domain.Transmitter
	in       domain.Receiver
	devices  *devices
	epoch    int64
	bdSeq    uint64
	missed   int
	events   []model.SimEvent
	finished bool
}

func NewSession(r *run.Run, link Link, options Options) *Session {
	s := &Session{run: r, link: link, options: options, out: domain.NewTransmitter(options.SessionID), devices: newDevices(r), bdSeq: options.BDSeq}
	r.SetDeliveryObserver(func(event model.SimEvent) { s.events = append(s.events, event) })
	return s
}

func (s *Session) Run(ctx context.Context) error {
	if err := validate(s.options); err != nil {
		return err
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
	if err := s.send(s.nodeBirth()); err != nil {
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

func (s *Session) adopt(state map[string]any) error {
	version, err := domain.Negotiate(state, s.options.ProtocolMax)
	if err != nil {
		return fmt.Errorf("handshake: %w", err)
	}
	if s.in.Accept(domain.Seq(state["seq"])) != domain.InOrder {
		return fmt.Errorf("%w: the first host record has seq %v", ErrHostGap, state["seq"])
	}
	s.out.UseVersion(version)
	s.epoch = domain.Int(state["authority_epoch"])
	return nil
}

func (s *Session) awaitState(ctx context.Context) (map[string]any, error) {
	inbound, met, err := s.link.Await(ctx, isType("STATE"))
	if err != nil || !met || len(inbound) != 1 || inbound[0].Invalid != nil {
		return nil, fmt.Errorf("plantwire: no valid STATE first after NBIRTH: %w", errors.Join(err, ErrHostLost))
	}
	return inbound[0].Record, nil
}

func (s *Session) send(records ...domain.Record) error {
	for _, record := range records {
		s.out.Queue(record)
	}
	if err := s.link.Send(s.out.Drain()); err != nil {
		return fmt.Errorf("plantwire: send: %w", err)
	}
	return nil
}

func (s *Session) close() error {
	return s.send(domain.NodeDeath(s.bdSeq, "shutdown"))
}

func isType(kind string) func(map[string]any) bool {
	return func(record map[string]any) bool { return record["type"] == kind }
}

func Serve(ctx context.Context, r *run.Run, socketPath string, barrierCap time.Duration, options Options) error {
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
