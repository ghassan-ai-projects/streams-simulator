package uds

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
)

var (
	ErrSimClosed   = errors.New("plantwire: the link to the simulated node is closed")
	ErrSimOverflow = errors.New("plantwire: the simulated node's mailbox overflowed")
)

type SimLink struct {
	up     *Mailbox[map[string]any]
	down   *Mailbox[domain.Inbound]
	wait   time.Duration
	waited int64
}

func NewSimLink(wait time.Duration) *SimLink {
	return &SimLink{up: NewMailbox[map[string]any](), down: NewMailbox[domain.Inbound](), wait: wait}
}

func (l *SimLink) Send(records []map[string]any) error {
	if l.up.Put(records...) {
		return nil
	}
	if l.up.IsShut() {
		return ErrSimClosed
	}
	return ErrSimOverflow
}

func (l *SimLink) Deliver(in domain.Inbound) bool { return l.down.Put(in) }

func (l *SimLink) TakeUp() []map[string]any { return l.up.Take() }

func (l *SimLink) Wake() { l.up.Wake() }

func (l *SimLink) UpReady() <-chan struct{} { return l.up.Ready() }

func (l *SimLink) Close() {
	l.up.Shut()
	l.down.Shut()
}

func (l *SimLink) WaitedMS() int64 { return l.waited }

func (l *SimLink) Await(ctx context.Context, until func(domain.Message) bool) ([]domain.Inbound, bool, error) {
	started := time.Now()
	defer func() { l.waited = time.Since(started).Milliseconds() }()
	return l.collect(ctx, until, l.wait)
}

func (l *SimLink) AwaitHandshake(ctx context.Context, until func(domain.Message) bool) ([]domain.Inbound, bool, error) {
	return l.collect(ctx, until, max(l.wait, HandshakeWait))
}

func (l *SimLink) Drain(ctx context.Context, wait time.Duration) ([]domain.Inbound, error) {
	inbound, _, err := l.collect(ctx, neverMet, wait)
	return inbound, err
}

func (l *SimLink) collect(ctx context.Context, until func(domain.Message) bool, wait time.Duration) ([]domain.Inbound, bool, error) {
	timer := time.NewTimer(wait)
	defer timer.Stop()
	return collectUntil(ctx, l.next, until, timer.C)
}

func (l *SimLink) next(ctx context.Context, expiry <-chan time.Time) (domain.Inbound, bool, error) {
	for {
		if in, ok := l.down.Pop(); ok {
			return in, false, nil
		}
		if expired, err := l.idle(ctx, expiry); expired || err != nil {
			return domain.Inbound{}, expired, err
		}
	}
}

func (l *SimLink) idle(ctx context.Context, expiry <-chan time.Time) (bool, error) {
	if l.down.IsShut() {
		return false, ErrSimClosed
	}
	select {
	case <-ctx.Done():
		return false, fmt.Errorf("plantwire: %w", ctx.Err())
	case <-expiry:
		return true, nil
	case <-l.down.Ready():
		return false, nil
	}
}

func Spawn(run func() error, after func()) <-chan error {
	done := make(chan error, 1)
	go func() {
		done <- run()
		after()
	}()
	return done
}
