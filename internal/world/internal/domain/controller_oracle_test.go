package domain

import (
	"math"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

type loopCase struct {
	name       string
	pid        model.Controller
	plantGain  float64
	low, high  float64
	sp, manual func(s float64) float64
}

func stepSchedule(points ...float64) func(s float64) float64 {
	return func(s float64) float64 {
		value := points[1]
		for i := 0; i+1 < len(points); i += 2 {
			if s >= points[i] {
				value = points[i+1]
			}
		}
		return value
	}
}

func schedulePoints(points ...float64) *model.Schedule {
	schedule := &model.Schedule{Interpolation: "step"}
	for i := 0; i+1 < len(points); i += 2 {
		schedule.Points = append(schedule.Points, model.SchedulePoint{AtS: points[i], Value: points[i+1]})
	}
	return schedule
}

func loopReference(lc loopCase, until int) []float64 {
	pv, visible, integral, previous := 0.0, 0.0, 0.0, 0.0
	trace := []float64{pv}
	sign := lc.pid.Sign()
	for k := 0; k < until; k++ {
		t := float64(k)
		e := sign * (lc.sp(t) - pv)
		p, d := lc.pid.Kp*e, -sign*lc.pid.Kd*(pv-previous)
		previous = pv
		var u float64
		if lc.manual(t) >= 0.5 {
			u = math.Min(math.Max(3, lc.low), lc.high)
			integral = u - p - d
		} else {
			if next := integral + lc.pid.Ki*e; !((p+next+d > lc.high && e > 0) || (p+next+d < lc.low && e < 0)) {
				integral = next
			}
			u = math.Min(math.Max(p+integral+d, lc.low), lc.high)
		}
		pv = lagStep(pv, lc.plantGain*visible, 100, 1)
		visible = u
		trace = append(trace, pv)
	}
	return trace
}

func loopWorld(t *testing.T, lc loopCase, spPoints, modePoints []float64) *World {
	t.Helper()
	return constWorld(t, func(s *model.DomainSpec) {
		kind := &s.Kinds[0]
		pid := lc.pid
		pid.PV, pid.SP, pid.Manual, pid.ManualOutput = "pv", "sp", "mode", "manual_out"
		kind.State = append(kind.State, model.State{Name: "sp"}, model.State{Name: "pv"}, model.State{Name: "out"}, model.State{Name: "mode"}, model.State{Name: "manual_out", Initial: 3})
		kind.Dynamics = append(kind.Dynamics,
			model.Dynamics{Target: "sp", Tier: "F0", F0: &model.F0Dyn{Schedule: schedulePoints(spPoints...)}},
			model.Dynamics{Target: "mode", Tier: "F0", F0: &model.F0Dyn{Schedule: schedulePoints(modePoints...)}},
			model.Dynamics{Target: "pv", Tier: "F1", DTMs: 1000, F1: &model.F1Dyn{Form: "first_order_lag", Gain: lc.plantGain, TimeConstantS: 100, Inputs: []model.F1Input{{State: "out", Coef: 1, CoefSet: true}}}},
			model.Dynamics{Target: "out", Tier: "F1", DTMs: 1000, F1: &model.F1Dyn{Form: model.FormPID, PID: &pid, Clamp: &model.Clamp{Min: ptr(lc.low), Max: ptr(lc.high)}}},
		)
	}, model.DefaultStartTimeNS)
}

func TestAClosedLoopAgreesWithAnIndependentRecurrence(t *testing.T) {
	t.Parallel()
	never := []float64{0, 0}
	cases := []struct {
		lc               loopCase
		spPoints, manual []float64
	}{
		{loopCase{name: "pid", pid: model.Controller{Kp: 2, Ki: 0.05, Kd: 0.5, Action: model.ActionReverse}, plantGain: 1, low: 0, high: 100}, []float64{0, 10}, never},
		{loopCase{name: "anti-windup", pid: model.Controller{Kp: 1, Ki: 0.1, Action: model.ActionReverse}, plantGain: 1, low: 0, high: 5}, []float64{0, 10, 600, 2}, never},
		{loopCase{name: "manual and back", pid: model.Controller{Kp: 2, Ki: 0.05, Action: model.ActionReverse}, plantGain: 1, low: 0, high: 100}, []float64{0, 10}, []float64{0, 0, 100, 1, 200, 0}},
		{loopCase{name: "direct action", pid: model.Controller{Kp: 1.5, Ki: 0.02, Action: model.ActionDirect}, plantGain: -1, low: 0, high: 50}, []float64{0, -5}, never},
	}
	for _, tc := range cases {
		tc.lc.sp, tc.lc.manual = stepSchedule(tc.spPoints...), stepSchedule(tc.manual...)
		reference := loopReference(tc.lc, 1200)
		w := loopWorld(t, tc.lc, tc.spPoints, tc.manual)
		for _, s := range []int{1, 2, 37, 99, 100, 101, 150, 201, 599, 600, 601, 602, 900, 1200} {
			if got := w.StateValue("e-1", "pv", model.DefaultStartTimeNS+int64(s)*1e9); math.Abs(got-reference[s]) > 1e-9 {
				t.Fatalf("%s: pv at %d s = %v, reference %v", tc.lc.name, s, got, reference[s])
			}
		}
	}
}

func TestAControllerLeavesSaturationAsSoonAsItsErrorReverses(t *testing.T) {
	t.Parallel()
	lc := loopCase{pid: model.Controller{Kp: 1, Ki: 0.1, Action: model.ActionReverse}, plantGain: 1, low: 0, high: 5}
	w := loopWorld(t, lc, []float64{0, 10, 600, 2}, []float64{0, 0})
	if out := w.StateValue("e-1", "out", model.DefaultStartTimeNS+599e9); out != 5 {
		t.Fatalf("the output saturates at 5 before the setpoint drops, got %v", out)
	}
	if out := w.StateValue("e-1", "out", model.DefaultStartTimeNS+602e9); out >= 5 {
		t.Fatalf("with anti-windup the output leaves saturation within a step of the drop, got %v", out)
	}
}
