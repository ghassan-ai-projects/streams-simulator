package quiesce

import (
	"testing"
	"time"
)

func TestRealClockTimerFiresAndStops(t *testing.T) {
	t.Parallel()
	fired := RealClock{}.NewTimer(time.Millisecond)
	select {
	case <-fired.C():
	case <-time.After(30 * time.Second):
		t.Fatal("the timer must fire")
	}
	stopped := RealClock{}.NewTimer(time.Hour)
	if !stopped.Stop() {
		t.Fatal("a pending timer stops")
	}
}
