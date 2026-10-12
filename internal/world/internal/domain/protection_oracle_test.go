package domain

import (
	"math"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func protectionEffector(name, state string, delta float64, requires ...model.Condition) model.Effector {
	return model.Effector{Name: name, ArgsSchema: map[string]any{"type": "object"}, Requires: requires,
		Ack:    model.Ack{LatencyMS: model.Latency{Mean: 1}, FailureModes: []model.FailureMode{{Mode: "ok", Probability: 1}}},
		Effect: model.Effect{StateDeltas: []model.StateDelta{{State: state, Delta: delta}}}}
}

func protectionWorld(t *testing.T, latch bool) *World {
	t.Helper()
	return constWorld(t, func(s *model.DomainSpec) {
		kind := &s.Kinds[0]
		kind.State = append(kind.State, model.State{Name: "temp", Initial: 0}, model.State{Name: "flow", Initial: 100},
			model.State{Name: "resets", Initial: 0}, model.State{Name: "low_flow", Initial: 0}, model.State{Name: "hot", Initial: 0},
			model.State{Name: "running", Initial: 1}, model.State{Name: "recycle", Initial: 0}, model.State{Name: "isolated", Initial: 0})
		kind.Dynamics = append(kind.Dynamics,
			model.Dynamics{Target: "temp", Tier: "F0", F0: &model.F0Dyn{Baseline: 50, TrendPerHour: 36}},
			model.Dynamics{Target: "hot", Tier: "F1", DTMs: 1000, F1: &model.F1Dyn{Form: model.FormTrip, When: &model.Condition{State: "temp", Operator: "gt", Threshold: 60}, ForS: 30, TrippedValue: 1}},
			model.Dynamics{Target: "low_flow", Tier: "F1", DTMs: 1000, F1: &model.F1Dyn{Form: model.FormTrip, When: &model.Condition{State: "flow", Operator: "lt", Threshold: 50}, ForS: 10, TrippedValue: 1, Latch: latch, ResetOn: "resets"}},
			model.Dynamics{Target: "recycle", Tier: "F1", DTMs: 1000, F1: &model.F1Dyn{Form: model.FormAntiRecycle, Watch: "running", Edge: model.EdgeFall, HoldS: 600}})
		kind.Boundaries = []model.Boundary{{Name: "oil_hot", State: "hot", Severity: "critical"}}
		kind.Effectors = []model.Effector{
			protectionEffector("cut", "flow", -80), protectionEffector("restore", "flow", 80), protectionEffector("reset", "resets", 1),
			protectionEffector("stop", "running", -1), protectionEffector("isolate", "isolated", 1),
			protectionEffector("start", "running", 1, model.Condition{State: "recycle", Operator: "lte", Threshold: 0},
				model.Condition{State: "low_flow", Operator: "lt", Threshold: 0.5}, model.Condition{State: "isolated", Operator: "lt", Threshold: 0.5}),
		}
	}, model.DefaultStartTimeNS)
}

func at(seconds float64) int64 { return model.DefaultStartTimeNS + int64(seconds*secondsPerNS) }

func invoke(t *testing.T, w *World, effector string, seconds float64) *InvokeResult {
	t.Helper()
	result, err := w.InvokeEffector(effector, "e-1", effector+"-"+string(rune('a'+int(seconds)%26))+string(rune('a'+int(seconds/26)%26)), map[string]any{}, at(seconds))
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func TestATripFiresOnTheStepAfterItsConditionHeldForItsDuration(t *testing.T) {
	t.Parallel()
	w := protectionWorld(t, false)
	crossing := (60.0 - 50) / (36.0 / 3600)
	firstSample := math.Floor(crossing) + 1
	visible := firstSample + 30 + 1
	if got := w.StateValue("e-1", "hot", at(visible-0.5)); got != 0 {
		t.Fatalf("tripped at %v s, before %v s", visible-0.5, visible)
	}
	if got := w.StateValue("e-1", "hot", at(visible)); got != 1 {
		t.Fatalf("not tripped at %v s (analytic crossing %v s)", visible, crossing)
	}
	if got := w.ViolatedBoundaries("e-1", at(visible)); len(got) != 1 || got[0] != "oil_hot" {
		t.Fatalf("violated boundaries = %v, want the trip's boundary", got)
	}
	if got := w.ViolatedBoundaries("e-1", at(visible-1)); len(got) != 0 {
		t.Fatalf("violated boundaries before the trip = %v", got)
	}
}

func TestALatchedTripHoldsUntilResetAndAnUnlatchedOneClears(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		latch               bool
		afterRestore, atEnd float64
	}{{latch: true, afterRestore: 1, atEnd: 0}, {latch: false, afterRestore: 0, atEnd: 0}} {
		w := protectionWorld(t, tc.latch)
		invoke(t, w, "cut", 100)
		if got := w.StateValue("e-1", "low_flow", at(111)); got != 0 {
			t.Fatalf("latch=%v: low flow tripped at 111 s, a step early", tc.latch)
		}
		if got := w.StateValue("e-1", "low_flow", at(112)); got != 1 {
			t.Fatalf("latch=%v: low flow not tripped at 112 s", tc.latch)
		}
		invoke(t, w, "restore", 200)
		if got := w.StateValue("e-1", "low_flow", at(250)); got != tc.afterRestore {
			t.Fatalf("latch=%v: after restore %v, want %v", tc.latch, got, tc.afterRestore)
		}
		invoke(t, w, "reset", 300)
		if got := w.StateValue("e-1", "low_flow", at(302)); got != tc.atEnd {
			t.Fatalf("latch=%v: after reset %v, want %v", tc.latch, got, tc.atEnd)
		}
	}
}

func TestPreconditionsRefuseALockedOutRecyclingOrIsolatedDevice(t *testing.T) {
	t.Parallel()
	w := protectionWorld(t, true)
	invoke(t, w, "stop", 100)
	if left := w.StateValue("e-1", "recycle", at(102)); left != 599 {
		t.Fatalf("anti-recycle left %v s at 102 s, want 599", left)
	}
	if result := invoke(t, w, "start", 300); result.Accepted || result.Reason != "effector_refused" {
		t.Fatalf("start during anti-recycle = %+v, want the generic refusal", result)
	}
	if calls := w.EffectorCalls(); calls[len(calls)-1].Reason != ReasonPreconditionRefused {
		t.Fatalf("the director's call log reason = %q, want %q", calls[len(calls)-1].Reason, ReasonPreconditionRefused)
	}
	if result := invoke(t, w, "start", 800); !result.Accepted {
		t.Fatalf("start after anti-recycle = %+v", result)
	}
	invoke(t, w, "cut", 900)
	if result := invoke(t, w, "start", 1000); result.Accepted {
		t.Fatalf("start while locked out = %+v", result)
	}
	invoke(t, w, "isolate", 1100)
	if result := invoke(t, w, "start", 1200); result.Accepted {
		t.Fatalf("start while isolated = %+v", result)
	}
}

func TestACallDatedAheadOfTheClockDoesNotChangeEarlierReadings(t *testing.T) {
	t.Parallel()
	ahead := protectionWorld(t, true)
	invoke(t, ahead, "cut", 100)
	if result := invoke(t, ahead, "start", 900); result.Accepted {
		t.Fatalf("start while locked out = %+v", result)
	}
	clean := protectionWorld(t, true)
	invoke(t, clean, "cut", 100)
	for _, seconds := range []float64{50, 105, 111, 112, 500} {
		if got, want := ahead.StateValue("e-1", "low_flow", at(seconds)), clean.StateValue("e-1", "low_flow", at(seconds)); got != want {
			t.Fatalf("t=%v: %v after a call dated 900 s, %v without it", seconds, got, want)
		}
	}
}

func TestReadingBoundariesAheadOfTheClockChangesNothingEarlier(t *testing.T) {
	t.Parallel()
	w := protectionWorld(t, true)
	if got := w.ViolatedBoundaries("e-1", at(5000)); len(got) != 1 {
		t.Fatalf("violated at 5000 s = %v, want the oil boundary", got)
	}
	invoke(t, w, "cut", 100)
	if got := w.StateValue("e-1", "low_flow", at(112)); got != 1 {
		t.Fatalf("low flow at 112 s = %v after an earlier read ahead of the clock, want 1", got)
	}
}
