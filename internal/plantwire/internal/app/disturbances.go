package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
)

var errNoEntity = errors.New("the world has no such entity")

func (s *Session) prepare() error {
	if err := errors.Join(validate(s.options), s.devices.safeStopsUnguarded()); err != nil {
		return fmt.Errorf("plantwire: %w", err)
	}
	for _, disturbance := range s.options.Disturbances {
		if s.run.World.Entity(disturbance.Entity) == nil {
			return fmt.Errorf("plantwire: disturbance %s on %s: %w", disturbance.Fault, disturbance.Entity, errNoEntity)
		}
	}
	return nil
}

func (s *Session) anchor(ctx context.Context) error {
	start := s.run.Status().ClockNS
	if s.options.Pace != nil {
		if err := s.skipTo(ctx, domain.AttachedSlice(start, s.options.SliceNS, s.options.Pace.WorldNowNS())); err != nil {
			return err
		}
	}
	shift := s.run.Status().ClockNS - start
	for _, disturbance := range s.options.Disturbances {
		if err := s.inject(disturbance, shift); err != nil {
			return fmt.Errorf("plantwire: disturbance %s on %s: %w", disturbance.Fault, disturbance.Entity, err)
		}
	}
	return nil
}

func (s *Session) skipTo(ctx context.Context, attachedNS int64) error {
	if attachedNS <= s.run.Status().ClockNS {
		return nil
	}
	if _, err := s.run.Advance(ctx, attachedNS, false); err != nil {
		return fmt.Errorf("plantwire: advance to the host's attach: %w", err)
	}
	s.events = nil
	return nil
}

func (s *Session) inject(disturbance domain.ScheduledFault, shift int64) error {
	id, err := s.run.InjectFault(disturbance.Entity, disturbance.Fault, disturbance.OnsetNS+shift, disturbance.Params)
	if err != nil {
		return fmt.Errorf("inject: %w", err)
	}
	if disturbance.ClearNS == 0 {
		return nil
	}
	if err := s.run.ClearFault(id, disturbance.ClearNS+shift); err != nil {
		return fmt.Errorf("clear: %w", err)
	}
	return nil
}
