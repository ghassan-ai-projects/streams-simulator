package uds

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
)

const MaxRecordsPerWait = 4096

var (
	ErrFlooded    = errors.New("plantwire: the host sent more records than a barrier allows")
	ErrHostSilent = errors.New("plantwire: no host heartbeat for three heartbeat intervals")
)

type beats struct {
	mu    sync.Mutex
	timer *time.Timer
}

func (b *beats) note(in domain.Inbound) {
	interval, ok := domain.HeartbeatInterval(in)
	if !ok {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if limit := domain.HeartbeatsForHostLoss * interval; b.timer == nil {
		b.timer = time.NewTimer(limit)
	} else {
		b.timer.Reset(limit)
	}
}

func (b *beats) alarm() <-chan time.Time {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.timer == nil {
		return nil
	}
	return b.timer.C
}

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
	l.beats.note(got)
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
	return collectUntil(ctx, l.next, until, timer.C)
}

type nextFunc func(ctx context.Context, expiry <-chan time.Time) (domain.Inbound, bool, error)

func collectUntil(ctx context.Context, next nextFunc, until func(domain.Message) bool, expiry <-chan time.Time) ([]domain.Inbound, bool, error) {
	var inbound []domain.Inbound
	for len(inbound) < MaxRecordsPerWait {
		got, expired, err := next(ctx, expiry)
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
	case <-l.beats.alarm():
		return domain.Inbound{}, false, ErrHostSilent
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

func (l *Link) Drain(ctx context.Context, wait time.Duration) ([]domain.Inbound, error) {
	inbound, _, err := l.collect(ctx, neverMet, wait)
	return inbound, err
}

func neverMet(domain.Message) bool { return false }

type WallPacer struct {
	schedule domain.WallSchedule
	began    time.Time
}

func StartWallPacer() *WallPacer {
	return NewWallPacer(domain.WallScheduleFrom(time.Now().UnixNano()))
}

func NewWallPacer(schedule domain.WallSchedule) *WallPacer {
	return &WallPacer{schedule: schedule, began: time.Now()}
}

func (p *WallPacer) wallNow() int64 { return p.began.UnixNano() + int64(time.Since(p.began)) }

func (p *WallPacer) StartNS() int64 { return p.schedule.WorldStartNS }

func (p *WallPacer) WorldNowNS() int64 { return p.schedule.WorldAt(p.wallNow()) }

func (p *WallPacer) WaitNS(worldNS int64) int64 {
	return p.schedule.WaitNS(worldNS, p.wallNow())
}
