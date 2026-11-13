package domain

import (
	"math"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

func sNS(seconds float64) int64 { return model.DefaultStartTimeNS + int64(seconds*1e9) }

func predictionOn(entity string, at, failure float64) model.Detection {
	return model.Detection{EntityID: entity, DetectedAt: model.FormatTime(sNS(at)), Label: "bearing", PredictedFailureAt: model.FormatTime(sNS(failure))}
}

// prognosisRun: D-3 first trips at 10000 s when nobody acts (overtemperature; its relay at 10100 s); a refused call at
// 5000 s and the first call that took effect at 9000 s, so predictions are graded before 9000 s.
func prognosisRun() (Evidence, *model.GroundTruthRecord) {
	nothing := &model.OutcomeFacts{Entities: []model.EntityOutcome{
		{EntityID: "D-1", Trips: []model.TripOutcome{{State: "relay_ok", FirstNS: sNS(9500), Count: 1}}},
		{EntityID: "D-3", Trips: []model.TripOutcome{{State: "relay_ok", FirstNS: sNS(10100), Count: 1}, {State: "overtemp_ok", FirstNS: sNS(10000), Count: 1}}}}}
	verdict := &model.Verdict{Detections: []model.Detection{
		predictionOn("D-3", 6000, 10300),
		predictionOn("D-3", 500, 10000),
		predictionOn("D-3", 2000, 11000),
		predictionOn("D-1", 3000, 9000),
		{EntityID: "D-3", DetectedAt: model.FormatTime(sNS(3000)), Label: "bearing"},
		predictionOn("D-3", 4000, 9400),
		predictionOn("D-3", 9500, 20000),
	}}
	calls := []world.EffectorCall{{Effector: "set_feed", EntityID: "F-1", AtNS: sNS(5000)}, {Effector: "set_feed", EntityID: "F-1", AtNS: sNS(9000), Accepted: true, EffectApplied: true}}
	label := &model.GroundTruthRecord{EntityID: "D-3", Label: "bearing", FirstObservableTimeNS: sNS(1000)}
	return Evidence{RunID: "r", Verdict: verdict, Calls: calls, DoNothing: nothing}, label
}

// By hand: graded are the predictions at 2000 s (11000 s: late by 1000 s, weighted 2000 / RUL* 8000 = 0.25, outside
// ±0.2), 4000 s (9400 s: early by 600 s, 600 / 6000 = 0.1, inside) and 6000 s (10300 s: late by 300, 600 / 4000 = 0.15,
// inside); 500 s precedes first_observable, 9500 s follows the first effective call. PH = 10000 − 4000 = 6000 s. λ 0.5:
// 2000 + 0.5·8000 = 6000 s, the 6000 s prediction in force, 600 / 4000 = 0.15: inside, RA 0.85. λ 0.75: 8000 s, the same
// prediction, 600 / 2000 = 0.3: outside, RA 0.7.
func TestPrognosticMetricsAgreeWithAHandComputedCase(t *testing.T) {
	t.Parallel()
	ev, label := prognosisRun()
	got := judgment(ev, label).Prognosis
	if got == nil {
		t.Fatal("no prognosis graded")
	}
	if got.FailureNS != sNS(10000) || got.GradedUntilNS != sNS(9000) || got.Predictions != 3 || got.HorizonNS != 6000e9 ||
		got.Alpha != 0.2 || got.Beta != 0.5 || got.LateWeight != 2 {
		t.Fatalf("prognosis = %+v, want failure 10000 s, graded until 9000 s, 3 predictions, PH 6000 s", *got)
	}
	want := []struct {
		lambda, at, accuracy float64
		inside               bool
	}{{0.5, 6000, 0.85, true}, {0.75, 8000, 0.7, false}}
	for i, w := range want {
		m := got.Lambdas[i]
		if m.Lambda != w.lambda || m.AtNS != sNS(w.at) || !m.Graded || m.AlphaLambda != w.inside || m.RelativeAccuracy == nil || math.Abs(*m.RelativeAccuracy-w.accuracy) > 1e-12 {
			t.Errorf("λ %.2f: %+v (accuracy %v), want at %.0f s, α-λ %v, RA %.2f", w.lambda, m, m.RelativeAccuracy, w.at, w.inside, w.accuracy)
		}
	}
}

func TestALambdaPointAfterTheFirstEffectiveCallIsNotGraded(t *testing.T) {
	t.Parallel()
	ev, label := prognosisRun()
	ev.Calls[1].AtNS = sNS(5500)
	got := judgment(ev, label).Prognosis
	if got.GradedUntilNS != sNS(5500) || got.Lambdas[0].Graded || got.Lambdas[1].Graded || got.Lambdas[0].RelativeAccuracy != nil {
		t.Fatalf("prognosis = %+v, want both λ points (6000 s, 8000 s) ungraded behind the first effective call at 5500 s", *got)
	}
	if got.Predictions != 2 || got.HorizonNS != 6000e9 {
		t.Fatalf("prognosis = %+v, want the two predictions before 5500 s and PH 6000 s", *got)
	}
}

func TestNoPrognosisWithoutAPredictionOrAFailureToGradeItAgainst(t *testing.T) {
	t.Parallel()
	for name, change := range map[string]func(*Evidence, *model.GroundTruthRecord){
		"no do-nothing replay":              func(ev *Evidence, _ *model.GroundTruthRecord) { ev.DoNothing = nil },
		"the entity never fails":            func(_ *Evidence, label *model.GroundTruthRecord) { label.EntityID = "D-2" },
		"no prediction":                     func(ev *Evidence, _ *model.GroundTruthRecord) { ev.Verdict.Detections = ev.Verdict.Detections[4:5] },
		"every prediction after the action": func(ev *Evidence, _ *model.GroundTruthRecord) { ev.Calls[1].AtNS = sNS(1500) },
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ev, label := prognosisRun()
			change(&ev, label)
			if got := judgment(ev, label).Prognosis; got != nil {
				t.Fatalf("prognosis = %+v, want none", *got)
			}
		})
	}
}

func TestAMultiLabelRunGradesPredictionsWithTheWholeRunsCalls(t *testing.T) {
	t.Parallel()
	ev, label := prognosisRun()
	other := &model.GroundTruthRecord{EntityID: "D-1", Label: "wear"}
	graded, err := ScoreRun(ev, []*model.GroundTruthRecord{label, other})
	if err != nil {
		t.Fatal(err)
	}
	if got := graded.Scorecards[0].Judgment.Prognosis; got == nil || got.GradedUntilNS != sNS(9000) || got.Predictions != 3 {
		t.Fatalf("D-3 prognosis = %+v, want the feeder's call at 9000 s to end the window although D-3's label does not expect it", got)
	}
	if got := graded.Scorecards[1].Judgment.Prognosis; got == nil || got.FailureNS != sNS(9500) || got.Predictions != 1 {
		t.Fatalf("D-1 prognosis = %+v, want its one prediction graded against its trip at 9500 s", got)
	}
}
