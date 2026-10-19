package domain

import (
	"cmp"
	"errors"
	"fmt"
	"maps"
	"math"
	"slices"
)

type TwinModel interface {
	Support(device, metric string) string
	Apply(command TwinCommand) error
	Predict(device, metric string, atNS int64) float64
}

type TwinInput struct {
	Trace        BoardTrace
	Commands     []TwinCommand
	Kinds        KindFiles
	Declared     map[string]float64
	SpanFraction float64
	Run          TwinEvidence
	Twin         TwinInfo
}

type twinStep struct {
	atNS    int64
	command *TwinCommand
	reading *BoardReading
}

type evaluator struct {
	input    TwinInput
	model    TwinModel
	fraction float64
	commands []TwinCommand
	tallies  []*metricTally
	index    map[[2]string]*metricTally
	first    int64
	hasFirst bool
}

func EvaluateTwin(input TwinInput, model TwinModel) (TwinReport, error) {
	e, err := newEvaluator(input, model)
	if err != nil {
		return TwinReport{}, err
	}
	for _, step := range e.timeline() {
		e.run(step)
	}
	return e.report(), nil
}

func newEvaluator(input TwinInput, model TwinModel) (*evaluator, error) {
	fraction, err := SpanFractionOf(input.SpanFraction)
	if err == nil {
		err = errors.Join(CheckTolerances(input.Declared, input.Trace.Metrics), checkTwinSpan(input.Trace))
	}
	if err != nil {
		return nil, err
	}
	e := &evaluator{input: input, model: model, fraction: fraction, commands: slices.Clone(input.Commands), index: map[[2]string]*metricTally{}}
	e.findFirstCommand()
	e.openTallies()
	return e, nil
}

func (e *evaluator) findFirstCommand() {
	for _, command := range e.commands {
		if command.Status == CommandApplied && (!e.hasFirst || command.AtNS < e.first) {
			e.first, e.hasFirst = command.AtNS, true
		}
	}
}

func (e *evaluator) openTallies() {
	for _, metric := range e.input.Trace.Metrics {
		facts := MetricFactsOf(e.input.Kinds[metric.Kind], metric.Metric)
		tolerance := ResolveTolerance(e.input.Declared, e.fraction, metric, facts)
		tally := newTally(metric, facts, tolerance, e.inForce)
		tally.support = e.model.Support(metric.Device, metric.Metric)
		e.tallies = append(e.tallies, tally)
		e.index[[2]string{metric.Device, metric.Metric}] = tally
	}
}

func (e *evaluator) timeline() []twinStep {
	var steps []twinStep
	for i := range e.commands {
		if e.commands[i].Status == CommandApplied {
			steps = append(steps, twinStep{atNS: e.commands[i].AtNS, command: &e.commands[i]})
		}
	}
	for i := range e.input.Trace.Readings {
		steps = append(steps, twinStep{atNS: e.input.Trace.Readings[i].AtNS, reading: &e.input.Trace.Readings[i]})
	}
	slices.SortStableFunc(steps, func(a, b twinStep) int { return cmp.Compare(a.atNS, b.atNS) })
	return steps
}

func (e *evaluator) run(step twinStep) {
	if step.command == nil {
		e.judge(*step.reading)
		return
	}
	if err := e.model.Apply(*step.command); err != nil {
		step.command.Status, step.command.Detail = CommandTwinRefused, clipDetail(err.Error())
	}
}

func clipDetail(text string) string {
	if len(text) > maxDetailBytes {
		return text[:maxDetailBytes]
	}
	return text
}

func (e *evaluator) judge(reading BoardReading) {
	tally := e.index[[2]string{reading.Device, reading.Metric}]
	if tally == nil {
		return
	}
	if reason := e.skipReason(tally, reading); reason != "" {
		tally.skip(reason)
		return
	}
	e.compare(tally, reading, e.model.Predict(reading.Device, reading.Metric, reading.AtNS))
}

func (e *evaluator) skipReason(tally *metricTally, reading BoardReading) string {
	if reason := readingSkipReason(reading); reason != "" {
		return reason
	}
	switch {
	case tally.support != "":
		return tally.support
	case tally.tolerance.Value == nil:
		return SkipNoTolerance
	case !e.hasFirst || reading.AtNS < e.first:
		return SkipBeforeCommand
	}
	return ""
}

func readingSkipReason(reading BoardReading) string {
	switch {
	case reading.Historical:
		return SkipHistorical
	case reading.Quality != "good":
		return SkipQuality
	case !reading.Numeric:
		return SkipNotNumeric
	}
	return ""
}

func (e *evaluator) compare(tally *metricTally, reading BoardReading, twin float64) {
	if math.IsNaN(twin) || math.IsInf(twin, 0) {
		tally.skip(SkipTwinNotFinite)
		return
	}
	tally.point(reading.AtNS, reading.Value, twin)
	if gap := math.Abs(reading.Value - twin); gap > *tally.tolerance.Value {
		tally.differ(reading.AtNS, reading.Value, twin, gap)
		return
	}
	tally.agree()
}

func (e *evaluator) inForce(atNS int64) []string {
	latest := map[[2]string]TwinCommand{}
	for _, command := range e.commands {
		key := [2]string{command.Device, command.Name}
		if kept, found := latest[key]; command.Status == CommandApplied && command.AtNS <= atNS && (!found || command.AtNS >= kept.AtNS) {
			latest[key] = command
		}
	}
	ids := []string{}
	for _, command := range slices.SortedFunc(maps.Values(latest), compareCommands) {
		ids = append(ids, command.ID)
	}
	return ids
}

func checkTwinSpan(trace BoardTrace) error {
	last := trace.StartNS
	for _, reading := range trace.Readings {
		last = max(last, reading.AtNS)
	}
	if last-trace.StartNS > MaxTwinSpanNS {
		return fmt.Errorf("%w: the readings run from %d to %d ns, over the %d ns allowed", ErrTwinSpan, trace.StartNS, last, MaxTwinSpanNS)
	}
	return nil
}
