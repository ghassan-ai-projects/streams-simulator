package domain

import (
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func kpiWorld(t *testing.T, mutate func(*model.DomainSpec)) *World {
	t.Helper()
	w, err := New(testSpec(t, func(s *model.DomainSpec) {
		alarmSpec(s)
		mutate(s)
	}), 42, "w-kpi", model.DefaultStartTimeNS, Options{EmitDisabled: true, TrackKPIs: true})
	if err != nil {
		t.Fatal(err)
	}
	return w
}

func alarmSpec(s *model.DomainSpec) {
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
	kind.KPIs = []string{"temp"}
}

type alarmReference struct {
	activations []float64
	activeS     float64
	longestS    float64
	activeAtEnd bool
}

func referenceAlarm(t *testing.T, until float64) alarmReference {
	t.Helper()
	w := alarmWorld(t)
	var ref alarmReference
	since := -1.0
	for s := 0.0; s <= until; s += 10 {
		code := w.StateValue("e-1", "alarm", atSeconds(s))
		active := code == model.AlarmUnacked || code == model.AlarmAcked
		if active && since < 0 && s > 0 {
			ref.activations = append(ref.activations, s)
		}
		if active && since < 0 {
			since = s
		}
		if !active && since >= 0 {
			ref.activeS, ref.longestS, since = ref.activeS+s-since, max(ref.longestS, s-since), -1
		}
	}
	if since >= 0 {
		ref.activeS, ref.longestS, ref.activeAtEnd = ref.activeS+until-since, max(ref.longestS, until-since), true
	}
	return ref
}

func TestTheKPITruthCountsAlarmActivationsAndActiveTimeOnTheAlarmScanGrid(t *testing.T) {
	t.Parallel()
	for _, until := range []float64{620, 1200} {
		w := kpiWorld(t, func(*model.DomainSpec) {})
		if _, _, err := w.Advance(atSeconds(until)); err != nil {
			t.Fatal(err)
		}
		ref := referenceAlarm(t, until)
		got := w.KPITruth().Entities[0].Alarms[0]
		if got.Activations != len(ref.activations) || got.ActiveS != ref.activeS || got.LongestActiveS != ref.longestS || got.ActiveAtEnd != ref.activeAtEnd || got.Chattering {
			t.Fatalf("until %v s: %+v, reference %+v", until, got, ref)
		}
	}
	w := kpiWorld(t, func(*model.DomainSpec) {})
	if _, _, err := w.Advance(atSeconds(1200)); err != nil {
		t.Fatal(err)
	}
	if system := w.KPITruth().AlarmSystem; system.Activations != 3 || system.PeakPer10Min != 3 || system.StandingAtEnd != 0 {
		t.Fatalf("alarm system = %+v, want the 3 activations, all within 10 minutes, none standing at 1200 s", system)
	}
	if got := len(referenceAlarm(t, 1200).activations); got != 3 {
		t.Fatalf("the scripted alarm activates %d times, want 3 (140 s, after the second rise, on unshelving at 710 s)", got)
	}
}

func TestTheKPITruthAveragesAKPIStateOverItsSamples(t *testing.T) {
	t.Parallel()
	w := kpiWorld(t, func(*model.DomainSpec) {})
	if _, _, err := w.Advance(atSeconds(1200)); err != nil {
		t.Fatal(err)
	}
	reference := alarmWorld(t)
	var sum, low, high float64
	low, high = 1e9, -1e9
	samples := 0
	for s := 0.0; s <= 1200; s += 10 {
		v := reference.StateValue("e-1", "temp", atSeconds(s))
		sum, low, high, samples = sum+v, min(low, v), max(high, v), samples+1
	}
	got := w.KPITruth().Entities[0].KPIs[0]
	if got.State != "temp" || got.Samples != samples || got.Mean != sum/float64(samples) || got.Min != low || got.Max != high {
		t.Fatalf("KPI = %+v, want mean %v over %d samples in [%v, %v]", got, sum/float64(samples), samples, low, high)
	}
}

func TestAlarmSystemKPIsFollowISA182(t *testing.T) {
	t.Parallel()
	minute := int64(60 * secondsPerNS)
	if !chattering([]int64{0, 20 * secondsPerNS, 59 * secondsPerNS}) || chattering([]int64{0, 30 * secondsPerNS, minute}) {
		t.Fatal("three activations inside one minute chatter; three spanning a full minute do not")
	}
	burst := make([]int64, 12)
	for i := range burst {
		burst[i] = 10*minute + int64(i)*30*secondsPerNS
	}
	if got := peakRate(append([]int64{0}, burst...)); got != 12 {
		t.Fatalf("peak per 10 minutes = %d, want the 12 of the burst", got)
	}
	if got := floodIntervals(burst); got != 1 {
		t.Fatalf("flood intervals = %d, want 1: more than 10 activations fall in one 10-minute interval", got)
	}
}

func TestAWorldWithoutKPITrackingHasNoKPITruth(t *testing.T) {
	t.Parallel()
	if truth := alarmWorld(t).KPITruth(); truth != nil {
		t.Fatalf("KPI truth = %+v, want none without TrackKPIs", truth)
	}
}

func TestARetiredEntitysAlarmEndsAtItsRetirementWhateverTheRunDoesAfter(t *testing.T) {
	t.Parallel()
	for _, until := range []float64{205, 215, 400} {
		w := kpiWorld(t, func(*model.DomainSpec) {})
		if _, _, err := w.Advance(atSeconds(190)); err != nil {
			t.Fatal(err)
		}
		w.Retire("e-1", "test", atSeconds(200))
		if _, _, err := w.Advance(atSeconds(until)); err != nil {
			t.Fatal(err)
		}
		alarm := w.KPITruth().Entities[0].Alarms[0]
		if alarm.ActiveS != 60 || alarm.LongestActiveS != 60 || alarm.ActiveAtEnd {
			t.Fatalf("run to %v s: %+v, want 60 s active from 140 s to the retirement at 200 s and no standing alarm", until, alarm)
		}
	}
}

func TestAFaultInjectedRightAfterAnAdvanceIsInTheSampleAtThatInstant(t *testing.T) {
	t.Parallel()
	at := atSeconds(600)
	scripted, chunked := kpiWorld(t, faultyTemp), kpiWorld(t, faultyTemp)
	if _, err := scripted.InjectFault("e-1", "hot", at, nil); err != nil {
		t.Fatal(err)
	}
	if _, _, err := scripted.Advance(at); err != nil {
		t.Fatal(err)
	}
	if _, _, err := chunked.Advance(at); err != nil {
		t.Fatal(err)
	}
	if _, err := chunked.InjectFault("e-1", "hot", 0, nil); err != nil {
		t.Fatal(err)
	}
	if got, want := chunked.KPITruth().Entities[0].KPIs[0], scripted.KPITruth().Entities[0].KPIs[0]; got != want {
		t.Fatalf("a fault injected after advancing to its onset gives %+v, scripted before the advance %+v", got, want)
	}
}

func faultyTemp(s *model.DomainSpec) {
	kind := &s.Kinds[0]
	kind.Faults = append(kind.Faults, model.Fault{ID: "hot", Affects: []model.FaultEffect{{State: "temp", Delta: 30}},
		Onset: model.FaultOnset{Shape: "step", Magnitude: 1}, Observability: kind.Faults[0].Observability})
}
