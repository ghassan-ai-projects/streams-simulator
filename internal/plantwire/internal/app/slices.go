package app

import (
	"context"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
)

const missedBarriersForHostLoss = 3

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
	return s.barrier(ctx, s.out.LastSeq())
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
	s.run.JournalPacing(syncSeq, s.link.WaitedMS(), !met)
	if processed := s.process(inbound); processed != nil || s.finished {
		return processed
	}
	if err != nil {
		s.queued = nil
		return fmt.Errorf("%w: %w", ErrHostLost, err)
	}
	return s.afterBarrier(met)
}

func (s *Session) afterBarrier(met bool) error {
	if !met {
		return s.countBarrier(false)
	}
	if err := s.runQueue(); err != nil {
		return err
	}
	return s.countBarrier(true)
}

func (s *Session) awaitSync(ctx context.Context, syncSeq uint64) ([]domain.Inbound, bool, error) {
	inbound, met, err := s.link.Await(ctx, func(record map[string]any) bool {
		return record["type"] == "HSYNC" && domain.Seq(record["through_seq"]) >= syncSeq
	})
	if err != nil {
		return inbound, met, fmt.Errorf("plantwire: barrier: %w", err)
	}
	return inbound, met, nil
}

func (s *Session) countBarrier(met bool) error {
	if met {
		s.missed = 0
		return nil
	}
	s.missed++
	if s.missed >= missedBarriersForHostLoss {
		s.queued = nil
		return ErrHostLost
	}
	return nil
}
