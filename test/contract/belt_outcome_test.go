package contract

import (
	"math"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

const outcomeHorizonS = 3600.0

func entityOutcome(t *testing.T, facts *model.OutcomeFacts, id string) model.EntityOutcome {
	t.Helper()
	for _, entity := range facts.Entities {
		if entity.EntityID == id {
			return entity
		}
	}
	t.Fatalf("no outcome for %s", id)
	return model.EntityOutcome{}
}

func recordedTrip(entity model.EntityOutcome, state string) model.TripOutcome {
	for _, trip := range entity.Trips {
		if trip.State == state {
			return trip
		}
	}
	return model.TripOutcome{}
}

func TestTheOutcomeRecorderTimesTheSurvivorsInverterTripsAndTheStrandedBelt(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("a plant-B hour of world time")
	}
	regime := beltRegimes[1]
	w, err := world.New(beltPlant(t, regime.name), 5, "belt", model.DefaultStartTimeNS, world.Options{Noiseless: true, EmitDisabled: true, RecordOutcome: true})
	if err != nil {
		t.Fatal(err)
	}
	b := &beltWorld{t: t, w: w}
	load, theta0, seen := survivorsAfterATrip(t, b, regime, "theta_inv")
	want := seen - faultAtS + tripAfter(inverterTauS, theta0, currentSquared(load))
	b.advance(outcomeHorizonS)
	facts := w.OutcomeFacts(beltNS(outcomeHorizonS))
	for _, survivor := range []string{"D-1", "D-2"} {
		trip := recordedTrip(entityOutcome(t, facts, survivor), "inverter_ok")
		got := float64(trip.FirstNS-model.DefaultStartTimeNS)/1e9 - faultAtS
		if trip.Count != 1 || math.Abs(got-want) > tripStepsS {
			t.Errorf("%s: recorded inverter trip %+v, %.1f s after D-3; closed form %.1f s", survivor, trip, got, want)
		}
	}
	belt := entityOutcome(t, facts, "B-1")
	if len(belt.Boundaries) != 1 || belt.Boundaries[0].Name != "stranded" || belt.Boundaries[0].Violations < 1 {
		t.Fatalf("belt boundaries %+v, want the stranded belt", belt.Boundaries)
	}
	stranded := belt.Boundaries[0]
	strandedS := float64(stranded.FirstNS-model.DefaultStartTimeNS) / 1e9
	left := outcomeHorizonS - strandedS
	t.Logf("plant b: inverter trips %.1f s after D-3 (closed form %.1f s); stranded %d time(s) from %.0f s for %.0f s of the %.0f s left",
		float64(recordedTrip(entityOutcome(t, facts, "D-1"), "inverter_ok").FirstNS-model.DefaultStartTimeNS)/1e9-faultAtS, want,
		stranded.Violations, strandedS, stranded.DurationS, left)
	if stranded.DurationS <= 0 || stranded.DurationS > left || (stranded.Violations == 1 && stranded.DurationS != left) {
		t.Errorf("stranded %+v: a duration of %.0f s left before the horizon, want it all for one violation that lasts", stranded, left)
	}
}
