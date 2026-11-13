package contract

import (
	"math"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

const boostS = 2.0

// breakawayMargin is what n drives at their boost limit have left after breaking the loaded belt loose.
func breakawayMargin(b *beltWorld, regime beltRegime, n int, enabled ...float64) float64 {
	drag := 0.0
	for _, on := range enabled {
		drag += 0.03 * regime.rated * (2 - on)
	}
	return float64(n)*regime.boost*regime.rated - (breakawayN + regime.breakaway*b.state("B-1", "m")) - drag
}

func (b *beltWorld) fastest(entity, name string, untilS float64) float64 {
	b.t.Helper()
	top := b.state(entity, name)
	for b.nowS < untilS {
		b.advance(b.nowS + 1)
		top = math.Max(top, b.state(entity, name))
	}
	return top
}

func TestThreeDrivesRestartALoadedBeltAndTwoDoNot(t *testing.T) {
	t.Parallel()
	for _, regime := range beltRegimes {
		three := newBeltWorld(t, beltPlant(t, regime.name))
		three.advance(60)
		three.invoke("belt_conveyor__stop_belt", "B-1", nil)
		three.advance(120)
		if margin := breakawayMargin(three, regime, 3, 1, 1, 1); margin < 0 || !three.invoke("belt_conveyor__start_belt", "B-1", nil) {
			t.Fatalf("plant %s: three drives have %.1f N left to break the loaded belt loose, or the start was refused", regime.name, margin)
		}
		if three.advance(180); three.state("B-1", "v") < 0.95*beltSpeed || three.state("B-1", "stall_ok") != 1 {
			t.Errorf("plant %s: three drives did not restart the loaded belt (v %.3f)", regime.name, three.state("B-1", "v"))
		}

		two := newBeltWorld(t, beltPlant(t, regime.name))
		two.advance(60)
		two.invoke("belt_conveyor__stop_belt", "B-1", nil)
		two.advance(120)
		two.inject("D-3", "conveyor_drive__vfd_fault", 1)
		margin := breakawayMargin(two, regime, 2, 1, 1, 0)
		t.Logf("plant %s: breakaway margin %.1f N with three drives, %.1f N with two", regime.name, breakawayMargin(two, regime, 3, 1, 1, 1), margin)
		if margin >= 0 {
			t.Fatalf("plant %s: two drives have %.1f N to spare at breakaway; the test needs a loaded belt they cannot start", regime.name, margin)
		}
		if two.advance(130); two.state("B-1", "stranded") != 0 {
			t.Errorf("plant %s: stranded within 10 s of losing the third drive", regime.name)
		}
		if two.advance(133); two.state("B-1", "stranded") != 1 {
			t.Errorf("plant %s: a stopped loaded belt two drives cannot start is not stranded 10 s later", regime.name)
		}
		two.advance(140)
		two.invoke("belt_conveyor__start_belt", "B-1", nil)
		if stall := two.firstTime("B-1", "stall_ok", 0, 160) - 140; stall > boostS+1 {
			t.Errorf("plant %s: the stall trip acts %.0f s after the start, want within the boost time + 1 s", regime.name, stall)
		}
		if top := two.fastest("B-1", "v", 160); top >= 0.1 {
			t.Errorf("plant %s: two drives moved the loaded belt to %.3f m/s", regime.name, top)
		}
	}
}

func TestTwoDrivesStartAnEmptyBelt(t *testing.T) {
	t.Parallel()
	for _, regime := range beltRegimes {
		b := newBeltWorld(t, beltPlant(t, regime.name))
		b.invoke("bulk_feeder__set_feed", "F-1", map[string]any{"percent": 0})
		b.advance(600)
		b.invoke("belt_conveyor__stop_belt", "B-1", nil)
		b.advance(640)
		b.inject("D-3", "conveyor_drive__vfd_fault", 1)
		b.advance(660)
		if margin := breakawayMargin(b, regime, 2, 1, 1, 0); margin < 0 || !b.invoke("belt_conveyor__start_belt", "B-1", nil) {
			t.Fatalf("plant %s: two drives have %.1f N left on the empty belt, or the start was refused", regime.name, margin)
		}
		if b.advance(720); b.state("B-1", "v") < 0.95*beltSpeed || b.state("B-1", "stall_ok") != 1 {
			t.Errorf("plant %s: two drives did not start the empty belt (v %.3f)", regime.name, b.state("B-1", "v"))
		}
	}
}

// stopAllowed is the plant's stop permissive computed from the design: n − 1 drives can still restart and carry the belt.
func stopAllowed(b *beltWorld, regime beltRegime, n int, enabled ...float64) bool {
	after := float64(n-1) / float64(n)
	restart := float64(n)*regime.boost*regime.rated*after - (breakawayN + regime.breakaway*b.state("B-1", "m"))
	carry := float64(n)*regime.rated*after - beltDemand(b, regime.rated, enabled...)
	for _, on := range enabled {
		restart -= 0.03 * regime.rated * (2 - on)
	}
	return restart >= 0 && carry >= 0
}

func TestTheStopPermissiveRefusesAStopTheRemainingDrivesCouldNotRecoverFrom(t *testing.T) {
	t.Parallel()
	for _, regime := range beltRegimes {
		b := newBeltWorld(t, beltPlant(t, regime.name))
		b.advance(60)
		rows := []struct {
			name, drive string
			emptyFirst  bool
			n           int
			enabled     []float64
			advanceS    float64
		}{
			{"loaded belt, three drives", "D-3", false, 3, []float64{1, 1, 1}, 60},
			{"empty belt, three drives", "D-3", true, 3, []float64{1, 1, 1}, 660},
			{"empty belt, two drives", "D-2", false, 2, []float64{1, 1, 0}, 700},
		}
		for _, row := range rows {
			if row.emptyFirst {
				b.invoke("bulk_feeder__set_feed", "F-1", map[string]any{"percent": 0})
			}
			b.advance(row.advanceS)
			want := stopAllowed(b, regime, row.n, row.enabled...)
			if got := b.invoke("conveyor_drive__stop_drive", row.drive, nil); got != want {
				t.Errorf("plant %s, %s (m %.1f kg): stop accepted %v, the permissive says %v", regime.name, row.name, b.state("B-1", "m"), got, want)
			}
			t.Logf("plant %s, %s: stop allowed %v", regime.name, row.name, want)
		}
	}
}

func TestASeizedDriveStallsTheBeltUnlessItDisengages(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("runs a gearbox without oil for half an hour, twice")
	}
	testsupport.SkipUnderRace(t, "half an hour of the belt, twice")
	const severity, gearRatePerHour, seizeAt = 4.0, 0.5, 0.95
	regime := beltRegimes[0]
	for _, seizedDrag := range []float64{1830, 0} {
		spec := beltVariant(t, regime.name, func(doc map[string]any) { driveParameters(doc)["seized_drag_n"] = seizedDrag })
		b := newBeltWorld(t, spec)
		b.advance(60)
		b.inject("D-3", "conveyor_drive__gear_oil_loss", severity)
		want := 60 + seizeAt/(gearRatePerHour*severity)*3600
		seized := b.firstTime("D-3", "free_ok", 0, want+10)
		if math.Abs(seized-want) > 2 {
			t.Errorf("seized drag %v: the gearbox seizes at %.0f s, closed form %.0f s", seizedDrag, seized, want)
		}
		g := b.state("D-3", "g")
		demand := beltDemand(b, regime.rated, 1, 1, 0) + 0.27*regime.rated*g + seizedDrag
		short := demand - 2*driveRunCap*regime.rated
		stall := b.firstTime("B-1", "stall_ok", 0, seized+60)
		t.Logf("seized drag %v: demand %.1f N against %.1f N of two drives; stall at %.0f s after the seizure", seizedDrag, demand, 2*driveRunCap*regime.rated, stall-seized)
		switch {
		case short > 0 && stall-seized > boostS+1:
			t.Errorf("seized drag %v: the belt is %.1f N short but did not stall within the boost time + 1 s", seizedDrag, short)
		case short <= 0 && (!math.IsInf(stall, 1) || b.state("B-1", "v") < 0.95*beltSpeed):
			t.Errorf("seized drag %v: two drives can carry the belt (%.1f N to spare) but it stopped", seizedDrag, -short)
		}
	}
}
