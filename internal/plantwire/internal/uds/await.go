package uds

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
)

func (l *Link) read() {
	scanner := bufio.NewScanner(l.conn)
	scanner.Buffer(make([]byte, 64*1024), 1<<20)
	for scanner.Scan() {
		record, err := domain.Parse(scanner.Bytes())
		if record != nil && !l.deliver(arrival{inbound: domain.Inbound{Record: record, Invalid: err}}) {
			return
		}
	}
	l.deliver(arrival{err: fmt.Errorf("plantwire: host connection: %w", firstError(scanner.Err(), io.EOF))})
}

func (l *Link) deliver(got arrival) bool {
	select {
	case l.incoming <- got:
		return true
	case <-l.done:
		return false
	}
}

func firstError(err, fallback error) error {
	if err != nil {
		return err
	}
	return fallback
}

func (l *Link) Await(ctx context.Context, until func(map[string]any) bool) ([]domain.Inbound, bool, error) {
	timer := time.NewTimer(l.wait)
	defer timer.Stop()
	var inbound []domain.Inbound
	for {
		got, expired, err := l.next(ctx, timer.C)
		if expired || err != nil {
			return inbound, false, err
		}
		if inbound = append(inbound, got); got.Invalid == nil && until(got.Record) {
			return inbound, true, nil
		}
	}
}

func (l *Link) next(ctx context.Context, expiry <-chan time.Time) (domain.Inbound, bool, error) {
	select {
	case <-ctx.Done():
		return domain.Inbound{}, false, fmt.Errorf("plantwire: %w", ctx.Err())
	case <-expiry:
		return domain.Inbound{}, true, nil
	case got := <-l.incoming:
		return got.inbound, false, got.err
	}
}
