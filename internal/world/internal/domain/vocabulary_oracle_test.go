package domain

import (
	"math"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func scheduleWorld(t *testing.T, schedule *model.Schedule) *World {
	t.Helper()
	return constWorld(t, func(s *model.DomainSpec) {
		kind := &s.Kinds[0]
		kind.State = append(kind.State, model.State{Name: "price", Initial: 0})
		kind.Dynamics = append(kind.Dynamics, model.Dynamics{Target: "price", Tier: "F0", F0: &model.F0Dyn{Schedule: schedule}})
	}, model.DefaultStartTimeNS)
}

func TestAScheduleIsATableLookup(t *testing.T) {
	t.Parallel()
	day := model.DefaultStartTimeNS
	tariff := scheduleWorld(t, &model.Schedule{PeriodS: 86400, Interpolation: model.InterpolateStep,
		Points: []model.SchedulePoint{{AtS: 0, Value: 0.10}, {AtS: 25200, Value: 0.30}, {AtS: 75600, Value: 0.12}}})
	ramp := scheduleWorld(t, &model.Schedule{PeriodS: 86400, Interpolation: model.InterpolateLinear,
		Points: []model.SchedulePoint{{AtS: 21600, Value: 10}, {AtS: 43200, Value: 20}}})
	once := scheduleWorld(t, &model.Schedule{Interpolation: model.InterpolateLinear,
		Points: []model.SchedulePoint{{AtS: 0, Value: 1}, {AtS: 3600, Value: 3}}})
	cases := []struct {
		w       *World
		seconds float64
		want    float64
	}{
		{tariff, 25199, 0.10}, {tariff, 25200, 0.30}, {tariff, 75599, 0.30}, {tariff, 86399, 0.12}, {tariff, 86400 + 25200, 0.30},
		{ramp, 0, 20 - 10*43200.0/64800}, {ramp, 32400, 15}, {ramp, 64800, 20 - 10*21600.0/64800},
		{once, 1800, 2}, {once, 7200, 3},
	}
	for _, tc := range cases {
		if got := tc.w.StateValue("e-1", "price", day+int64(tc.seconds*secondsPerNS)); math.Abs(got-tc.want) > 1e-9 {
			t.Errorf("at %v s: got %v, table %v", tc.seconds, got, tc.want)
		}
	}
}

func TestARegimeSwitchesAtTheAnalyticCrossing(t *testing.T) {
	t.Parallel()
	start := model.DefaultStartTimeNS
	w := constWorld(t, func(s *model.DomainSpec) {
		kind := &s.Kinds[0]
		kind.State = append(kind.State, model.State{Name: "oil", Initial: 0})
		kind.Dynamics = append(kind.Dynamics, model.Dynamics{Target: "oil", Tier: "F1", DTMs: 1000, F1: &model.F1Dyn{
			Form: model.FormRegime, Selector: "oil", Threshold: 3, Inputs: []model.F1Input{{State: "u", Coef: 1, CoefSet: true}},
			Below: &model.Regime{Gain: 2, TimeConstantS: 600}, Above: &model.Regime{Gain: 5, TimeConstantS: 200}}})
	}, start)
	crossing := 600 * math.Log(4.0)
	switchAt := math.Ceil(crossing)
	atSwitch := 4 * (1 - math.Exp(-switchAt/600))
	closed := func(seconds float64) float64 {
		if seconds <= switchAt {
			return 4 * (1 - math.Exp(-seconds/600))
		}
		return 10 + (atSwitch-10)*math.Exp(-(seconds-switchAt)/200)
	}
	for _, seconds := range []float64{300, crossing - 1, switchAt, switchAt + 60, 1800, 3600} {
		got := w.StateValue("e-1", "oil", start+int64(seconds*secondsPerNS))
		if math.Abs(got-closed(seconds)) > 1e-6 {
			t.Fatalf("t=%v: got %v, closed form %v (crossing %v s)", seconds, got, closed(seconds), crossing)
		}
	}
}

func TestADemandPeakIsTheLargestBlockMeanWithARatchet(t *testing.T) {
	t.Parallel()
	start := model.DefaultStartTimeNS
	power := &model.Schedule{PeriodS: 86400, Interpolation: model.InterpolateStep,
		Points: []model.SchedulePoint{{AtS: 0, Value: 50}, {AtS: 25200, Value: 120}, {AtS: 25800, Value: 80}}}
	w := constWorld(t, func(s *model.DomainSpec) {
		kind := &s.Kinds[0]
		kind.State = append(kind.State, model.State{Name: "power", Initial: 0}, model.State{Name: "demand", Initial: 0})
		kind.Dynamics = append(kind.Dynamics,
			model.Dynamics{Target: "power", Tier: "F0", F0: &model.F0Dyn{Schedule: power}},
			model.Dynamics{Target: "demand", Tier: "F1", DTMs: 60000, F1: &model.F1Dyn{Form: model.FormPeak,
				Inputs: []model.F1Input{{State: "power", Coef: 1, CoefSet: true}}, BlockS: 900, PeriodS: 86400, Ratchet: 0.5, PreviousPeak: 100}})
	}, start)
	for _, seconds := range []float64{3600, 25500, 27000, 86400 + 3600, 86400 + 30000} {
		got := w.StateValue("e-1", "demand", start+int64(seconds*secondsPerNS))
		if want := recomputedDemand(power, seconds); math.Abs(got-want) > 1e-9 {
			t.Fatalf("t=%v: got %v, recomputed %v", seconds, got, want)
		}
	}
}

func recomputedDemand(power *model.Schedule, untilS float64) float64 {
	previous, peak, sum, samples := 100.0, 0.0, 0.0, 0
	for at := 0.0; at+60 <= untilS; at += 60 {
		sum += scheduleValue(power, int64(at*secondsPerNS), 0)
		samples++
		if math.Mod(at+60, 900) == 0 {
			peak, sum, samples = math.Max(peak, sum/float64(samples)), 0, 0
		}
		if math.Mod(at+60, 86400) == 0 {
			previous, peak = peak, 0
		}
	}
	return math.Max(peak, 0.5*previous)
}

func peakWorld(t *testing.T, start int64) *World {
	t.Helper()
	power := &model.Schedule{PeriodS: 86400, Interpolation: model.InterpolateStep,
		Points: []model.SchedulePoint{{AtS: 0, Value: 50}, {AtS: 25200, Value: 120}, {AtS: 25800, Value: 80}}}
	return constWorld(t, func(s *model.DomainSpec) {
		kind := &s.Kinds[0]
		kind.State = append(kind.State, model.State{Name: "power", Initial: 0}, model.State{Name: "demand", Initial: 0}, model.State{Name: "oil", Initial: 0})
		kind.Dynamics = append(kind.Dynamics,
			model.Dynamics{Target: "power", Tier: "F0", F0: &model.F0Dyn{Schedule: power}},
			model.Dynamics{Target: "demand", Tier: "F1", DTMs: 60000, F1: &model.F1Dyn{Form: model.FormPeak,
				Inputs: []model.F1Input{{State: "power", Coef: 1, CoefSet: true}}, BlockS: 900, PeriodS: 86400, Ratchet: 0.5, PreviousPeak: 100}},
			model.Dynamics{Target: "oil", Tier: "F1", DTMs: 1000, F1: &model.F1Dyn{Form: model.FormRegime, Selector: "x", Threshold: 1,
				Inputs: []model.F1Input{{State: "demand", Coef: 0.1, CoefSet: true}},
				Below:  &model.Regime{Gain: 1, TimeConstantS: 300}, Above: &model.Regime{Gain: 2, TimeConstantS: 120}}})
	}, start)
}

func TestDemandBlocksFollowTheClockWhateverTheStart(t *testing.T) {
	t.Parallel()
	start := model.DefaultStartTimeNS + 450e9
	got := peakWorld(t, start).StateValue("e-1", "demand", model.DefaultStartTimeNS+27000e9)
	if want := (10*120.0 + 5*80.0) / 15; math.Abs(got-want) > 1e-9 {
		t.Fatalf("demand = %v, want the 07:00-07:15 clock block mean %v", got, want)
	}
}

func TestScheduleRegimeAndPeakDoNotDependOnReadOrder(t *testing.T) {
	t.Parallel()
	start := model.DefaultStartTimeNS + 37123e6
	names := []string{"power", "demand", "oil", "x"}
	baseline := peakWorld(t, start)
	shuffled := peakWorld(t, start)
	for step := int64(1); step <= 40; step++ {
		at := start + step*997e9
		want := make([]float64, len(names))
		for i, name := range names {
			want[i] = baseline.StateValue("e-1", name, at)
		}
		for i := range names {
			j := (i + int(step)) % len(names)
			if got := shuffled.StateValue("e-1", names[j], at); got != want[j] {
				t.Fatalf("t=%d s %s: got %v, in-order read %v", step*997, names[j], got, want[j])
			}
		}
	}
}
