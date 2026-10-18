package app

import (
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
)

const (
	maxQueued  = 1024
	maxLedger  = 65536
	maxWaiting = 64
)

func (s *Session) command(m domain.Message) error {
	c := domain.CommandOf(m)
	switch {
	case c.Content() == "":
		return s.answer(c.ID, c, refused("malformed"))
	case s.safeStop(c):
		return s.executeSafeStop(c)
	}
	return s.checked(c)
}

func (s *Session) checked(c domain.Command) error {
	if reason := s.admit(c); reason != "" {
		return s.answer(c.ID, c, refused(reason))
	}
	if entry, conflict := s.ledger.Recall(c); conflict || entry != nil {
		return s.repeat(c, entry)
	}
	return s.perform(c)
}

func (s *Session) repeat(c domain.Command, entry *domain.Entry) error {
	switch {
	case entry == nil:
		return s.answer(c.ID, c, refused("duplicate"))
	case entry.Pending && len(entry.Waiting) >= maxWaiting:
		return s.answer(c.ID, c, refused("not_ready"))
	case entry.Pending:
		entry.Waiting = append(entry.Waiting, c.ID)
		return s.send(domain.Receipt(c.ID, "queued", "", s.boot(c)))
	}
	return s.answer(c.ID, c, replayed(entry.Answer))
}

func (s *Session) perform(c domain.Command) error {
	if len(s.ledger) >= maxLedger {
		return s.answer(c.ID, c, refused("not_ready"))
	}
	if c.NotBeforeNS > s.now() {
		return s.hold(c)
	}
	return s.execute(c)
}

func (s *Session) hold(c domain.Command) error {
	if len(s.queued) >= maxQueued {
		return s.answer(c.ID, c, refused("not_ready"))
	}
	s.ledger.Hold(c)
	s.queued = append(s.queued, c)
	return s.send(domain.Receipt(c.ID, "queued", "", s.boot(c)))
}

func (s *Session) admit(c domain.Command) string {
	for _, check := range []func(domain.Command) string{s.identify, s.fresh, s.onBoot, s.authorize} {
		if reason := check(c); reason != "" {
			return reason
		}
	}
	return ""
}

func (s *Session) identify(c domain.Command) string {
	entity, known := s.devices.entity(c.Device)
	switch {
	case !known:
		return "unknown_device"
	case s.run.World.EffectorOf(entity, c.Name) == nil:
		return "unknown_command"
	case s.run.World.CheckEffectorArgs(c.Name, c.Args) != nil:
		return "out_of_range"
	}
	return ""
}

func (s *Session) fresh(c domain.Command) string {
	if c.NotBeforeNS > s.now()+domain.MaxDeadlineAheadNS {
		return "out_of_range"
	}
	return domain.Freshness(c, s.now())
}

func (s *Session) onBoot(c domain.Command) string {
	if c.Boot != s.boot(c) {
		return "wrong_boot"
	}
	return ""
}

func (s *Session) authorize(c domain.Command) string {
	switch {
	case c.Epoch != s.epoch:
		return "stale_epoch"
	case !slices.Contains(s.options.Policies, c.Policy):
		return "policy_not_allowed"
	}
	return ""
}

func (s *Session) safeStop(c domain.Command) bool {
	entity, known := s.devices.entity(c.Device)
	if !known {
		return false
	}
	effector := s.run.World.EffectorOf(entity, c.Name)
	return effector != nil && effector.SafeStop
}

func (s *Session) boot(c domain.Command) string {
	boot, _ := s.devices.bootOf(c.Device)
	return boot
}

func (s *Session) now() int64 { return s.run.Status().ClockNS }

func (s *Session) answer(commandID string, c domain.Command, answer domain.Answer) error {
	if answer.Status == "" || answer.Silent {
		return nil
	}
	records := []domain.Record{domain.Receipt(commandID, answer.Status, answer.Reason, s.boot(c))}
	if answer.Outcome != "" {
		records = append(records, domain.Result(commandID, answer.Outcome, s.boot(c)))
	}
	return s.send(records...)
}

func refused(reason string) domain.Answer { return domain.Answer{Status: "refused", Reason: reason} }

func replayed(answer domain.Answer) domain.Answer {
	answer.Silent = false
	return answer
}
