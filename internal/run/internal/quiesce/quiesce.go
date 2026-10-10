// Package quiesce is the timer edge of the run module: the deadline source
// for await_consumer waits. The real implementation is a wall-clock timer;
// tests inject a fake they can fire, so deterministic tests never sleep.
package quiesce

import "time"

// DefaultTimeout bounds an await_consumer wait before the run is marked
// incomplete. The duration is fixed; the clock is injectable.
const DefaultTimeout = 30 * time.Second

// Clock supplies the deadline for await_consumer waits.
type Clock interface {
	NewTimer(time.Duration) Timer
}

// Timer is one deadline from a Clock.
type Timer interface {
	C() <-chan time.Time
	Stop() bool
}

// RealClock is the wall-clock Clock.
type RealClock struct{}

type realTimer struct{ t *time.Timer }

func (rt realTimer) C() <-chan time.Time { return rt.t.C }

func (rt realTimer) Stop() bool { return rt.t.Stop() }

// NewTimer starts a wall-clock timer of duration d.
func (RealClock) NewTimer(d time.Duration) Timer {
	return realTimer{t: time.NewTimer(d)}
}
