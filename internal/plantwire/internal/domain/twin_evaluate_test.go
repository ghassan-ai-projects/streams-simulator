package domain

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestAReadingWithinToleranceAgreesWithTheTwin(t *testing.T) {
	t.Parallel()
	report := twEvaluate(t, twInput(twTrace(twTemperature(200, 31)), twApplied("c-1", 100)), newFakeTwin())
	if report.Verdict != VerdictAgrees || report.Summary.Compared != 1 || report.Summary.Within != 1 || report.Summary.Diverged != 0 {
		t.Fatalf("report = %+v, want one reading compared and within", report.Summary)
	}
	if len(report.Divergences) != 0 || report.Metrics[0].Verdict != VerdictAgrees {
		t.Fatalf("divergences = %+v, metric verdict %s, want none and agrees", report.Divergences, report.Metrics[0].Verdict)
	}
	if tol := report.Metrics[0].Tolerance; tol.Source != ToleranceDefaultSpan || tol.Value == nil || *tol.Value != 2.5 {
		t.Fatalf("tolerance = %+v, want the default span of 2.5", tol)
	}
}

func TestAStretchBeyondToleranceIsOneDivergenceWithItsWorstPoint(t *testing.T) {
	t.Parallel()
	input := twInput(twTrace(twTemperature(300, 40), twTemperature(400, 41), twTemperature(500, 36)), twApplied("c-1", 100))
	input.Declared = map[string]float64{"zone-01.temperature": 2}
	report := twEvaluate(t, input, newFakeTwin())
	want := Divergence{Name: "zone-01.temperature#1", Device: "zone-01", Metric: "temperature", FirstNS: 300, LastNS: 500, Readings: 3,
		Tolerance: 2, WorstError: 11, WorstAtNS: 400, BoardValue: 41, TwinValue: 30, Commands: []string{"c-1"}}
	if report.Verdict != VerdictDiverges || report.Summary.Diverged != 3 || report.Summary.Divergences != 1 {
		t.Fatalf("summary = %+v, verdict %s, want three diverged readings in one divergence", report.Summary, report.Verdict)
	}
	if len(report.Divergences) != 1 || !reflect.DeepEqual(report.Divergences[0], want) {
		t.Fatalf("divergences = %+v, want exactly %+v", report.Divergences, want)
	}
}

func TestTwoStretchesSeparatedByAWithinReadingAreTwoDivergences(t *testing.T) {
	t.Parallel()
	input := twInput(twTrace(twTemperature(200, 40), twTemperature(300, 31), twTemperature(400, 42)), twApplied("c-1", 100))
	input.Declared = map[string]float64{"zone-01.temperature": 2}
	report := twEvaluate(t, input, newFakeTwin())
	if len(report.Divergences) != 2 || report.Divergences[0].Name != "zone-01.temperature#1" || report.Divergences[1].Name != "zone-01.temperature#2" {
		t.Fatalf("divergences = %+v, want #1 and #2", report.Divergences)
	}
	if first, second := report.Divergences[0], report.Divergences[1]; first.FirstNS != 200 || first.LastNS != 200 || second.FirstNS != 400 || second.Readings != 1 {
		t.Fatalf("stretches = %+v and %+v, want one reading each at 200 and 400", first, second)
	}
	if report.Summary.Within != 1 || report.Summary.Diverged != 2 || report.Summary.Compared != 3 {
		t.Fatalf("summary = %+v, want 3 compared, 1 within, 2 diverged", report.Summary)
	}
}

func TestANotComparedReadingInsideAStretchDoesNotSplitIt(t *testing.T) {
	t.Parallel()
	suspect := twTemperature(250, 40)
	suspect.Quality = "suspect"
	input := twInput(twTrace(twTemperature(200, 40), suspect, twTemperature(300, 40)), twApplied("c-1", 100))
	input.Declared = map[string]float64{"zone-01.temperature": 2}
	report := twEvaluate(t, input, newFakeTwin())
	if len(report.Divergences) != 1 || report.Divergences[0].Readings != 2 || report.Divergences[0].FirstNS != 200 || report.Divergences[0].LastNS != 300 {
		t.Fatalf("divergences = %+v, want one stretch over 200 to 300 with two readings", report.Divergences)
	}
	if reasons := report.Metrics[0].NotCompared; len(reasons) != 1 || reasons[0] != (ReasonCount{Reason: SkipQuality, Count: 1}) {
		t.Fatalf("not compared = %+v, want one quality_not_good", reasons)
	}
}

func TestTheFirstOfEqualWorstErrorsIsTheWorstPoint(t *testing.T) {
	t.Parallel()
	input := twInput(twTrace(twTemperature(200, 40), twTemperature(300, 40)), twApplied("c-1", 100))
	report := twEvaluate(t, input, newFakeTwin())
	if len(report.Divergences) != 1 || report.Divergences[0].WorstAtNS != 200 || report.Divergences[0].WorstError != 10 {
		t.Fatalf("divergences = %+v, want the worst point at 200 with error 10", report.Divergences)
	}
}

func TestWithoutAnAppliedCommandTheReportCannotJudge(t *testing.T) {
	t.Parallel()
	report := twEvaluate(t, twInput(twTrace(twTemperature(200, 30))), newFakeTwin())
	if report.Verdict != VerdictCannotJudge || report.Reason != ReasonNoCommands {
		t.Fatalf("verdict %s reason %q, want cannot_judge for no_commands_applied", report.Verdict, report.Reason)
	}
	if report.Metrics[0].Verdict != VerdictCannotJudge || report.Summary.MetricsCannotJudge != 1 || report.Summary.Compared != 0 {
		t.Fatalf("metric %+v summary %+v, want the metric to cannot_judge", report.Metrics[0], report.Summary)
	}
}

