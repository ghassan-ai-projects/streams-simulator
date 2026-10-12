package app

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
)

func (s *Session) process(inbound []domain.Inbound) error {
	for _, arrival := range inbound {
		if err := s.arrive(arrival); err != nil {
			return err
		}
	}
	return nil
}

func (s *Session) arrive(arrival domain.Inbound) error {
	s.run.JournalWire("host", arrival.Record)
	fresh, err := s.inOrder(&arrival)
	if err != nil || !fresh {
		return err
	}
	return s.handle(arrival)
}

func (s *Session) inOrder(arrival *domain.Inbound) (bool, error) {
	arrival.Within(s.options.SessionID, s.version)
	if arrival.Unsequenced {
		return true, nil
	}
	switch s.in.Accept(arrival.Seq) {
	case domain.Duplicate:
		return false, nil
	case domain.Gap:
		return false, fmt.Errorf("%w: seq %d", ErrHostGap, arrival.Seq)
	}
	return true, nil
}

func (s *Session) handle(arrival domain.Inbound) error {
	if arrival.Invalid != nil {
		return s.refuseMalformed(arrival.Message)
	}
	switch arrival.Type {
	case "STATE":
		s.hostState(arrival.Message)
	case "NCMD":
		return s.rebirth()
	}
	return s.deviceRequest(arrival.Message)
}

func (s *Session) deviceRequest(m domain.Message) error {
	switch m.Type {
	case "DQUERY":
		return s.send(s.devices.state(m.Device)...)
	case "DCMD":
		return s.command(m)
	}
	return nil
}

func (s *Session) refuseMalformed(m domain.Message) error {
	if m.Type != "DCMD" || m.CommandID == "" || domain.WireID(m.CommandID) != m.CommandID {
		return nil
	}
	return s.send(domain.Receipt(m.CommandID, "refused", "malformed", "none"))
}

func (s *Session) hostState(m domain.Message) {
	if m.AuthorityEpoch < s.epoch {
		return
	}
	s.epoch = m.AuthorityEpoch
	if !m.Online {
		s.finished = true
	}
}

func (s *Session) rebirth() error {
	s.bdSeq++
	return s.send(append([]domain.Record{s.nodeBirth()}, s.devices.births(s.run.Status().ClockNS)...)...)
}

func (s *Session) nodeBirth() domain.Record {
	identity := domain.NodeIdentity{Environment: s.options.Environment, ProtocolMax: s.options.ProtocolMax,
		BDSeq: s.bdSeq, PriorLastSeq: s.options.PriorLastSeq, SafeStops: s.options.SafeStop}
	return domain.NodeBirth(identity, s.devices.kinds(), s.devices.boards())
}
