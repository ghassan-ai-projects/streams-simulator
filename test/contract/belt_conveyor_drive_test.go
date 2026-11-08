package contract

import (
	"math"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

const (
	noLoadCurrent = 0.35
	relayTauS     = 1000.0
	inverterTauS  = 700.0
	windingTauS   = 900.0
	faultAtS      = 600.0
	// a trip is sampled at the start of a 1 s step and the overload input is held over the step: two steps of slack
	tripStepsS = 2.0
)

func currentSquared(load float64) float64 {
	return noLoadCurrent*noLoadCurrent + (1-noLoadCurrent*noLoadCurrent)*load*load
}

// tripAfter is the time a lagged I² state starting at theta0 takes to cross the trip level under a constant I².
func tripAfter(tau, theta0, i2 float64) float64 {
	if i2 <= tripTheta {
		return math.Inf(1)
	}
	return tau * math.Log((i2-theta0)/(i2-tripTheta))
}

// survivorsAfterATrip runs a plant to its steady state, trips D-3's inverter and returns the survivors' load, θ0 and the
// time the survivors first see the new share (the belt reads its drives one sample late).
func survivorsAfterATrip(t *testing.T, b *beltWorld, regime beltRegime, theta string) (load, theta0, seenS float64) {
	t.Helper()
	b.advance(faultAtS)
	theta0, share := b.state("D-1", theta), b.state("D-1", "share")
	b.inject("D-3", "conveyor_drive__vfd_fault", 1)
	for b.state("D-1", "share") == share {
		b.advance(b.nowS + 1)
	}
	return beltDemand(b, regime.rated, 1, 1, 0) / 2 / regime.rated, theta0, b.nowS
}

func TestWithoutMarginTheSurvivorsInvertersTripOnTheOverloadClock(t *testing.T) {
	t.Parallel()
	regime := beltRegimes[1]
	b := newBeltWorld(t, beltPlant(t, regime.name))
	load, theta0, seen := survivorsAfterATrip(t, b, regime, "theta_inv")
	want := seen - faultAtS + tripAfter(inverterTauS, theta0, currentSquared(load))
	got := b.firstTime("D-1", "inverter_ok", 0, faultAtS+600) - faultAtS
	t.Logf("plant b: survivors at load %.4f (I² %.4f, θ0 %.4f) from %.0f s: inverter trip %.1f s after D-3, closed form %.1f s", load, currentSquared(load), theta0, seen-faultAtS, got, want)
	if math.Abs(got-want) > tripStepsS {
		t.Errorf("the survivors' inverters trip %.1f s after D-3, closed form %.1f s", got, want)
	}
	if b.state("D-2", "inverter_ok") != 0 || b.state("D-1", "relay_ok") != 1 {
		t.Errorf("both survivors' inverters, and no relay, should have tripped")
	}
}

func TestWithASlowInverterTheMotorRelayTripsOnItsOwnClock(t *testing.T) {
	t.Parallel()
	regime := beltRegimes[1]
	spec := beltVariant(t, regime.name, func(doc map[string]any) { driveParameters(doc)["inverter_tau_s"] = 3600 })
	b := newBeltWorld(t, spec)
	load, theta0, seen := survivorsAfterATrip(t, b, regime, "theta_r")
	want := seen - faultAtS + tripAfter(relayTauS, theta0, currentSquared(load))
	got := b.firstTime("D-1", "relay_ok", 0, faultAtS+600) - faultAtS
	t.Logf("plant b, inverter τ 3600 s: relay trip %.1f s after D-3, closed form %.1f s", got, want)
	if math.Abs(got-want) > tripStepsS || b.state("D-1", "inverter_ok") != 1 {
		t.Errorf("the relay trips %.1f s after D-3 (inverter_ok %v), closed form %.1f s", got, b.state("D-1", "inverter_ok"), want)
	}
}

func TestWithMarginTheSurvivorsCarryTheBeltAndTheirWindingsFollowAFirstOrderStep(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("runs the survivors for three relay time constants")
	}
	testsupport.SkipUnderRace(t, "fifty noiseless minutes of the belt")
	regime := beltRegimes[0]
	b := newBeltWorld(t, beltPlant(t, regime.name))
	load, _, seen := survivorsAfterATrip(t, b, regime, "theta_r")
	i2 := currentSquared(load)
	if i2 >= tripTheta {
		t.Fatalf("plant a: survivors at load %.4f reach I² %.4f, above the trip level %v", load, i2, tripTheta)
	}
	t0, settled := b.state("D-1", "winding"), ambientC+24+56*load*load
	for _, after := range []float64{300, 900, 1800, 3000} {
		b.advance(seen + after)
		want := settled - (settled-t0)*math.Exp(-after/windingTauS)
		if got := b.state("D-1", "winding"); math.Abs(got-want) > 0.1 {
			t.Errorf("%.0f s after the trip: winding %.3f °C, closed form %.3f °C", after, got, want)
		}
		for _, protection := range []string{"relay_ok", "inverter_ok", "overtemp_ok"} {
			if b.state("D-1", protection) != 1 || b.state("D-2", protection) != 1 {
				t.Fatalf("plant a: a survivor's %s tripped %.0f s after D-3", protection, after)
			}
		}
	}
	t.Logf("plant a: survivors at load %.4f, I² %.4f settles below %v; winding %.2f -> %.2f °C", load, i2, tripTheta, t0, settled)
}

func TestADryBearingWearsWithTheCubeOfItsLoad(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("runs a dry bearing for an hour")
	}
	testsupport.SkipUnderRace(t, "an hour of the belt")
	const severity = 0.8
	regime := beltRegimes[0]
	b := newBeltWorld(t, beltPlant(t, regime.name))
	b.advance(60)
	load := beltDemand(b, regime.rated, 1, 1, 1) / 3 / regime.rated
	b.inject("D-3", "conveyor_drive__lube_loss", severity)
	for _, hours := range []float64{0.25, 0.5, 1} {
		b.advance(60 + hours*3600)
		want := regime.wear * load * load * load * severity * hours * hours / 2
		got := b.state("D-3", "bearing_damage")
		t.Logf("%.2f h: damage %.5f, closed form %.5f", hours, got, want)
		if math.Abs(got-want) > 0.01*want {
			t.Errorf("%.2f h after the grease is gone: damage %.5f, closed form %.5f", hours, got, want)
		}
	}
}
