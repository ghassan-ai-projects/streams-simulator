package app

import (
	"cmp"
	"errors"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

func (s *Session) execute(c domain.Command) error {
	answer := s.invoke(c, c.Key)
	return s.answerAll(c, s.ledger.Settle(c, answer), answer)
}

func (s *Session) executeSafeStop(c domain.Command) error {
	if reason := cmp.Or(s.identify(c), s.onBoot(c)); reason != "" {
		return s.answer(c.ID, c, refused(reason))
	}
	s.safeStops++
	return s.answer(c.ID, c, s.invoke(c, fmt.Sprintf("%s#%d", c.ID, s.safeStops)))
}

func (s *Session) invoke(c domain.Command, worldID string) domain.Answer {
	entity, _ := s.devices.entity(c.Device)
	name, args, err := s.devices.rawCommand(entity, c.Name, c.Args)
	if err != nil {
		return refused("out_of_range")
	}
	result, err := s.run.InvokeEffector(name, entity, worldID, args, s.now())
	return outcome(result, err)
}

func (s *Session) answerAll(c domain.Command, waiting []string, answer domain.Answer) error {
	for _, commandID := range waiting {
		if err := s.answer(commandID, c, answer); err != nil {
			return err
		}
		answer.Silent = false
	}
	return nil
}

func outcome(result *world.InvokeResult, err error) domain.Answer {
	switch {
	case errors.Is(err, world.ErrInterlockRefused):
		return refused("interlocked")
	case errors.Is(err, world.ErrInvalidInvocation):
		return refused("out_of_range")
	case errors.Is(err, world.ErrCommandIDReused):
		return refused("duplicate")
	case err != nil:
		return refused("not_ready")
	}
	return acknowledged(result)
}

func acknowledged(result *world.InvokeResult) domain.Answer {
	switch {
	case result.Mode == world.ModeAckLost:
		return domain.Answer{Status: "accepted", Outcome: "succeeded", Silent: true}
	case result.Accepted:
		return domain.Answer{Status: "accepted", Outcome: "succeeded"}
	case result.Cause == world.ReasonPreconditionRefused:
		return refused("permissive_not_met")
	}
	return refused("not_ready")
}

func (s *Session) runQueue() error {
	due := s.queued[:0:0]
	for _, c := range s.queued {
		if c.NotBeforeNS > s.now() {
			due = append(due, c)
			continue
		}
		if err := s.executeQueued(c); err != nil {
			return err
		}
	}
	s.queued = due
	return nil
}

func (s *Session) executeQueued(c domain.Command) error {
	if reason := s.admit(c); reason != "" {
		return s.answerAll(c, s.ledger.Release(c), refused(reason))
	}
	return s.execute(c)
}
