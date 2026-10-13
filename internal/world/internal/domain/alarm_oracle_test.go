package domain

import (
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func alarmWorld(t *testing.T) *World {
	t.Helper()
	return constWorld(t, func(s *model.DomainSpec) {
		kind := &s.Kinds[0]
		kind.State = append(kind.State, model.State{Name: "temp"}, model.State{Name: "acks"}, model.State{Name: "shelved"}, model.State{Name: "alarm"})
		alarm := &model.Alarm{Watch: "temp", Limit: ptr(80), Direction: model.AlarmHigh, Deadband: 2, OnDelayS: 30, OffDelayS: 20,
			Priority: "high", Acknowledge: "acks", Shelve: "shelved"}
		kind.Dynamics = append(kind.Dynamics,
			model.Dynamics{Target: "temp", Tier: "F0", F0: &model.F0Dyn{Schedule: schedulePoints(0, 50, 100, 85, 300, 79, 400, 70, 500, 90, 800, 50)}},
			model.Dynamics{Target: "acks", Tier: "F0", F0: &model.F0Dyn{Schedule: schedulePoints(0, 0, 200, 1, 450, 2)}},
			model.Dynamics{Target: "shelved", Tier: "F0", F0: &model.F0Dyn{Schedule: schedulePoints(0, 0, 600, 1, 700, 0)}},
			model.Dynamics{Target: "alarm", Tier: "F1", DTMs: 10000, F1: &model.F1Dyn{Form: model.FormAlarm, Alarm: alarm}},
		)
	}, model.DefaultStartTimeNS)
}

func TestAnAlarmFollowsTheISA182StateMachine(t *testing.T) {
	t.Parallel()
	expected := []struct {
		at   float64
		code float64
		why  string
	}{
		{90, model.AlarmNormal, "below the limit"},
		{135, model.AlarmNormal, "above the limit for less than the 30 s on-delay"},
		{140, model.AlarmUnacked, "on-delay elapsed at the 130 s step"},
		{205, model.AlarmUnacked, "the acknowledgement lands at the 200 s step"},
		{210, model.AlarmAcked, "acknowledged"},
		{350, model.AlarmAcked, "79 is inside the 2-unit deadband"},
		{425, model.AlarmAcked, "clear for less than the 20 s off-delay"},
		{430, model.AlarmNormal, "cleared after being acknowledged"},
		{540, model.AlarmUnacked, "a new activation needs a new acknowledgement"},
		{610, model.AlarmShelved, "shelved"},
		{710, model.AlarmUnacked, "unshelved while still active"},
		{830, model.AlarmReturned, "returned to normal unacknowledged"},
		{1200, model.AlarmReturned, "stays until acknowledged"},
	}
	w := alarmWorld(t)
	for _, e := range expected {
		if got := w.StateValue("e-1", "alarm", atSeconds(e.at)); got != e.code {
			t.Fatalf("at %v s the alarm is %v, want %v: %s", e.at, got, e.code, e.why)
		}
	}
	fresh := alarmWorld(t)
	if got := fresh.StateValue("e-1", "alarm", atSeconds(830)); got != model.AlarmReturned {
		t.Fatalf("a world read first at 830 s gives %v, the same history gives %v", got, model.AlarmReturned)
	}
}
