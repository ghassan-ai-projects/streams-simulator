package app

import (
	"context"
	"sync"
	"testing"
)

// The director's read tools and an advance run on different goroutines when
// two MCP calls overlap. The race detector is the oracle: a read of the
// world's clock or queues outside the run's command lock is a data race.
func TestWorldReadsDoNotOverlapAnAdvance(t *testing.T) {
	t.Parallel()
	d := newTestDirector(t)
	worldID := createWorld(t, d)
	start := d.World(worldID).Run.Status().ClockNS
	const step = int64(60e9)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := int64(1); i <= 40; i++ {
			if _, err := d.Advance(context.Background(), worldID, start+i*step, false); err != nil {
				t.Error(err)
				return
			}
		}
	}()
	go func() {
		defer wg.Done()
		for range 200 {
			_, _ = d.ClockState(worldID)
			_, _ = d.DescribeWorld(worldID)
			_, _ = d.ListFaults(worldID)
		}
	}()
	wg.Wait()
}
