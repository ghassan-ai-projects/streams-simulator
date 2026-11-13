package contract

import (
	"math"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

const (
	referenceSampleS = 1.0
	promoteS         = 2.0
	stopTauS         = 2.0
	stoppedSpeed     = 0.1
	strandHoldS      = 10.0
	lineDrives       = 3
)

// masterLost runs an S6 plant to its steady state and trips D-1's inverter at faultAtS.
func masterLost(t *testing.T, b *beltWorld) *beltWorld {
	t.Helper()
	b.advance(faultAtS)
	b.inject("D-1", "drive_master__vfd_fault", 1)
	return b
}

// stopAllClock is the line's stop after the master is lost, in seconds after it, in closed form. The k-th drive of the
// chain (the master is drive 0) stops k reference samples after the master; the belt counts its running drives one sample
// late; the shortfall begins when the drives the belt counts cannot carry the running demand at their run limit; the
// stall trip latches boost_s later; the belt coasts down on τ 2 s from the shortfall to the stopped speed; the stranded
// boundary holds 10 s after that.
func stopAllClock(regime beltRegime) (shortS, stallS, stoppedS, strandedS float64) {
	for k := 1; k <= lineDrives; k++ {
		stopped := float64(k)
		demand := beltEmptyN + beltPerKgN*steadyMassKg + beltSpeed*nominalFeedKgS + 0.03*regime.rated*(lineDrives+stopped)
		if (lineDrives-stopped)*driveRunCap*regime.rated < demand {
			shortS = float64(k) * referenceSampleS
			break
		}
	}
	stoppedS = shortS + stopTauS*math.Log(beltSpeed/stoppedSpeed)
	return shortS, shortS + boostS, stoppedS, stoppedS + strandHoldS
}

// firstTimes steps one second at a time for the given time and records when each watched state first takes its value.
func (b *beltWorld) firstTimes(forS float64, watch map[string]float64) map[string]float64 {
	marks := map[string]float64{}
	for end := b.nowS + forS; b.nowS < end; {
		b.advance(b.nowS + 1)
		for key, value := range watch {
			entity, state := splitKey(key)
			if _, seen := marks[key]; !seen && b.state(entity, state) == value {
				marks[key] = b.nowS - faultAtS
			}
		}
	}
	return marks
}

func TestLosingTheMasterStopsTheFollowersAndStrandsTheLoadedBelt(t *testing.T) {
	t.Parallel()
	regime := beltRegimes[0]
	b := masterLost(t, newBeltWorld(t, beltPlant(t, "s6-stop")))
	marks := b.firstTimes(60, map[string]float64{"D-2.enabled": 0, "D-3.enabled": 0, "B-1.stall_ok": 0, "B-1.stopped": 1, "B-1.stranded": 1})
	short, stall, stopped, stranded := stopAllClock(regime)
	t.Logf("stop_all: D-2 off %.0f s, D-3 off %.0f s; shortfall from %.0f s; stall %.0f s (closed form %.0f s), stopped %.0f s (%.2f s), stranded %.0f s (%.2f s)",
		marks["D-2.enabled"], marks["D-3.enabled"], short, marks["B-1.stall_ok"], stall, marks["B-1.stopped"], stopped, marks["B-1.stranded"], stranded)
	if marks["D-2.enabled"] != referenceSampleS || marks["D-3.enabled"] != 2*referenceSampleS {
		t.Errorf("the followers stop %.0f s and %.0f s after the master, want one reference sample each down the line", marks["D-2.enabled"], marks["D-3.enabled"])
	}
	for key, want := range map[string]float64{"B-1.stall_ok": stall, "B-1.stopped": math.Ceil(stopped), "B-1.stranded": math.Ceil(stranded)} {
		if got, seen := marks[key]; !seen || got != want {
			t.Errorf("%s first at %.0f s after the master, closed form %.0f s", key, got, want)
		}
	}
	margin := 2*regime.boost*regime.rated - (breakawayN + regime.breakaway*b.state("B-1", "m")) - 3*2*0.03*regime.rated
	if got := b.state("B-1", "restart_margin"); math.Abs(got-margin) > 1e-6 || got >= 0 {
		t.Errorf("restart margin %.3f N, closed form %.3f N: two followers cannot break the loaded belt loose", got, margin)
	}
	if b.state("D-2", "promoted") != 0 || b.state("D-2", "speed_lead") != 0 {
		t.Errorf("stop_all: D-2 took the speed loop")
	}
}

func TestARestartWithoutTheMasterStallsAgainAndCountsASecondStall(t *testing.T) {
	t.Parallel()
	b := masterLost(t, newBeltWorld(t, beltPlant(t, "s6-stop")))
	b.advance(faultAtS + 119)
	stopped := b.invoke("belt_conveyor__stop_belt", "B-1", nil)
	b.advance(faultAtS + 120)
	if !stopped || !b.invoke("belt_conveyor__start_belt", "B-1", nil) {
		t.Fatal("the restart (stop_belt, then start_belt) was refused")
	}
	stall := b.firstTime("B-1", "stall_ok", 0, b.nowS+30) - (faultAtS + 120)
	b.advance(faultAtS + 180)
	facts := worldOutcome(b.w)
	t.Logf("a restart 120 s after the master is lost stalls %.0f s later; v %.3f m/s", stall, b.state("B-1", "v"))
	if stall > boostS+1 || b.state("B-1", "v") > stoppedSpeed || tripCount(facts, "B-1", "stall_ok") != 2 {
		t.Errorf("restart: stall after %.0f s, v %.3f, stalls counted %d; want the futile start to stall within %v s and count a second stall",
			stall, b.state("B-1", "v"), tripCount(facts, "B-1", "stall_ok"), boostS+1)
	}
}

func TestPromotionHandsTheSpeedLoopToD2AndTheSurvivorsShareTheBeltAsAfterAnyTrip(t *testing.T) {
	t.Parallel()
	regime := beltRegimes[0]
	for _, hold := range []float64{promoteS, 5} {
		plant := beltPlant(t, "s6-promote")
		if hold != promoteS {
			plant = beltVariant(t, "s6-promote", func(doc map[string]any) { followerParameters(doc)["promote_s"] = hold })
		}
		b := masterLost(t, newBeltWorld(t, plant))
		marks := b.firstTimes(60, map[string]float64{"D-2.speed_lead": 1, "D-3.speed_lead": 1, "D-2.enabled": 0, "D-3.enabled": 0})
		load := beltDemand(b, regime.rated, 0, 1, 1) / 2 / regime.rated
		t.Logf("promote_s %.0f: D-2 leads %.0f s after the master is lost (closed form %.0f s); survivors' load %.6f and %.6f, closed form %.6f",
			hold, marks["D-2.speed_lead"], referenceSampleS+hold, b.state("D-2", "load"), b.state("D-3", "load"), load)
		if marks["D-2.speed_lead"] != referenceSampleS+hold || len(marks) != 1 {
			t.Errorf("promote_s %.0f: first times %v, want only D-2 leading %.0f s after the master", hold, marks, referenceSampleS+hold)
		}
		for _, drive := range []string{"D-2", "D-3"} {
			if got := b.state(drive, "load"); math.Abs(got-load) > 1e-5*load {
				t.Errorf("promote_s %.0f: %s load %.6f, closed form %.6f", hold, drive, got, load)
			}
		}
	}
}

func followerParameters(doc map[string]any) map[string]any {
	return doc["equipment"].([]any)[3].(map[string]any)["parameters"].(map[string]any)
}

func splitKey(key string) (entity, state string) {
	entity, state, _ = strings.Cut(key, ".")
	return entity, state
}

func tripCount(facts *model.OutcomeFacts, entity, state string) int {
	for _, outcome := range facts.Entities {
		for _, trip := range outcome.Trips {
			if outcome.EntityID == entity && trip.State == state {
				return trip.Count
			}
		}
	}
	return 0
}
