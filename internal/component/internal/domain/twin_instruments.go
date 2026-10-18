package domain

import (
	"errors"
	"fmt"
)

var errConversionNeeded = errors.New("component: the instrument's unit needs a conversion")

type conversionNeeded struct{ message string }

func (c conversionNeeded) Error() string { return c.message }

func (c conversionNeeded) Is(target error) bool { return target == errConversionNeeded }

func (t *twinBuild) attachInstrument(plan *metricPlan, at string) error {
	if plan.choice.Instrument == instrumentNone {
		return nil
	}
	instrument, err := t.lib.instrumentFor(plan.choice.Instrument, plan.file)
	if err != nil {
		return t.refuseRule(ErrInstrumentMismatch, instrumentRule(err), at, "metric %q: %v", plan.file.Name, err)
	}
	plan.instrument = instrument
	return nil
}

func instrumentRule(err error) twinRule {
	if errors.Is(err, errConversionNeeded) {
		return twinRule{KeywordConversionNeeded, "unit"}
	}
	return ruleOf(ErrInstrumentMismatch)
}

func (l *Library) instrumentFor(ref string, metric kindMetric) (*Instrument, error) {
	key, err := l.resolve(ref)
	if err != nil {
		return nil, err
	}
	instrument, found := l.instruments[key]
	if !found {
		return nil, fmt.Errorf("%s is not an instrument of the library", ref)
	}
	if err := l.fit(instrument, metric); err != nil {
		return nil, err
	}
	return instrument, nil
}

func (l *Library) fit(instrument *Instrument, metric kindMetric) error {
	switch {
	case metric.DataType == "boolean":
		return fmt.Errorf("%s measures a number, and the metric is boolean", instrument.ID)
	case metric.EURange == nil:
		return fmt.Errorf("%s needs the metric to state its eu_range", instrument.ID)
	case instrument.Range != nil && (metric.EURange.Min < instrument.Range.Min || metric.EURange.Max > instrument.Range.Max):
		return fmt.Errorf("%s reads %v to %v, narrower than the eu_range %v to %v of the metric",
			instrument.ID, instrument.Range.Min, instrument.Range.Max, metric.EURange.Min, metric.EURange.Max)
	}
	return l.sameUnit(instrument, metric)
}

func (l *Library) sameUnit(instrument *Instrument, metric kindMetric) error {
	conv, err := l.vocabulary.convert(instrument.UnitFamily, metric.Unit, instrument.Unit)
	if err != nil {
		return fmt.Errorf("%s measures %s in %s: %w", instrument.ID, instrument.UnitFamily, instrument.Unit, err)
	}
	if !conv.identity() {
		return conversionNeeded{fmt.Sprintf("%s reads in %s, which needs a conversion from the metric's %s: a twin keeps the unit of its kind file",
			instrument.ID, instrument.Unit, metric.Unit)}
	}
	return nil
}
