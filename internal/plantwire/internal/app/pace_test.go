package app

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/uds"
)

type fakePacer struct {
	host        *scriptedHost
	released    []int64
	syncsBefore []int
	interrupt   int
	cancel      context.CancelFunc
	worldNow    int64
	lag         int64
	wait        int64
}

func (p *fakePacer) WaitNS(worldNS int64) int64 {
	p.released = append(p.released, worldNS)
	if p.lag > 0 {
		p.worldNow = worldNS + p.lag
	}
	p.syncsBefore = append(p.syncsBefore, countOf(p.host.types(), "NSYNC"))
	if len(p.released) == p.interrupt {
		p.cancel()
	}
	return p.wait
}

func (p *fakePacer) WorldNowNS() int64 { return p.worldNow }

func pacedSession(t *testing.T, host *scriptedHost, slices int64, interrupt int) (*fakePacer, error) {
	t.Helper()
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	pacer := &fakePacer{host: host, interrupt: interrupt, cancel: cancel, wait: 1}
	options := sliceOptions(slices)
	options.Pace = pacer
	return pacer, NewSession(utilityRun(t), host, options).Run(ctx)
}

func clockModeOf(host *scriptedHost) any {
	metrics, _ := birthsOf(host)[0]["node_metrics"].(map[string]any)
	return metrics["clock_mode"]
}

func TestAPacedNodeReleasesEverySliceBeforeItsDataAndSaysItsClockIsTheWall(t *testing.T) {
	t.Parallel()
	host := &scriptedHost{t: t}
	pacer, err := pacedSession(t, host, 3, 0)
	if err != nil {
		t.Fatal(err)
	}
	want := []int64{sliceEnd(1), sliceEnd(2), sliceEnd(3)}
	if !slices.Equal(pacer.released, want) || !slices.Equal(pacer.syncsBefore, []int{0, 1, 2}) {
		t.Fatalf("released %v with %v barriers sent before each, want %v each before its own barrier", pacer.released, pacer.syncsBefore, want)
	}
	if mode := clockModeOf(host); mode != "wall" {
		t.Errorf("a paced node announces clock_mode %v, want wall", mode)
	}
	unpaced := &scriptedHost{t: t}
	if _, err := session(t, unpaced, 1); err != nil || clockModeOf(unpaced) != "virtual" {
		t.Errorf("an unpaced node announces clock_mode %v (%v), want virtual", clockModeOf(unpaced), err)
	}
}

func TestAnInterruptDuringThePaceEndsTheSessionBeforeThatSlice(t *testing.T) {
	t.Parallel()
	host := &scriptedHost{t: t}
	pacer, err := pacedSession(t, host, 5, 2)
	if !errors.Is(err, context.Canceled) || len(pacer.released) != 2 || countOf(host.types(), "NSYNC") != 1 {
		t.Fatalf("err %v after %d releases and %d barriers, want the interrupt after one slice", err, len(pacer.released), countOf(host.types(), "NSYNC"))
	}
}

func TestPacingIsNotTheBarrierAHostThatStopsHeartbeatingIsStillLost(t *testing.T) {
	t.Parallel()
	host := &scriptedHost{t: t, afterSync: func(h *scriptedHost, syncs int) {
		h.silent = true
		if syncs == 3 {
			h.hangUp = uds.ErrHostSilent
		}
	}}
	pacer, err := pacedSession(t, host, 10, 0)
	if !errors.Is(err, ErrHostLost) || len(pacer.released) != countOf(host.types(), "NSYNC") || len(pacer.released) > 4 {
		t.Fatalf("err %v after %d releases and %d barriers, want host loss within four slices", err, len(pacer.released), countOf(host.types(), "NSYNC"))
	}
}
