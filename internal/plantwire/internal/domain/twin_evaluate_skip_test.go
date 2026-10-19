package domain

import (
	"errors"
	"math"
	"reflect"
	"testing"
)

func TestEachReasonForNotComparingAReadingIsRecordedOnItsOwn(t *testing.T) {
	t.Parallel()
	historical := twTemperature(200, 30)
	historical.Historical = true
	suspect := twTemperature(200, 30)
	suspect.Quality = "suspect"
	text := twTemperature(200, 30)
	text.Numeric = false
	tests := []struct {
		name    string
		reading BoardReading
		support map[string]string
		ranged  bool
		twin    float64
		reason  string
	}{
		{"a historical batch", historical, nil, true, 30, SkipHistorical},
		{"a reading that is not good", suspect, nil, true, 30, SkipQuality},
		{"a value that is not a number", text, nil, true, 30, SkipNotNumeric},
		{"a device the twin has no entity for", twTemperature(200, 30), map[string]string{"zone-01.temperature": SkipNoEntity}, true, 30, SkipNoEntity},
		{"a metric the twin has no channel for", twTemperature(200, 30), map[string]string{"zone-01.temperature": SkipNoChannel}, true, 30, SkipNoChannel},
		{"a metric with no tolerance declared", twTemperature(200, 30), nil, false, 30, SkipNoTolerance},
		{"a reading before the first applied command", twTemperature(99, 30), nil, true, 30, SkipBeforeCommand},
		{"a twin that predicts not a number", twTemperature(200, 30), nil, true, math.NaN(), SkipTwinNotFinite},
		{"a twin that predicts infinity", twTemperature(200, 30), nil, true, math.Inf(1), SkipTwinNotFinite},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			input := twInput(twTrace(tt.reading), twApplied("c-1", 100))
			input.Kinds = twKinds(tt.ranged)
			twin := newFakeTwin()
			twin.support = tt.support
			twin.predict["zone-01.temperature"] = tt.twin
			report := twEvaluate(t, input, twin)
			if want := []ReasonCount{{Reason: tt.reason, Count: 1}}; !reflect.DeepEqual(report.Metrics[0].NotCompared, want) {
				t.Fatalf("not compared = %+v, want %+v", report.Metrics[0].NotCompared, want)
			}
			if report.Verdict != VerdictCannotJudge || report.Reason != ReasonNoReadings || report.Summary.Compared != 0 {
				t.Fatalf("verdict %s reason %q compared %d, want no reading compared", report.Verdict, report.Reason, report.Summary.Compared)
			}
		})
	}
}

func TestAReadingAtTheCommandsOwnTimeIsCompared(t *testing.T) {
	t.Parallel()
	report := twEvaluate(t, twInput(twTrace(twTemperature(100, 30)), twApplied("c-1", 100)), newFakeTwin())
	if report.Summary.Compared != 1 || report.Summary.NotCompared != 0 {
		t.Fatalf("summary = %+v, want the reading at the command's time compared", report.Summary)
	}
}

func TestTheCommandAtTheSameTimeAsAReadingIsAppliedFirst(t *testing.T) {
	t.Parallel()
	twin := newFakeTwin()
	twEvaluate(t, twInput(twTrace(twTemperature(200, 30)), twApplied("c-1", 200)), twin)
	applied, predicted := twin.position("apply c-1"), twin.position("predict zone-01.temperature@200")
	if applied < 0 || predicted < 0 || applied > predicted {
		t.Fatalf("twin log = %v, want the command applied before the reading is predicted", twin.log)
	}
}

func TestReadingsAreJudgedInTimeOrderWhateverTheTraceOrder(t *testing.T) {
	t.Parallel()
	twin := newFakeTwin()
	twEvaluate(t, twInput(twTrace(twTemperature(300, 30), twTemperature(200, 30)), twApplied("c-1", 50)), twin)
	early, late := twin.position("predict zone-01.temperature@200"), twin.position("predict zone-01.temperature@300")
	if early < 0 || late < 0 || early > late {
		t.Fatalf("twin log = %v, want the reading at 200 predicted before the one at 300", twin.log)
	}
}

