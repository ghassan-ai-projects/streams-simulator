// Package runner is the director's wall-time and goroutine edge: the pacer
// that makes simulated time follow wall time at a requested rate, and the
// goroutine a job runs on, so the use cases stay free of clock reads and
// go statements.
package runner

import (
	"context"
	"fmt"
	"time"
)

type Pacer struct {
	rate    float64
	started time.Time
}

func NewPacer(rate float64) *Pacer {
	return &Pacer{rate: rate, started: time.Now()}
}

func (p *Pacer) WaitUntil(ctx context.Context, simElapsedNS int64) error {
	wait := p.wallWait(simElapsedNS)
	if wait <= 0 {
		return interrupted(ctx)
	}
	timer := time.NewTimer(wait)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return interrupted(ctx)
	case <-timer.C:
		return nil
	}
}

func interrupted(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("runner: pacing interrupted: %w", err)
	}
	return nil
}

func (p *Pacer) wallWait(simElapsedNS int64) time.Duration {
	if p.rate <= 0 {
		return 0
	}
	return time.Until(p.started.Add(time.Duration(float64(simElapsedNS) / p.rate)))
}

func Go(work func()) {
	go work()
}
