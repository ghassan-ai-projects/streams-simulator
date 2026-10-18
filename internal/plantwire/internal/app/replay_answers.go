package app

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
)

func (p *replayer) answer(inbound []domain.Inbound) error {
	for _, arrival := range inbound {
		if err := p.react(arrival); err != nil {
			return err
		}
	}
	return nil
}

func (p *replayer) react(in domain.Inbound) error {
	in.Within(p.session, p.version)
	if in.Unsequenced {
		return nil
	}
	switch p.in.Accept(in.Seq) {
	case domain.Duplicate:
		return nil
	case domain.Gap:
		return fmt.Errorf("plantwire replay: %w: seq %d", ErrHostGap, in.Seq)
	}
	return p.handle(in)
}

func (p *replayer) handle(in domain.Inbound) error {
	switch {
	case in.Type == "DCMD" && in.Invalid != nil:
		return p.refuse(in.CommandID, "malformed", "none")
	case in.Invalid != nil:
		return nil
	case in.Type == "STATE":
		p.over = p.over || !in.Online
	case in.Type == "DCMD":
		reason, boot := p.refusal(in.Device)
		return p.refuse(in.CommandID, reason, boot)
	}
	return nil
}

func (p *replayer) refusal(device string) (reason, boot string) {
	if recorded, known := p.plan.Boots[device]; known {
		return "not_ready", recorded
	}
	return "unknown_device", "none"
}

func (p *replayer) refuse(commandID, reason, boot string) error {
	if commandID == "" || domain.WireID(commandID) != commandID {
		return nil
	}
	p.stats.Refused++
	return p.send(domain.Receipt(commandID, "refused", reason, boot))
}
