// Package runner is the director's wall-time and goroutine edge: the pacer
// that makes simulated time follow wall time at a requested rate, and the
// goroutine a job runs on, so the use cases stay free of clock reads and
// go statements.
package runner

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Pacer struct {
	mu          sync.Mutex
	rate        float64
	anchorWall  time.Time
	anchorSimNS int64
	changed     chan struct{}
}

func NewPacer(rate float64) *Pacer {
	return &Pacer{rate: rate, anchorWall: time.Now(), changed: make(chan struct{})}
}

func (p *Pacer) Reanchor(simElapsedNS int64, rate float64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.rate, p.anchorSimNS, p.anchorWall = rate, simElapsedNS, time.Now()
	close(p.changed)
	p.changed = make(chan struct{})
}

func (p *Pacer) WaitUntil(ctx context.Context, simElapsedNS int64) error {
	for {
		wait, changed := p.plan(simElapsedNS)
		if wait <= 0 {
			return interrupted(ctx)
		}
		done, err := waitFor(ctx, wait, changed)
		if done {
			return err
		}
	}
}

func waitFor(ctx context.Context, wait time.Duration, changed <-chan struct{}) (bool, error) {
	timer := time.NewTimer(wait)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return true, interrupted(ctx)
	case <-changed:
		return false, nil
	case <-timer.C:
		return true, nil
	}
}

func interrupted(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("runner: pacing interrupted: %w", err)
	}
	return nil
}

func (p *Pacer) plan(simElapsedNS int64) (time.Duration, <-chan struct{}) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.wallWait(simElapsedNS), p.changed
}

func (p *Pacer) wallWait(simElapsedNS int64) time.Duration {
	if p.rate <= 0 {
		return 0
	}
	ahead := float64(simElapsedNS-p.anchorSimNS) / p.rate
	return time.Until(p.anchorWall.Add(time.Duration(ahead)))
}

func Go(work func()) {
	go work()
}
