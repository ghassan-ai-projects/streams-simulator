package app

import (
	"errors"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
)

func (s *Session) prepare() error {
	if err := errors.Join(validate(s.options), s.devices.safeStopsUnguarded()); err != nil {
		return fmt.Errorf("plantwire: %w", err)
	}
	for _, disturbance := range s.options.Disturbances {
		if err := s.inject(disturbance); err != nil {
			return fmt.Errorf("plantwire: disturbance %s on %s: %w", disturbance.Fault, disturbance.Entity, err)
		}
	}
	return nil
}

func (s *Session) inject(disturbance domain.ScheduledFault) error {
	id, err := s.run.InjectFault(disturbance.Entity, disturbance.Fault, disturbance.OnsetNS, disturbance.Params)
	if err != nil {
		return fmt.Errorf("inject: %w", err)
	}
	if disturbance.ClearNS == 0 {
		return nil
	}
	if err := s.run.ClearFault(id, disturbance.ClearNS); err != nil {
		return fmt.Errorf("clear: %w", err)
	}
	return nil
}