func TestASeriesKeepsItsFirstTwoThousandPointsAndCountsTheRest(t *testing.T) {
	t.Parallel()
	report := twEvaluate(t, twInput(twTrace(twReadings(MaxSeriesPoints+1, 30)...), twApplied("c-1", 100)), newFakeTwin())
	series := report.Metrics[0].Series
	if len(series.Points) != MaxSeriesPoints || series.Total != MaxSeriesPoints+1 || !series.Truncated {
		t.Fatalf("series = %d points, total %d, truncated %v; want %d points of %d, truncated", len(series.Points), series.Total, series.Truncated, MaxSeriesPoints, MaxSeriesPoints+1)
	}
	if series.Points[0].AtNS != 200 || series.Points[MaxSeriesPoints-1].AtNS != 200+MaxSeriesPoints-1 {
		t.Fatalf("first and last kept points at %d and %d, want the first two thousand in time order", series.Points[0].AtNS, series.Points[MaxSeriesPoints-1].AtNS)
	}
}

func TestASeriesWithinItsLimitIsNotTruncated(t *testing.T) {
	t.Parallel()
	report := twEvaluate(t, twInput(twTrace(twReadings(3, 30)...), twApplied("c-1", 100)), newFakeTwin())
	if series := report.Metrics[0].Series; len(series.Points) != 3 || series.Total != 3 || series.Truncated {
		t.Fatalf("series = %+v, want three points, none truncated", series)
	}
}

func TestACommandToAKindWithADeclaredToleranceUsesIt(t *testing.T) {
	t.Parallel()
	input := twInput(twTrace(twTemperature(200, 32)), twApplied("c-1", 100))
	input.Declared = map[string]float64{"thermal_sensor.temperature": 1}
	report := twEvaluate(t, input, newFakeTwin())
	if tol := report.Metrics[0].Tolerance; tol.Source != ToleranceDeclared || *tol.Value != 1 || report.Summary.Diverged != 1 {
		t.Fatalf("tolerance %+v diverged %d, want the kind's declared 1 and a divergence of 2", tol, report.Summary.Diverged)
	}
}

func TestAnUndeclaredToleranceOrASpanFractionOutsideTheUnitIsRefused(t *testing.T) {
	t.Parallel()
	undeclared := twInput(twTrace(twTemperature(200, 30)), twApplied("c-1", 100))
	undeclared.Declared = map[string]float64{"zone-09.temperature": 1}
	badSpan := twInput(twTrace(twTemperature(200, 30)), twApplied("c-1", 100))
	badSpan.SpanFraction = 2
	for name, input := range map[string]TwinInput{"an undeclared key": undeclared, "a span fraction of 2": badSpan} {
		if _, err := EvaluateTwin(input, newFakeTwin()); !errors.Is(err, ErrTwinTolerance) {
			t.Errorf("%s: EvaluateTwin = %v, want ErrTwinTolerance", name, err)
		}
	}
}

func TestATraceSpanningMoreThanAWeekIsRefusedBeforeTheTwinRuns(t *testing.T) {
	t.Parallel()
	week := MaxTwinSpanNS
	input := twInput(twTrace(twTemperature(0, 30), twTemperature(week+1, 30)), twApplied("c-1", 0))
	twin := newFakeTwin()
	if _, err := EvaluateTwin(input, twin); !errors.Is(err, ErrTwinSpan) || len(twin.log) != 0 {
		t.Fatalf("EvaluateTwin = %v, twin log %v, want ErrTwinSpan before anything is applied or predicted", err, twin.log)
	}
	exact := twInput(twTrace(twTemperature(0, 30), twTemperature(week, 30)), twApplied("c-1", 0))
	if _, err := EvaluateTwin(exact, newFakeTwin()); err != nil {
		t.Fatalf("EvaluateTwin of exactly a week = %v, want it accepted", err)
	}
}
