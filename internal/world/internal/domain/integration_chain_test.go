package domain

import (
	"math"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/randutil"
)

type lagLink struct {
	name  string
	dtS   float64
	tau   float64
	input string
}

type lagReference struct {
	links map[string]lagLink
	u     float64
	grids map[string][]float64
}

func newLagReference(links []lagLink, u float64) *lagReference {
	r := &lagReference{links: map[string]lagLink{}, u: u, grids: map[string][]float64{}}
	for _, link := range links {
		r.links[link.name] = link
	}
	return r
}

func (r *lagReference) input(link lagLink, t float64) float64 {
	if link.input == "u" {
		return r.u
	}
	return r.value(link.input, t)
}

func (r *lagReference) grid(name string, j int) float64 {
	values := r.grids[name]
	for len(values) <= j {
		k := len(values)
		if k == 0 {
			values = append(values, 0)
			continue
		}
		link := r.links[name]
		values = append(values, lagStep(values[k-1], r.input(link, float64(k-1)*link.dtS), link.tau, link.dtS))
		r.grids[name] = values
	}
	r.grids[name] = values
	return values[j]
}

func lagStep(x, u, tau, h float64) float64 {
	f := func(x float64) float64 { return (u - x) / tau }
	k1 := f(x)
	k2 := f(x + h/2*k1)
	k3 := f(x + h/2*k2)
	k4 := f(x + h*k3)
	return x + h/6*(k1+2*k2+2*k3+k4)
}

func (r *lagReference) value(name string, t float64) float64 {
	link := r.links[name]
	j := int(math.Floor(t/link.dtS + 1e-12))
	t0 := float64(j) * link.dtS
	x := r.grid(name, j)
	if t-t0 < 1e-12 {
		return x
	}
	return lagStep(x, r.input(link, t0), link.tau, t-t0)
}

func lagChainWorld(t *testing.T, links []lagLink, u float64) *World {
	t.Helper()
	return constWorld(t, func(s *model.DomainSpec) {
		kind := &s.Kinds[0]
		kind.State[0].Initial = u
		for _, link := range links {
			kind.State = append(kind.State, model.State{Name: link.name})
			kind.Dynamics = append(kind.Dynamics, model.Dynamics{Target: link.name, Tier: "F1", DTMs: int64(link.dtS * 1000),
				F1: &model.F1Dyn{Form: "first_order_lag", TimeConstantS: link.tau, Gain: 1, Inputs: []model.F1Input{{State: link.input, Coef: 1, CoefSet: true}}}})
		}
	}, model.DefaultStartTimeNS)
}

func TestChainsOfDifferingStepsAgreeWithAnIndependentRecursionWhateverTheReadSchedule(t *testing.T) {
	t.Parallel()
	chains := map[string][]lagLink{
		"a1-b60-c45":     {{"c", 45, 50, "b"}, {"b", 60, 80, "a"}, {"a", 1, 40, "u"}},
		"a1-b60-c45-d30": {{"d", 30, 50, "c"}, {"c", 45, 50, "b"}, {"b", 60, 80, "a"}, {"a", 1, 40, "u"}},
		"a1-b7-c3-d4":    {{"c", 3, 50, "b"}, {"b", 7, 80, "a"}, {"a", 1, 40, "u"}, {"d", 4, 60, "c"}},
	}
	for name, links := range chains {
		for _, every := range []int64{0, 1e9, 7e9} {
			reference := newLagReference(links, 200)
			for _, probe := range []float64{12, 13.5, 100, 250, 1000} {
				w := lagChainWorld(t, links, 200)
				readEvery(w, links[len(links)-1].name, every, probe)
				for _, link := range links {
					got := w.StateValue("e-1", link.name, model.DefaultStartTimeNS+int64(probe*1e9))
					if want := reference.value(link.name, probe); math.Abs(got-want) > 1e-9 {
						t.Fatalf("%s, reads every %d ns: %s(%v s) = %v, reference %v", name, every, link.name, probe, got, want)
					}
				}
			}
		}
	}
}

func readEvery(w *World, state string, every int64, untilS float64) {
	if every == 0 {
		return
	}
	for at := model.DefaultStartTimeNS; at < model.DefaultStartTimeNS+int64(untilS*1e9); at += every {
		w.StateValue("e-1", state, at)
	}
}

func TestRandomLagChainsAgreeWithAnIndependentRecursion(t *testing.T) {
	t.Parallel()
	rng := randutil.NewSplitMix64(11)
	steps := []float64{0.5, 1, 2, 3, 5, 7, 10, 45, 60}
	names := []string{"a", "b", "c", "d", "e"}
	for trial := range 200 {
		var links []lagLink
		for i, name := range names[:2+rng.Intn(4)] {
			input := "u"
			if i > 0 {
				input = names[rng.Intn(i)]
			}
			links = append([]lagLink{{name, steps[rng.Intn(len(steps))], 20 + float64(rng.Intn(480)), input}}, links...)
		}
		reference, w := newLagReference(links, 200), lagChainWorld(t, links, 200)
		probe := 30 + rng.Float64()*400
		for _, link := range links {
			got := w.StateValue("e-1", link.name, model.DefaultStartTimeNS+int64(probe*1e9))
			if want := reference.value(link.name, probe); math.Abs(got-want) > 1e-6 {
				t.Fatalf("trial %d %v: %s(%v s) = %v, reference %v", trial, links, link.name, probe, got, want)
			}
		}
	}
}
