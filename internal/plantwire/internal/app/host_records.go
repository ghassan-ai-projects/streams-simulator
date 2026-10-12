package app

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
)

func (s *Session) process(inbound []domain.Inbound) error {
	for _, arrival := range inbound {
		switch s.in.Accept(domain.Seq(arrival.Record["seq"])) {
		case domain.Duplicate:
			continue
		case domain.Gap:
			return fmt.Errorf("%w: seq %v", ErrHostGap, arrival.Record["seq"])
		}
		if err := s.handle(arrival); err != nil {
			return err
		}
	}
	return nil
}

func (s *Session) handle(arrival domain.Inbound) error {
	record := arrival.Record
	if arrival.Invalid != nil {
		return s.refuseMalformed(record)
	}
	switch record["type"] {
	case "STATE":
		s.hostState(record)
	case "NCMD":
		return s.rebirth()
	}
	return s.deviceRequest(record)
}

func (s *Session) deviceRequest(record map[string]any) error {
	switch record["type"] {
	case "DQUERY":
		return s.send(s.devices.state(record["device"].(string))...)
	case "DCMD":
		return s.command(record)
	}
	return nil
}

func (s *Session) refuseMalformed(record map[string]any) error {
	commandID, named := record["command_id"].(string)
	if record["type"] != "DCMD" || !named || commandID == "" {
		return nil
	}
	return s.send(domain.Receipt(commandID, "refused", "malformed", "none"))
}

func (s *Session) command(record map[string]any) error {
	boot, known := s.devices.bootOf(record["device"].(string))
	reason := "not_ready"
	if !known {
		reason = "unknown_device"
	}
	return s.send(domain.Receipt(record["command_id"].(string), "refused", reason, boot))
}

func (s *Session) hostState(record map[string]any) {
	epoch := domain.Int(record["authority_epoch"])
	if epoch < s.epoch {
		return
	}
	s.epoch = epoch
	if online, _ := record["online"].(bool); !online {
		s.finished = true
	}
}

func (s *Session) rebirth() error {
	s.bdSeq++
	return s.send(append([]domain.Record{s.nodeBirth()}, s.devices.births(s.run.Status().ClockNS)...)...)
}

func (s *Session) nodeBirth() domain.Record {
	identity := domain.NodeIdentity{Environment: s.options.Environment, ProtocolMax: s.options.ProtocolMax,
		ContractVersion: s.options.ContractVersion, BDSeq: s.bdSeq, PriorLastSeq: s.options.PriorLastSeq}
	return domain.NodeBirth(identity, s.devices.kinds(), s.devices.boards())
}
