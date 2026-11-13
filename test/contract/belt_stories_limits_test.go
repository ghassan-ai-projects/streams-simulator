package contract

import (
	"math"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

const (
	nominalFeedKgS  = 41.6667
	limpMarginLoPct = 2.0
	limpMarginHiPct = 6.0
)

// limpLimits are the highest feeds, in percent, at which the two drives left after D-3 is stopped can still restart the
// belt they leave loaded, and stay at or under their rating while running.
func limpLimits(regime beltRegime) (restartPct, ratingPct float64) {
	drag := 0.12 * regime.rated
	restart := (2*regime.boost*regime.rated - breakawayN - drag) / (regime.breakaway * steadyMassKg)
	rating := (2*regime.rated - beltEmptyN - drag) / (beltPerKgN*steadyMassKg + beltSpeed*nominalFeedKgS)
	return 100 * restart, 100 * rating
}

func TestTheSafeLimpFeedsSitAFewPointsBelowWhatTwoDrivesCanRestartAndCarry(t *testing.T) {
	t.Parallel()
	table := loadStories(t)
	for _, regime := range beltRegimes {
		restart, rating := limpLimits(regime)
		limit, chosen := math.Min(restart, rating), table.SafeLimpFeedPct[regime.name]
		t.Logf("plant %s: two drives restart the belt up to %.1f %% feed and stay within rating up to %.1f %%; safe limp feed %.0f %%", regime.name, restart, rating, chosen)
		if chosen > limit-limpMarginLoPct || chosen < limit-limpMarginHiPct {
			t.Errorf("plant %s: safe limp feed %.0f %%, want %.0f-%.0f points below the limit %.1f %%", regime.name, chosen, limpMarginLoPct, limpMarginHiPct, limit)
		}
	}
}

// survivorsTrip trips D-3 in a steady plant, sets the feed after a delay and reports whether a survivor's protection trips.
func survivorsTrip(t *testing.T, regime beltRegime, delayS, percent float64) float64 {
	t.Helper()
	b := newBeltWorld(t, beltPlant(t, regime.name))
	b.advance(faultAtS)
	b.inject("D-3", "conveyor_drive__vfd_fault", 1)
	b.advance(faultAtS + delayS)
	b.invoke("bulk_feeder__set_feed", "F-1", map[string]any{"percent": percent})
	return b.firstTime("D-1", "inverter_ok", 0, faultAtS+1200) - faultAtS
}

// surgeTrips sends one feed surge of the given severity into the steady plant and reports whether a drive trips.
func surgeTrips(t *testing.T, regime beltRegime, severity float64) bool {
	t.Helper()
	b := newBeltWorld(t, beltPlant(t, regime.name))
	b.advance(faultAtS)
	b.inject("F-1", "bulk_feeder__feed_surge", severity)
	return !math.IsInf(b.firstTime("D-1", "inverter_ok", 0, faultAtS+1500), 1)
}

func bisect(lo, hi, step float64, holds func(float64) bool) float64 {
	for hi-lo > step {
		mid := (lo + hi) / 2
		if holds(mid) {
			lo = mid
		} else {
			hi = mid
		}
	}
	return lo
}

func TestHowLongPlantBHasToReactAndHowLargeASurgeItRidesOut(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("searches reaction times and surge sizes in plant b")
	}
	testsupport.SkipUnderRace(t, "about thirty plant-b runs of twenty minutes")
	table, regime := loadStories(t), beltRegimes[1]
	if trip := survivorsTrip(t, regime, 10, 0); !math.IsInf(trip, 1) {
		t.Errorf("S2-B reference: feed 0 ten seconds after D-3 still trips a survivor %.0f s after D-3", trip)
	}
	latest := bisect(10, 600, 1, func(delay float64) bool { return math.IsInf(survivorsTrip(t, regime, delay, 0), 1) })
	b4 := survivorsTrip(t, regime, 1, 67)
	t.Logf("S2-B: feed 0 saves the survivors up to %.0f s after D-3; feed 67 %% one second after D-3 (B4) trips them %.0f s after D-3", latest, b4)
	bound := bisect(0, 2, 0.01, func(severity float64) bool { return !surgeTrips(t, regime, severity) })
	t.Logf("S4-B: plant b rides out one 300 s surge up to %.2fx the feed", 1+bound)
	for _, story := range table.Stories {
		if story.ID != "s4-b" {
			continue
		}
		for _, row := range story.Rows {
			for _, severity := range rowNumbers(row, "pulse_severity") {
				if severity > bound {
					t.Errorf("s4-b seed %.0f: a surge of severity %.2f is above what plant b rides out (%.2f)", row["seed"], severity, bound)
				}
			}
		}
	}
}
