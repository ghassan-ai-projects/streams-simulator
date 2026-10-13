package uds

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
)

const MaxRecordsPerWait = 4096

var ErrFlooded = errors.New("plantwire: the host sent more records than a barrier allows")

func (l *Link) read() {
	scanner := bufio.NewScanner(l.conn)
	scanner.Buffer(make([]byte, 64*1024), domain.MaxBirthBytes)
	for scanner.Scan() {
		if in, ok := domain.Decode(scanner.Bytes()); ok && !l.deliver(in) {
			return
		}
	}
	l.failed = fmt.Errorf("plantwire: host connection: %w", firstError(scanner.Err(), io.EOF))
	close(l.incoming)
}

func (l *Link) deliver(got domain.Inbound) bool {
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

func (l *Link) Await(ctx context.Context, until func(domain.Message) bool) ([]domain.Inbound, bool, error) {
	started := time.Now()
	defer func() { l.waited = time.Since(started).Milliseconds() }()
	return l.collect(ctx, until, l.wait)
}

func (l *Link) AwaitHandshake(ctx context.Context, until func(domain.Message) bool) ([]domain.Inbound, bool, error) {
	return l.collect(ctx, until, max(l.wait, HandshakeWait))
}

func (l *Link) collect(ctx context.Context, until func(domain.Message) bool, wait time.Duration) ([]domain.Inbound, bool, error) {
	timer := time.NewTimer(wait)
	defer timer.Stop()
	var inbound []domain.Inbound
	for len(inbound) < MaxRecordsPerWait {
		got, expired, err := l.next(ctx, timer.C)
		if expired || err != nil {
			return inbound, false, err
		}
		if inbound = append(inbound, got); got.Invalid == nil && until(got.Message) {
			return inbound, true, nil
		}
	}
	return inbound, false, ErrFlooded
}

func (l *Link) next(ctx context.Context, expiry <-chan time.Time) (domain.Inbound, bool, error) {
	if ctx.Err() != nil {
		return stopped(ctx)
	}
	select {
	case <-ctx.Done():
		return stopped(ctx)
	case <-expiry:
		return domain.Inbound{}, true, nil
	case got, open := <-l.incoming:
		return l.arrived(got, open)
	}
}

func (l *Link) arrived(got domain.Inbound, open bool) (domain.Inbound, bool, error) {
	if !open {
		return domain.Inbound{}, false, l.failed
	}
	return got, false, nil
}

func stopped(ctx context.Context) (domain.Inbound, bool, error) {
	return domain.Inbound{}, false, fmt.Errorf("plantwire: %w", ctx.Err())
}

func (l *Link) WaitedMS() int64 { return l.waited }
