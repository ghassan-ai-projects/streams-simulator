package app

import (
	"errors"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
)

type interruptSide struct {
	gate *domain.InterruptGate
	port *ControlPort
	log  InterruptLog
}

func (s *Session) judged(in domain.Inbound, c domain.Command) error {
	outcome := s.ints.gate.Judge(in, domain.ClockWorld, s.now())
	if err := s.applyGate(outcome); err != nil || !outcome.Proceed {
		return err
	}
	return s.dispatch(c)
}

func (s *Session) interruptsAtBarrier() error {
	if s.ints == nil {
		return nil
	}
	s.recordInterrupts(s.ints.gate.SetInventory(s.devices.interruptInventory(), domain.ClockWorld, s.now()).Entries)
	for call, ok := s.ints.port.take(); ok; call, ok = s.ints.port.take() {
		if err := s.control(call); err != nil {
			return err
		}
	}
	return s.applyGate(s.ints.gate.Advance(domain.ClockWorld, s.now()))
}

func (s *Session) control(call ControlCall) error {
	outcome := s.ints.gate.Control(call.Entry, domain.ClockWorld, s.now())
	err := s.applyGate(outcome)
	if outcome.Reply["result"] == domain.ResultApplied {
		err = errors.Join(err, s.purgeQueued(call.Entry))
	}
	call.Reply <- domain.ControlReply(outcome.Reply)
	return err
}

func (s *Session) purgeQueued(request domain.InterruptEntry) error {
	var failed error
	kept := s.queued[:0:0]
	for _, c := range s.queued {
		if s.ints.gate.Covers(request, c) {
			failed = errors.Join(failed, s.refuseQueued(c))
		} else {
			kept = append(kept, c)
		}
	}
	s.queued = kept
	return failed
}

func (s *Session) refuseQueued(c domain.Command) error {
	var failed error
	for _, id := range s.ledger.Release(c) {
		s.recordInterrupts([]domain.InterruptEntry{domain.PurgeEntry(id, c.Device, domain.ReasonLocalControl, domain.ClockWorld, s.now())})
		failed = errors.Join(failed, s.send(domain.Receipt(id, "refused", domain.ReasonLocalControl, "none")))
	}
	return failed
}

func (s *Session) applyGate(o domain.GateOutcome) error {
	s.recordInterrupts(o.Entries)
	return errors.Join(s.performStops(o.Stops), s.answerGate(o))
}

func (s *Session) answerGate(o domain.GateOutcome) error {
	for _, queued := range o.Queued {
		receipt := domain.Receipt(queued.CommandID, "queued", "", s.boot(domain.Command{Device: queued.Device}))
		if err := s.send(receipt); err != nil {
			return err
		}
	}
	for _, denied := range o.Refused {
		if err := s.send(domain.Receipt(denied.CommandID, "refused", denied.Reason, "none")); err != nil {
			return err
		}
	}
	return s.forwardApproved(o.Forward)
}

func (s *Session) forwardApproved(commands []domain.Inbound) error {
	for _, in := range commands {
		if err := s.dispatch(domain.CommandOf(in.Message)); err != nil {
			return err
		}
	}
	return nil
}

func (s *Session) performStops(orders []domain.StopOrder) error {
	for _, order := range orders {
		s.performStop(order)
	}
	var failed error
	for _, order := range orders {
		failed = errors.Join(failed, s.send(s.devices.state(order.Device)...))
	}
	return failed
}

func (s *Session) performStop(order domain.StopOrder) {
	s.safeStops++
	id := fmt.Sprintf("node-stop-%d", s.safeStops)
	answer := s.invoke(domain.Command{Device: order.Device, Name: order.Command, Args: map[string]any{}}, id)
	s.recordInterrupts([]domain.InterruptEntry{domain.EffectEntry(id, order.Device, stopOutcome(answer), domain.ClockWorld, s.now())})
}

func stopOutcome(answer domain.Answer) string {
	if answer.Reason != "" {
		return answer.Status + ":" + answer.Reason
	}
	return answer.Status
}

func (s *Session) recordInterrupts(entries []domain.InterruptEntry) {
	recordInterrupts(entries, s.ints.log, s.run.JournalInterrupt)
	if domain.ChangesInterrupts(entries) {
		view := s.ints.gate.View()
		s.ints.port.Publish(view)
		if s.live != nil && s.live.SetInterrupts(view) {
			s.liveChanged = true
		}
	}
}
