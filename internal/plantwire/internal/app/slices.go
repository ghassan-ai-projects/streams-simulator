package app

import (
	"context"
	"fmt"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
)

func (s *Session) step(ctx context.Context, at int64) error {
	if s.options.Pace == nil {
		return s.slice(ctx, at)
	}
	if err := s.release(ctx, at); err != nil {
		return err
	}
	return s.slice(ctx, at)
}

func (s *Session) release(ctx context.Context, at int64) error {
	wait := s.options.Pace.WaitNS(at)
	if wait <= 0 {
		return nil
	}
	held, err := s.link.Drain(ctx, time.Duration(wait))
	s.held = append(s.held, held...)
	if err == nil {
		return nil
	}
	if ctx.Err() != nil {
		return fmt.Errorf("plantwire: pace: %w", err)
	}
	return s.hostLost(s.takeHeld(), fmt.Errorf("pace: %w", err))
}

func (s *Session) takeHeld() []domain.Inbound {
	held := s.held
	s.held = nil
	return held
}

func (s *Session) slice(ctx context.Context, at int64) error {
	if _, err := s.run.Advance(ctx, at, false); err != nil {
		return fmt.Errorf("plantwire: advance: %w", err)
	}
	if err := s.send(s.data()...); err != nil {
		return err
	}
	if err := s.send(domain.Sync(s.out.LastSeq(), at)); err != nil {
		return err
	}
	err := s.barrier(ctx, s.out.LastSeq())
	s.flushLive()
	return err
}

func (s *Session) data() []domain.Record {
	var records []domain.Record
	for _, event := range s.events {
		records = append(records, s.devices.observe(event)...)
	}
	s.events = nil
	return records
}

func (s *Session) barrier(ctx context.Context, syncSeq uint64) error {
	inbound, met, err := s.awaitSync(ctx, syncSeq)
	inbound = append(s.takeHeld(), inbound...)
	s.run.JournalPacing(syncSeq, s.link.WaitedMS(), !met)
	if err != nil {
		return s.hostLost(inbound, err)
	}
	if err := s.interruptsAtBarrier(); err != nil {
		return err
	}
	return s.settleBarrier(inbound)
}

func (s *Session) settleBarrier(inbound []domain.Inbound) error {
	if processed := s.process(inbound); processed != nil || s.finished {
		return processed
	}
	return s.runQueue()
}

func (s *Session) hostLost(inbound []domain.Inbound, cause error) error {
	for i := range inbound {
		s.journalWire("host", inbound[i].Record)
		if inbound[i].Within(s.options.SessionID, s.version); inbound[i].Invalid == nil && inbound[i].Type == "STATE" {
			s.hostState(inbound[i].Message)
		}
	}
	if s.finished {
		return nil
	}
	return fmt.Errorf("%w: %w", ErrHostLost, cause)
}

func (s *Session) awaitSync(ctx context.Context, syncSeq uint64) ([]domain.Inbound, bool, error) {
	inbound, met, err := s.link.Await(ctx, func(m domain.Message) bool {
		return m.Type == "HSYNC" && m.ThroughSeq >= syncSeq
	})
	if err != nil {
		return inbound, met, fmt.Errorf("plantwire: barrier: %w", err)
	}
	return inbound, met, nil
}
