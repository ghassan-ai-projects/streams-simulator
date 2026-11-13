package contract

import (
	"math"
	"testing"
)

func weightedLoads(b *beltWorld, rated float64, drives ...string) (loads map[string]float64, total float64) {
	loads = map[string]float64{}
	for _, drive := range drives {
		loads[drive] = b.state(drive, "load")
		total += loads[drive] * rated
	}
	return loads, total
}

func TestASkewedDriveCarriesItsWeightedShareAndTheDrivesStillCarryTheWholeDemand(t *testing.T) {
	t.Parallel()
	regime := beltRegimes[0]
	b := newBeltWorld(t, beltPlant(t, regime.name))
	b.inject("D-3", "conveyor_drive__share_skew", 0.2)
	b.advance(60)
	if got := b.state("D-3", "weight"); got != 1.2 || b.state("B-1", "weights") != 3.2 {
		t.Fatalf("D-3 weight %v, belt weights %v, want 1.2 and 3.2", got, b.state("B-1", "weights"))
	}
	demand := beltDemand(b, regime.rated, 1, 1, 1)
	loads, total := weightedLoads(b, regime.rated, "D-1", "D-2", "D-3")
	for drive, weight := range map[string]float64{"D-1": 1, "D-2": 1, "D-3": 1.2} {
		if want := weight * demand / 3.2 / regime.rated; math.Abs(loads[drive]-want) > 1e-6*want {
			t.Errorf("%s: load %.6f, w·demand/Σw/F_rated = %.6f", drive, loads[drive], want)
		}
	}
	if math.Abs(total-demand) > 1e-6*demand {
		t.Errorf("the drives carry %.4f N, the belt demands %.4f N", total, demand)
	}
}

func TestAStoppedSkewedDriveLeavesTheSurvivorsAnEqualShare(t *testing.T) {
	t.Parallel()
	regime := beltRegimes[0]
	b := newBeltWorld(t, beltPlant(t, regime.name))
	b.inject("D-3", "conveyor_drive__share_skew", 0.2)
	b.invoke("bulk_feeder__set_feed", "F-1", map[string]any{"percent": 0})
	b.advance(660)
	if !b.invoke("conveyor_drive__stop_drive", "D-3", nil) {
		t.Fatal("stopping D-3 on an empty belt was refused")
	}
	b.advance(700)
	demand := beltDemand(b, regime.rated, 1, 1, 0)
	loads, total := weightedLoads(b, regime.rated, "D-1", "D-2", "D-3")
	if loads["D-3"] != 0 || loads["D-1"] != loads["D-2"] || math.Abs(total-demand) > 1e-6*demand {
		t.Fatalf("loads %v carry %.4f N of %.4f N, want D-3 at 0 and the survivors equal with the whole demand", loads, total, demand)
	}
}
