// Package quiesce is the timer edge of the run module: the deadline source
// for await_consumer waits. The real implementation is a wall-clock timer;
// tests inject a fake they can fire, so deterministic tests never sleep.
package quiesce

import "time"

const DefaultTimeout = 30 * time.Second

type Clock interface {
	NewTimer(time.Duration) Timer
}

type Timer interface {
	C() <-chan time.Time
	Stop() bool
}

type RealClock struct{}

type realTimer struct{ t *time.Timer }

func (rt realTimer) C() <-chan time.Time { return rt.t.C }

func (rt realTimer) Stop() bool { return rt.t.Stop() }

func (RealClock) NewTimer(d time.Duration) Timer {
	return realTimer{t: time.NewTimer(d)}
}