func TestCommandsAppliedButNoReadingComparedGivesNoReadingsCompared(t *testing.T) {
	t.Parallel()
	report := twEvaluate(t, twInput(twTrace(twTemperature(50, 30)), twApplied("c-1", 100)), newFakeTwin())
	if report.Verdict != VerdictCannotJudge || report.Reason != ReasonNoReadings {
		t.Fatalf("verdict %s reason %q, want cannot_judge for no_readings_compared", report.Verdict, report.Reason)
	}
	if reasons := report.Metrics[0].NotCompared; len(reasons) != 1 || reasons[0].Reason != SkipBeforeCommand {
		t.Fatalf("not compared = %+v, want the reading before the first command skipped", reasons)
	}
}

func TestANotAppliedCommandIsNeverGivenToTheTwin(t *testing.T) {
	t.Parallel()
	refused := TwinCommand{ID: "c-2", Device: "fan-01", Name: "set_pwm_lease", AtNS: 120, Status: CommandNotApplied, Reason: NotAppliedRefusedByBoard}
	twin := newFakeTwin()
	report := twEvaluate(t, twInput(twTrace(twTemperature(200, 30)), twApplied("c-1", 100), refused), twin)
	if twin.position("apply c-2") >= 0 || twin.position("apply c-1") < 0 {
		t.Fatalf("twin log = %v, want only c-1 applied", twin.log)
	}
	if report.Summary.CommandsApplied != 1 || report.Summary.CommandsNotApplied != 1 || report.Commands[1].Status != CommandNotApplied {
		t.Fatalf("summary = %+v commands = %+v, want one applied and one not applied", report.Summary, report.Commands)
	}
}

func TestAnApplyErrorTurnsTheCommandTwinRefusedAndExcludesItFromWhatIsInForce(t *testing.T) {
	t.Parallel()
	twin := newFakeTwin()
	twin.refuse["c-2"] = errors.New(strings.Repeat("r", 300))
	input := twInput(twTrace(twTemperature(200, 40)), twApplied("c-1", 100), twApplied("c-2", 150))
	input.Declared = map[string]float64{"zone-01.temperature": 2}
	report := twEvaluate(t, input, twin)
	if report.Commands[1].Status != CommandTwinRefused || report.Commands[1].Detail != strings.Repeat("r", 200) {
		t.Fatalf("c-2 = %+v, want twin_refused with the detail clipped to 200 bytes", report.Commands[1])
	}
	if len(report.Divergences) != 1 || !reflect.DeepEqual(report.Divergences[0].Commands, []string{"c-1"}) {
		t.Fatalf("divergences = %+v, want only c-1 in force", report.Divergences)
	}
	if report.Summary.CommandsApplied != 1 || report.Summary.CommandsNotApplied != 1 {
		t.Fatalf("summary = %+v, want one applied and one not applied", report.Summary)
	}
}

func TestTheReportIsTheSameWhenTheSameInputIsJudgedTwice(t *testing.T) {
	t.Parallel()
	input := twInput(twTrace(twTemperature(200, 40), twTemperature(300, 31)), twApplied("c-1", 100))
	first, err := json.Marshal(twEvaluate(t, input, newFakeTwin()))
	if err != nil {
		t.Fatal(err)
	}
	second, err := json.Marshal(twEvaluate(t, input, newFakeTwin()))
	if err != nil {
		t.Fatal(err)
	}
	if string(first) != string(second) {
		t.Fatalf("reports differ:\n%s\n%s", first, second)
	}
}

func TestEmptyListsInTheReportMarshalAsArraysNotNull(t *testing.T) {
	t.Parallel()
	raw, err := json.Marshal(twEvaluate(t, twInput(twTrace(), twApplied("c-1", 100)), newFakeTwin()))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"kinds":[]`, `"not_compared_reasons":[]`, `"points":[]`} {
		if !strings.Contains(string(raw), want) {
			t.Errorf("report JSON lacks %s: %s", want, raw)
		}
	}
}

func TestTheDivergencesListMarshalsAsAnArrayWhenNothingDiverges(t *testing.T) {
	t.Parallel()
	raw, err := json.Marshal(twEvaluate(t, twInput(twTrace(twTemperature(200, 30)), twApplied("c-1", 100)), newFakeTwin()))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"divergences":[]`) {
		t.Fatalf("report JSON lacks an empty divergences array: %s", raw)
	}
}

func TestTheReportEchoesItsRuleAndEvidence(t *testing.T) {
	t.Parallel()
	input := twInput(twTrace(twTemperature(200, 30)), twApplied("c-1", 100))
	input.Twin.StartNS = 100
	report := twEvaluate(t, input, newFakeTwin())
	if report.SchemaVersion != TwinReportVersion || report.ToleranceRule.Default != "span_fraction" || report.ToleranceRule.SpanFraction != DefaultSpanFraction {
		t.Fatalf("report header = %s %+v, want the schema version and the span_fraction rule", report.SchemaVersion, report.ToleranceRule)
	}
	if report.Run != input.Run || report.Twin.Domain != "arduino-thermal" || report.Twin.StartNS != 100 {
		t.Fatalf("run %+v twin %+v, want the evidence of the input echoed", report.Run, report.Twin)
	}
}
