package app

import (
	"sync"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// BeginRun reads the seal and RevealTruth asks the director whether a run is
// open; neither may hold its lock while taking the other's.
func TestBeginRunAndRevealTruthNeverDeadlock(t *testing.T) {
	t.Parallel()
	d := newTestDirector(t)
	const n = 150
	ids := make([]string, n)
	runs := make([]string, n)
	for i := range n {
		ids[i] = createWorld(t, d)
		runs[i] = d.World(ids[i]).Run.ID
		if err := d.SealTruth(runs[i], &model.GroundTruthRecord{ScenarioID: "x", Domain: "aquaculture-pond", Label: "l", EntityID: "e"}); err != nil {
			t.Fatal(err)
		}
	}
	done := make(chan struct{})
	go func() {
		var wg sync.WaitGroup
		start := make(chan struct{})
		for i := range n {
			wg.Add(2)
			go func() { defer wg.Done(); <-start; _, _ = d.BeginRun(ids[i], "l") }()
			go func() { defer wg.Done(); <-start; _, _ = d.RevealTruth(runs[(i+1)%n], true) }()
		}
		close(start)
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(30 * time.Second):
		t.Fatal("BeginRun and RevealTruth deadlocked on the director and truth locks")
	}
}
