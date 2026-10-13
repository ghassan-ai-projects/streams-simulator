package domain

import (
	"math"
	"math/rand"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

type coolingReference struct {
	temp, duty []float64
}

func newCoolingReference(seconds int) coolingReference {
	r := coolingReference{temp: []float64{0}, duty: []float64{0}}
	for k := 0; k < seconds; k++ {
		sample := max((k+9)/10-1, 0) * 10
		r.temp = append(r.temp, lagStep(r.temp[k], 10-r.duty[sample], 200, 1))
		if k%2 == 0 {
			r.duty = append(r.duty, r.duty[k], lagStep(r.duty[k], 0.5*r.temp[k], 50, 2))
		}
	}
	return r
}

func coolingLoopWorld(t *testing.T) *World {
	t.Helper()
	spec, err := domain.Load(testsupport.LoopDomain())
	if err != nil {
		t.Fatal(err)
	}
	w, err := New(spec, 9, "w-loop", model.DefaultStartTimeNS, Options{EmitDisabled: true})
	if err != nil {
		t.Fatal(err)
	}
	return w
}

func TestAFeedbackLoopClosedByASampledReadAgreesWithAnIndependentRecurrence(t *testing.T) {
	t.Parallel()
	r := newCoolingReference(4000)
	for _, s := range []int{1, 9, 10, 11, 20, 21, 600, 1234, 4000} {
		w := coolingLoopWorld(t)
		near(t, "tank temp", w.StateValue("tank-1", "temp", atSeconds(float64(s))), r.temp[s])
	}
	for _, s := range []int{2, 10, 12, 600, 1234, 4000} {
		w := coolingLoopWorld(t)
		near(t, "cooler duty", w.StateValue("cooler-1", "duty", atSeconds(float64(s))), r.duty[s])
	}
}

func TestAFeedbackLoopDoesNotDependOnTheOrderItIsRead(t *testing.T) {
	t.Parallel()
	rng := rand.New(rand.NewSource(23)) //nolint:gosec // a seeded, reproducible read schedule, not security
	w := coolingLoopWorld(t)
	at := 0.0
	for range 150 {
		at += rng.Float64() * 90
		entity, state := "tank-1", "temp"
		if rng.Intn(2) == 0 {
			entity, state = "cooler-1", "duty"
		}
		if got, want := w.StateValue(entity, state, atSeconds(at)), coolingLoopWorld(t).StateValue(entity, state, atSeconds(at)); got != want {
			t.Fatalf("%s.%s at %v s = %v after the schedule, %v read once", entity, state, at, got, want)
		}
	}
}

func TestAFeedbackLoopKeepsABoundedHistory(t *testing.T) {
	t.Parallel()
	w := coolingLoopWorld(t)
	longest := func() int {
		most := 0
		for _, id := range []string{"tank-1", "cooler-1"} {
			for _, s := range w.entities[id].states {
				most = max(most, len(s.history))
			}
		}
		return most
	}
	var afterOneDay int
	for s := 0.0; s <= 2*86400; s += 600 {
		w.StateValue("tank-1", "temp", atSeconds(s))
		if s == 86400 {
			afterOneDay = longest()
		}
	}
	if afterOneDay > 700 || longest() > afterOneDay {
		t.Fatalf("history grows over a loop: %d snapshots after a day, %d after two", afterOneDay, longest())
	}
	if steady := w.StateValue("tank-1", "temp", atSeconds(2*86400)); math.Abs(steady-10/1.5) > 1e-6 {
		t.Fatalf("the loop settles at heat/(1+0.5) = %v, got %v", 10/1.5, steady)
	}
}
