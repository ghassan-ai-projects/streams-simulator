package contract

import (
	"math"
	"slices"
	"strings"
	"testing"
)

const (
	feedLagS = 3.0
	stepS    = 1.0
)

// materialOnBelt solves dm/dt = u(t) − m/T with T = L/v and a feed that relaxes from u0 to u1 with time constant τ.
func materialOnBelt(m0, u0, u1, v, t float64) float64 {
	transit := beltLength / v
	lagged := (u0 - u1) / (1/transit - 1/feedLagS)
	return u1*transit + lagged*math.Exp(-t/feedLagS) + (m0-u1*transit-lagged)*math.Exp(-t/transit)
}

// The world holds a step's inputs at the step start, so the belt sees the feed up to one step late: the tolerance is half
// a step of the feed change.
func TestMaterialOnTheBeltFollowsAFeedStepAtConstantSpeed(t *testing.T) {
	t.Parallel()
	b := newBeltWorld(t, beltPlant(t, "a"))
	b.advance(600)
	m0, u0, v := b.state("B-1", "m"), b.state("F-1", "feed"), b.state("B-1", "v")
	b.inject("F-1", "bulk_feeder__shift_pattern", 1)
	u1, onset := 0.7*u0, b.nowS
	worst := 0.0
	for _, after := range []float64{5, 15, 30, 60, 120, 300} {
		b.advance(onset + after)
		want, got := materialOnBelt(m0, u0, u1, v, after), b.state("B-1", "m")
		worst = math.Max(worst, math.Abs(got-want))
		if math.Abs(got-want) > 0.5*(u0-u1)*stepS {
			t.Errorf("%.0f s after the step: m = %.2f kg, closed form %.2f kg", after, got, want)
		}
		if speed := b.state("B-1", "v"); speed != v {
			t.Fatalf("the belt speed moved from %v to %v during a feed step", v, speed)
		}
	}
	t.Logf("m from %.1f to %.1f kg, worst deviation from the closed form %.3f kg", m0, u1*beltLength/v, worst)
}

func TestEachRunningDriveDeliversTheBeltDemandShared(t *testing.T) {
	t.Parallel()
	regime := beltRegimes[0]
	b := newBeltWorld(t, beltPlant(t, regime.name))
	b.advance(60)
	want := beltDemand(b, regime.rated, 1, 1, 1) / 3
	if got := b.state("D-1", "share"); math.Abs(got-want) > 1e-6*want {
		t.Errorf("three drives: share %.4f N, demand / 3 = %.4f N", got, want)
	}
	b.invoke("bulk_feeder__set_feed", "F-1", map[string]any{"percent": 0})
	b.advance(660)
	if !b.invoke("conveyor_drive__stop_drive", "D-3", nil) {
		t.Fatal("stopping D-3 on an empty belt was refused")
	}
	b.advance(700)
	want = beltDemand(b, regime.rated, 1, 1, 0) / 2
	if got := b.state("D-1", "share"); math.Abs(got-want) > 1e-6*want || b.state("B-1", "drives") != 2 {
		t.Errorf("two drives: share %.4f N, demand / 2 = %.4f N (drives %v)", got, want, b.state("B-1", "drives"))
	}
	if got, load := b.state("D-1", "load"), want/regime.rated; math.Abs(got-load) > 1e-5 {
		t.Errorf("load %.5f, share / F_rated = %.5f", got, load)
	}
}

func TestATwoDriveBeltSharesHalfEachAndASurvivorCarriesAll(t *testing.T) {
	t.Parallel()
	regime := beltRegimes[0]
	spec := beltVariant(t, regime.name, func(doc map[string]any) {
		doc["equipment"].([]any)[2].(map[string]any)["names"] = []string{"D-1", "D-2"}
		connections := doc["connections"].([]any)
		doc["connections"] = slices.DeleteFunc(connections, func(raw any) bool {
			link := raw.(map[string]any)
			return strings.HasPrefix(link["from"].(string), "D-3") || strings.HasPrefix(link["to"].(string), "D-3")
		})
	})
	b := newBeltWorld(t, spec)
	b.invoke("bulk_feeder__set_feed", "F-1", map[string]any{"percent": 30})
	b.advance(300)
	half := beltDemand(b, regime.rated, 1, 1) / 2
	if driveRunCap*regime.rated < 2*half {
		t.Fatalf("the test needs a feed one drive can carry: demand %.1f N", 2*half)
	}
	for _, drive := range []string{"D-1", "D-2"} {
		if got := b.state(drive, "share"); math.Abs(got-half) > 1e-6*half {
			t.Errorf("%s: share %.4f N, demand / 2 = %.4f N", drive, got, half)
		}
	}
	b.inject("D-2", "conveyor_drive__vfd_fault", 1)
	b.advance(310)
	all := beltDemand(b, regime.rated, 1, 0)
	if got := b.state("D-1", "share"); math.Abs(got-all) > 1e-6*all {
		t.Errorf("one survivor: share %.4f N, the whole demand %.4f N", got, all)
	}
	t.Logf("2-drive belt: %.2f N each, %.2f N on the survivor (%.3f times)", half, all, all/half)
}
