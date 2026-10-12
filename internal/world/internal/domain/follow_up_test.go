package domain

import (
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func followUpEffector(name string, schema map[string]any, deltas []model.StateDelta) model.Effector {
	return model.Effector{Name: name, ArgsSchema: map[string]any{"type": "object", "properties": schema},
		Ack:    model.Ack{LatencyMS: model.Latency{Mean: 1}, FailureModes: []model.FailureMode{{Mode: "ok", Probability: 1}}},
		Effect: model.Effect{StateDeltas: deltas}}
}

func followUpWorld(t *testing.T) *World {
	t.Helper()
	number := map[string]any{"type": "number"}
	return constWorld(t, func(s *model.DomainSpec) {
		kind := &s.Kinds[0]
		kind.State = append(kind.State, model.State{Name: "pressure", Initial: 7}, model.State{Name: "band_low", Initial: 6.5},
			model.State{Name: "band_high", Initial: 7.5}, model.State{Name: "stages", Initial: 0}, model.State{Name: "repaired", Initial: 0},
			model.State{Name: "wasted", Initial: 0}, model.State{Name: "fan", Initial: 0}, model.State{Name: "heat", Initial: 0}, model.State{Name: "orders", Initial: 0})
		kind.Dynamics = append(kind.Dynamics, model.Dynamics{Target: "stages", Tier: "F1", DTMs: 1000, F1: &model.F1Dyn{Form: model.FormSequencer,
			StageUp: &model.Condition{State: "pressure", Operator: "lt", ThresholdState: "band_low"}, StageDown: &model.Condition{State: "pressure", Operator: "gt", ThresholdState: "band_high"},
			Stages: 3, HoldS: 60}})
		order := followUpEffector("work_order", map[string]any{"cause": map[string]any{"type": "string"}}, []model.StateDelta{{State: "orders", Delta: 1}})
		order.Responder = &model.Responder{MatchArg: "cause", DelayS: 900, ClearFault: true,
			OnMatch: []model.StateDelta{{State: "repaired", Delta: 1}}, OnMiss: []model.StateDelta{{State: "wasted", Delta: 1}}}
		fan := followUpEffector("run_fan", map[string]any{"speed": number, "lease_s": number}, []model.StateDelta{{State: "fan", AssignFromArg: "speed"}})
		fan.Lease = &model.Lease{Arg: "lease_s", Fallback: model.FallbackSafeOff, SafeDeltas: []model.StateDelta{{State: "fan", Delta: 0}}}
		heater := followUpEffector("heat", map[string]any{"lease_s": number}, []model.StateDelta{{State: "heat", Delta: 2}})
		heater.Lease = &model.Lease{Arg: "lease_s", Fallback: model.FallbackRevert}
		kind.Effectors = []model.Effector{order, fan, heater,
			followUpEffector("drop", nil, []model.StateDelta{{State: "pressure", Delta: -1.5}}), followUpEffector("rise", nil, []model.StateDelta{{State: "pressure", Delta: 3}})}
	}, model.DefaultStartTimeNS)
}

func call(t *testing.T, w *World, effector, id string, seconds float64, args map[string]any) {
	t.Helper()
	if _, err := w.InvokeEffector(effector, "e-1", id, args, at(seconds)); err != nil {
		t.Fatal(err)
	}
}

func TestASequencerStagesOncePerHoldWithinItsBand(t *testing.T) {
	t.Parallel()
	w := followUpWorld(t)
	call(t, w, "drop", "d", 100, nil)
	call(t, w, "rise", "r", 400, nil)
	for _, tc := range []struct{ seconds, stages float64 }{{101, 0}, {102, 1}, {161, 1}, {162, 2}, {222, 3}, {350, 3}, {402, 2}, {462, 1}, {522, 0}, {700, 0}} {
		if got := w.StateValue("e-1", "stages", at(tc.seconds)); got != tc.stages {
			t.Fatalf("t=%v: %v stages, want %v", tc.seconds, got, tc.stages)
		}
	}
}

func TestAResponderActsOnTheFaultItIsToldAbout(t *testing.T) {
	t.Parallel()
	w := followUpWorld(t)
	if _, err := w.InjectFault("e-1", "f1", at(0), nil); err != nil {
		t.Fatal(err)
	}
	call(t, w, "work_order", "miss", 100, map[string]any{"cause": "pump_wear"})
	call(t, w, "work_order", "match", 200, map[string]any{"cause": "f1"})
	if repaired, wasted := w.StateValue("e-1", "repaired", at(1050)), w.StateValue("e-1", "wasted", at(1050)); repaired != 0 || wasted != 1 {
		t.Fatalf("at 1050 s repaired %v wasted %v, want the miss only", repaired, wasted)
	}
	if repaired := w.StateValue("e-1", "repaired", at(1101)); repaired != 1 || len(w.ViolatedBoundaries("e-1", at(1101))) != 0 {
		t.Fatalf("at 1101 s repaired %v, want the match", repaired)
	}
	if active := w.activeInjections("e-1", "f1", at(1101)); len(active) != 0 {
		t.Fatalf("faults after the repair = %v", active)
	}
	if active := w.activeInjections("e-1", "f1", at(1099)); len(active) != 1 {
		t.Fatalf("faults before the responder arrives = %v", active)
	}
}

func TestALeaseFallsBackUnlessRenewed(t *testing.T) {
	t.Parallel()
	lapsed := followUpWorld(t)
	call(t, lapsed, "run_fan", "a", 100, map[string]any{"speed": 5.0, "lease_s": 60.0})
	call(t, lapsed, "heat", "h", 100, map[string]any{"lease_s": 60.0})
	renewed := followUpWorld(t)
	call(t, renewed, "run_fan", "a", 100, map[string]any{"speed": 5.0, "lease_s": 60.0})
	call(t, renewed, "run_fan", "b", 150, map[string]any{"speed": 5.0, "lease_s": 60.0})
	call(t, renewed, "run_fan", "c", 160, map[string]any{"speed": 5.0})
	call(t, renewed, "heat", "h1", 100, map[string]any{"lease_s": 60.0})
	call(t, renewed, "heat", "h2", 150, map[string]any{"lease_s": 60.0})
	for _, tc := range []struct {
		w             *World
		state         string
		seconds, want float64
	}{{lapsed, "fan", 150, 5}, {lapsed, "fan", 161, 0}, {lapsed, "heat", 150, 2}, {lapsed, "heat", 161, 0}, {renewed, "fan", 200, 5}, {renewed, "fan", 211, 0},
		{renewed, "heat", 200, 4}, {renewed, "heat", 211, 0}, {renewed, "heat", 1000, 0}} {
		if got := tc.w.StateValue("e-1", tc.state, at(tc.seconds)); got != tc.want {
			t.Fatalf("%s at %v s = %v, want %v", tc.state, tc.seconds, got, tc.want)
		}
	}
}

func TestALeaseInMillisecondsLapsesOnTime(t *testing.T) {
	t.Parallel()
	w := constWorld(t, func(s *model.DomainSpec) {
		kind := &s.Kinds[0]
		kind.State = append(kind.State, model.State{Name: "fan", Initial: 0})
		fan := followUpEffector("run_fan", map[string]any{"speed": map[string]any{"type": "number"}, "lease_ms": map[string]any{"type": "number"}}, []model.StateDelta{{State: "fan", AssignFromArg: "speed"}})
		fan.Lease = &model.Lease{Arg: "lease_ms", Unit: "ms", Fallback: model.FallbackSafeOff, SafeDeltas: []model.StateDelta{{State: "fan", Delta: 0}}}
		kind.Effectors = []model.Effector{fan}
	}, model.DefaultStartTimeNS)
	call(t, w, "run_fan", "a", 100, map[string]any{"speed": 5.0, "lease_ms": 30000.0})
	if before, after := w.StateValue("e-1", "fan", at(129)), w.StateValue("e-1", "fan", at(131)); before != 5 || after != 0 {
		t.Fatalf("fan %v before and %v after a 30000 ms lease", before, after)
	}
}
