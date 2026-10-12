package domain

import (
	"math"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func storeWorld(t *testing.T, a0, b0, ca, cb, k, pump float64, flows []model.Flow, inflow float64) *World {
	t.Helper()
	return constWorld(t, func(s *model.DomainSpec) {
		kind := &s.Kinds[0]
		kind.State = append(kind.State, model.State{Name: "pump", Initial: pump}, model.State{Name: "a", Initial: a0}, model.State{Name: "b", Initial: b0})
		var inputs []model.F1Input
		if inflow != 0 {
			inputs = []model.F1Input{{State: "u", Coef: inflow, CoefSet: true}}
		}
		kind.Dynamics = append(kind.Dynamics,
			model.Dynamics{Target: "a", Tier: "F1", DTMs: 1000, F1: &model.F1Dyn{Form: model.FormStore, Capacity: ca, Inputs: inputs}},
			model.Dynamics{Target: "b", Tier: "F1", DTMs: 1000, F1: &model.F1Dyn{Form: model.FormStore, Capacity: cb}})
		kind.Flows = flows
	}, model.DefaultStartTimeNS)
}

func TestTwoStoresJoinedByAFlowAgreeWithTheClosedForm(t *testing.T) {
	t.Parallel()
	start := model.DefaultStartTimeNS
	for _, ca := range []float64{1, 4} {
		for _, cb := range []float64{2, 0.5} {
			for _, k := range []float64{0.002, 0.01} {
				for _, pump := range []float64{0.5, 1} {
					w := storeWorld(t, 10, 2, ca, cb, k, pump, []model.Flow{{Name: "transfer", From: "a", To: "b", Conductance: k, Scale: []string{"pump"}}}, 0)
					lambda := k * pump * (1/ca + 1/cb)
					volume := ca*10 + cb*2
					for _, seconds := range []float64{30, 600, 3600} {
						d := 8 * math.Exp(-lambda*seconds)
						wantA, wantB := (volume+cb*d)/(ca+cb), (volume-ca*d)/(ca+cb)
						at := start + int64(seconds*secondsPerNS)
						gotA, gotB := w.StateValue("e-1", "a", at), w.StateValue("e-1", "b", at)
						if math.Abs(gotA-wantA) > 1e-6 || math.Abs(gotB-wantB) > 1e-6 {
							t.Fatalf("ca=%v cb=%v k=%v pump=%v t=%v: got a=%v b=%v, closed form a=%v b=%v", ca, cb, k, pump, seconds, gotA, gotB, wantA, wantB)
						}
					}
				}
			}
		}
	}
}

func TestADrainedStoreWithAnInflowAgreesWithTheClosedForm(t *testing.T) {
	t.Parallel()
	start := model.DefaultStartTimeNS
	for _, k := range []float64{0.001, 0.02} {
		for _, inflow := range []float64{0.5, 3} {
			w := storeWorld(t, 5, 0, 2, 1, k, 1, []model.Flow{{Name: "outlet", From: "a", Conductance: k}}, inflow)
			steady := inflow * 2 / k
			for _, seconds := range []float64{60, 900, 7200} {
				want := steady + (5-steady)*math.Exp(-k*seconds/2)
				got := w.StateValue("e-1", "a", start+int64(seconds*secondsPerNS))
				if math.Abs(got-want) > 1e-6*math.Max(1, math.Abs(want)) {
					t.Fatalf("k=%v inflow=%v t=%v: got %v, closed form %v", k, inflow, seconds, got, want)
				}
			}
		}
	}
}

func TestCoupledStoresDoNotDependOnReadOrder(t *testing.T) {
	t.Parallel()
	start := model.DefaultStartTimeNS
	flows := []model.Flow{{Name: "transfer", From: "a", To: "b", Conductance: 0.01, Scale: []string{"pump"}}}
	at := start + 1800*int64(secondsPerNS) + 250e6
	aFirst := storeWorld(t, 10, 2, 1, 2, 0.01, 1, flows, 0)
	a1, b1 := aFirst.StateValue("e-1", "a", at), aFirst.StateValue("e-1", "b", at)
	bFirst := storeWorld(t, 10, 2, 1, 2, 0.01, 1, flows, 0)
	b2, a2 := bFirst.StateValue("e-1", "b", at), bFirst.StateValue("e-1", "a", at)
	often := storeWorld(t, 10, 2, 1, 2, 0.01, 1, flows, 0)
	for read := start; read < at; read += 7300e6 {
		often.StateValue("e-1", "b", read)
	}
	a3, b3 := often.StateValue("e-1", "a", at), often.StateValue("e-1", "b", at)
	if a1 != a2 || a1 != a3 || b1 != b2 || b1 != b3 {
		t.Fatalf("read order changed the levels: a %v %v %v, b %v %v %v", a1, a2, a3, b1, b2, b3)
	}
}
