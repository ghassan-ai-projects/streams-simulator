package domain

import (
	"math"
	"slices"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

var coupledUnits = []string{"unit-1", "unit-2", "unit-3"}

type headerReference struct {
	flows   map[string]*lagReference
	retired map[string]float64
}

func newHeaderReference(setpoints map[string]float64, retired map[string]float64) headerReference {
	r := headerReference{flows: map[string]*lagReference{}, retired: retired}
	for _, unit := range coupledUnits {
		r.flows[unit] = newLagReference([]lagLink{{"flow", 1, 100, "u"}}, setpoints[unit])
	}
	return r
}

func (r headerReference) present(t float64) []float64 {
	var flows []float64
	for _, unit := range coupledUnits {
		if at, gone := r.retired[unit]; !gone || t <= at {
			flows = append(flows, r.flows[unit].value("flow", t))
		}
	}
	return flows
}

func (r headerReference) sum(t float64) float64 {
	var total float64
	for _, flow := range r.present(t) {
		total += flow
	}
	return total
}

func (r headerReference) level(t float64) float64 {
	level := 100.0
	for k := 0; float64(k) < t; k++ {
		level += 0.01 * (r.sum(float64(k)) - 20)
	}
	return level
}

func (r headerReference) average(t float64) float64 {
	var avg, from float64
	for ; from+70 <= t; from += 70 {
		avg = lagStep(avg, r.mean(from), 50, 70)
	}
	if t > from {
		avg = lagStep(avg, r.mean(from), 50, t-from)
	}
	return avg
}

func (r headerReference) mean(t float64) float64 {
	return r.sum(t) / float64(len(r.present(t)))
}

func coupledWorld(t *testing.T) *World {
	t.Helper()
	spec, err := domain.Load(testsupport.CoupledDomain())
	if err != nil {
		t.Fatal(err)
	}
	w, err := New(spec, 3, "w-coupled", model.DefaultStartTimeNS, Options{EmitDisabled: true})
	if err != nil {
		t.Fatal(err)
	}
	return w
}

func atSeconds(s float64) int64 { return model.DefaultStartTimeNS + int64(s*secondsPerNS) }

func near(t *testing.T, what string, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("%s = %v, reference %v", what, got, want)
	}
}

func TestAHeaderFollowsAndReadsTheAggregatesOfItsLinkedUnits(t *testing.T) {
	t.Parallel()
	w := coupledWorld(t)
	if _, err := w.InjectFault("unit-2", "derate", atSeconds(0), nil); err != nil {
		t.Fatal(err)
	}
	if _, err := w.InjectFault("unit-3", "trip", atSeconds(300), nil); err != nil {
		t.Fatal(err)
	}
	r := newHeaderReference(map[string]float64{"unit-1": 10, "unit-2": 6, "unit-3": 10}, nil)
	for _, s := range []float64{60, 298, 302, 1000, 3600} {
		at := atSeconds(s)
		near(t, "inflow", w.StateValue("hdr-a", "inflow", at), r.sum(s))
		near(t, "peak", w.StateValue("hdr-a", "peak", at), slices.Max(r.present(s)))
		near(t, "low", w.StateValue("hdr-a", "low", at), slices.Min(r.present(s)))
		near(t, "level", w.StateValue("hdr-a", "level", at), r.level(s))
		near(t, "mean flow", w.StateValue("hdr-a", "avg_flow", at), r.average(s))
		running := 3.0
		if s >= 300 {
			running = 2
		}
		near(t, "units running", w.StateValue("hdr-a", "units_running", at), running)
	}
}

func TestAnUnlinkedHeaderKeepsItsOwnDynamicsAndTheDefaults(t *testing.T) {
	t.Parallel()
	w := coupledWorld(t)
	defaults := newLagReference([]lagLink{{"avg", 70, 50, "u"}}, 7)
	for _, s := range []float64{60, 1000} {
		at := atSeconds(s)
		near(t, "inflow", w.StateValue("hdr-b", "inflow", at), 5)
		near(t, "peak", w.StateValue("hdr-b", "peak", at), -1)
		near(t, "units running", w.StateValue("hdr-b", "units_running", at), 0)
		near(t, "level", w.StateValue("hdr-b", "level", at), 100+0.01*(5-20)*s)
		near(t, "mean flow", w.StateValue("hdr-b", "avg_flow", at), defaults.value("avg", s))
	}
}

func TestARetiredUnitLeavesTheAggregatesAfterItsRetirement(t *testing.T) {
	t.Parallel()
	w := coupledWorld(t)
	w.Retire("unit-1", "test", atSeconds(500))
	r := newHeaderReference(map[string]float64{"unit-1": 10, "unit-2": 10, "unit-3": 10}, map[string]float64{"unit-1": 500})
	for _, s := range []float64{400, 500, 600, 2000} {
		near(t, "inflow", w.StateValue("hdr-a", "inflow", atSeconds(s)), r.sum(s))
		near(t, "level", w.StateValue("hdr-a", "level", atSeconds(s)), r.level(s))
	}
}

func TestCouplingDoesNotDependOnTheOrderEntitiesAreRead(t *testing.T) {
	t.Parallel()
	const end = 4000.0
	schedules := map[string]func(w *World){
		"header only":       func(*World) {},
		"units first":       func(w *World) { readAll(w, coupledUnits, "flow", end) },
		"a unit far ahead":  func(w *World) { w.StateValue("unit-2", "flow", atSeconds(end)) },
		"units every 3 s":   func(w *World) { readSteps(w, "unit-3", "flow", 3, end) },
		"header every 7 s":  func(w *World) { readSteps(w, "hdr-a", "level", 7, end) },
		"mixed every 0.5 s": func(w *World) { readSteps(w, "unit-1", "flow", 0.5, 100) },
	}
	r := newHeaderReference(map[string]float64{"unit-1": 10, "unit-2": 10, "unit-3": 10}, nil)
	for name, schedule := range schedules {
		w := coupledWorld(t)
		schedule(w)
		at := atSeconds(end)
		if got, want := w.StateValue("hdr-a", "level", at), r.level(end); got != coupledLevel(t) || math.Abs(got-want) > 1e-9 {
			t.Fatalf("%s: level %v, header-only %v, reference %v", name, got, coupledLevel(t), want)
		}
		if got := w.StateValue("hdr-a", "avg_flow", at); math.Abs(got-r.average(end)) > 1e-9 {
			t.Fatalf("%s: mean flow %v, reference %v", name, got, r.average(end))
		}
	}
}

func coupledLevel(t *testing.T) float64 {
	return coupledWorld(t).StateValue("hdr-a", "level", atSeconds(4000))
}

func readAll(w *World, entities []string, state string, s float64) {
	for _, id := range entities {
		w.StateValue(id, state, atSeconds(s))
	}
}

func readSteps(w *World, entity, state string, every, until float64) {
	for s := 0.0; s <= until; s += every {
		w.StateValue(entity, state, atSeconds(s))
	}
}
