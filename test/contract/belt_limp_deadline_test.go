package contract

import (
	"slices"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

const (
	limpWatchS    = 1800.0
	limpFeedPct   = 18.0
	latestLimpS   = 98.0
	latestCutS    = 103.0
	deadlineHighS = 300.0
)

// limpSaves trips D-3 in the steady weak plant at full feed (S2-B), sets the feed once after the delay and reports whether,
// over the half hour after that command, no survivor protection trips and the belt is never stranded.
func limpSaves(t *testing.T, delayS, percent float64) bool {
	t.Helper()
	b := newBeltWorld(t, beltPlant(t, "b"))
	b.advance(faultAtS)
	b.inject("D-3", "conveyor_drive__vfd_fault", 1)
	b.advance(faultAtS + delayS)
	b.invoke("bulk_feeder__set_feed", "F-1", map[string]any{"percent": percent})
	b.advance(faultAtS + delayS + limpWatchS)
	return quietAfterTheTrip(worldOutcome(b.w))
}

func quietAfterTheTrip(facts *model.OutcomeFacts) bool {
	for _, entity := range facts.Entities {
		for _, trip := range entity.Trips {
			if slices.Contains(survivors, entity.EntityID) && slices.Contains(storyProtections, trip.State) {
				return false
			}
		}
		for _, boundary := range entity.Boundaries {
			if entity.EntityID == "B-1" && boundary.Name == "stranded" {
				return false
			}
		}
	}
	return true
}

// latestOnTheSecondGrid is the largest whole second in [lo, hi) at which saves holds, for a property that holds at lo,
// fails at hi and switches once in between.
func latestOnTheSecondGrid(lo, hi float64, saves func(float64) bool) float64 {
	for hi-lo > 1 {
		mid := lo + float64(int((hi-lo)/2))
		if saves(mid) {
			lo = mid
		} else {
			hi = mid
		}
	}
	return lo
}

func TestHowLateAfterD3TripsASingleFeedCommandStillSavesPlantB(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("bisects the reaction deadline of plant b on a one-second grid")
	}
	testsupport.SkipUnderRace(t, "about thirty plant-b runs of half an hour")
	for _, tc := range []struct {
		name            string
		percent, pinned float64
	}{{"limp feed 18 %", limpFeedPct, latestLimpS}, {"cut to 0", 0, latestCutS}} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			saves := func(delayS float64) bool { return limpSaves(t, delayS, tc.percent) }
			if !saves(0) || saves(deadlineHighS) {
				t.Fatalf("set_feed %.0f %%: saves at once %v, %.0f s after the trip %v; want saved at once and lost by then", tc.percent, saves(0), deadlineHighS, saves(deadlineHighS))
			}
			latest := latestOnTheSecondGrid(0, deadlineHighS, saves)
			t.Logf("S2-B: one set_feed %.0f %% saves the survivors and the belt up to %.0f s after D-3 trips", tc.percent, latest)
			if !saves(latest-1) || saves(latest+1) {
				t.Errorf("set_feed %.0f %%: saved %v at %.0f s and %v at %.0f s, want the story to switch once at %.0f s",
					tc.percent, saves(latest-1), latest-1, saves(latest+1), latest+1, latest)
			}
			if latest != tc.pinned {
				t.Errorf("set_feed %.0f %%: the deadline is %.0f s, M2-DESIGN.md and PLAN.md section 1.1 record %.0f s", tc.percent, latest, tc.pinned)
			}
		})
	}
}
