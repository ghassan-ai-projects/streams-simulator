package domain

import (
	"errors"
	"fmt"
)

const (
	FitUnitIncompatible = "unit_incompatible"
	FitRangeUnstated    = "range_unstated"
	FitRangeNarrower    = "range_narrower"
	FitReadingNotNumber = "reading_not_number"

	rangeSlack = 1e-9
)

type Fit struct {
	OK     bool   `json:"ok"`
	Code   string `json:"code,omitempty"`
	Member string `json:"member,omitempty"`
}

type fitResult struct {
	Fit
	conv    conversion
	message string
}

type fitRefusal struct {
	fit     Fit
	message string
}

func (r fitRefusal) Error() string { return r.message }

func (l *Library) fit(instrument *Instrument, metric kindMetric) fitResult {
	if metric.DataType == "boolean" {
		return refusedFit(FitReadingNotNumber, "data_type", "%s measures a number, and the metric is boolean", instrument.ID)
	}
	conv, err := l.vocabulary.convert(instrument.UnitFamily, metric.Unit, instrument.Unit)
	switch {
	case err != nil:
		return refusedFit(FitUnitIncompatible, "unit", "%s measures %s in %s: %v", instrument.ID, instrument.UnitFamily, instrument.Unit, err)
	case metric.EURange == nil:
		return refusedFit(FitRangeUnstated, "eu_range", "%s needs the metric to state its eu_range", instrument.ID)
	}
	return rangeFit(instrument, *metric.EURange, conv)
}

func rangeFit(instrument *Instrument, eu Range, conv conversion) fitResult {
	reach := inInstrumentUnit(conv, eu)
	if instrument.Range != nil && !covers(*instrument.Range, reach) {
		return refusedFit(FitRangeNarrower, "eu_range", "%s reads %v to %v %s, narrower than the eu_range %v to %v of the metric, which is %v to %v in that unit",
			instrument.ID, instrument.Range.Min, instrument.Range.Max, instrument.Unit, eu.Min, eu.Max, reach.Min, reach.Max)
	}
	return fitResult{Fit: Fit{OK: true}, conv: conv}
}

func refusedFit(code, member, format string, args ...any) fitResult {
	message := fmt.Sprintf(format, args...)
	return fitResult{Fit: Fit{Code: code, Member: member}, message: message}
}

func inInstrumentUnit(conv conversion, eu Range) Range {
	low, high := conv.apply(eu.Min), conv.apply(eu.Max)
	return Range{Min: min(low, high), Max: max(low, high)}
}

func covers(instrument, wanted Range) bool {
	slack := rangeSlack * (instrument.Max - instrument.Min)
	return wanted.Min >= instrument.Min-slack && wanted.Max <= instrument.Max+slack
}

func (r fitResult) err() error {
	if r.OK {
		return nil
	}
	return fitRefusal{fit: r.Fit, message: r.message}
}

func (t *twinBuild) attachInstrument(plan *metricPlan, at string) error {
	if plan.choice.Instrument == instrumentNone {
		return nil
	}
	instrument, conv, err := t.lib.instrumentFor(plan.choice.Instrument, plan.file)
	if err != nil {
		return t.refuseRule(ErrInstrumentMismatch, instrumentRule(err), at, "metric %q: %v", plan.file.Name, err)
	}
	plan.instrument, plan.conv = instrument, conv
	return nil
}

func instrumentRule(err error) twinRule {
	rule := ruleOf(ErrInstrumentMismatch)
	var refusal fitRefusal
	if errors.As(err, &refusal) {
		rule.member = refusal.fit.Member
	}
	return rule
}

func (l *Library) instrumentFor(ref string, metric kindMetric) (*Instrument, conversion, error) {
	key, err := l.resolve(ref)
	if err != nil {
		return nil, conversion{}, err
	}
	instrument, found := l.instruments[key]
	if !found {
		return nil, conversion{}, fmt.Errorf("%s is not an instrument of the library", ref)
	}
	verdict := l.fit(instrument, metric)
	return instrument, verdict.conv, verdict.err()
}
