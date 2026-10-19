package domain

import (
	"cmp"
	"fmt"
	"maps"
	"math"
	"slices"
	"strconv"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

const toleranceRuleText = "A metric whose device or kind has no declared tolerance is judged within span_fraction times the span " +
	"(max minus min) of its eu_range in the kind file; a metric with neither is not judged."

type EURange struct{ Min, Max float64 }

type MetricFacts struct {
	Unit  string
	Range *EURange
}

func MetricFactsOf(kindFile map[string]any, metric string) MetricFacts {
	for _, entry := range listOf(kindFile["metrics"]) {
		fields, _ := entry.(map[string]any)
		if text(fields, "name") == metric {
			return MetricFacts{Unit: text(fields, "unit"), Range: rangeOf(fields["eu_range"])}
		}
	}
	return MetricFacts{}
}

func rangeOf(value any) *EURange {
	bounds, _ := value.(map[string]any)
	low, lowKnown := numberOf(bounds["min"])
	high, highKnown := numberOf(bounds["max"])
	if !lowKnown || !highKnown || high <= low {
		return nil
	}
	return &EURange{Min: low, Max: high}
}

func SpanFractionOf(fraction float64) (float64, error) {
	switch {
	case fraction == 0:
		return DefaultSpanFraction, nil
	case math.IsNaN(fraction) || fraction < 0 || fraction > 1:
		return 0, fmt.Errorf("%w: the span fraction %v is outside (0, 1]", ErrTwinTolerance, fraction)
	}
	return fraction, nil
}

func ResolveTolerance(declared map[string]float64, fraction float64, metric BoardMetric, facts MetricFacts) Tolerance {
	for _, key := range []string{metric.Device + "." + metric.Metric, metric.Kind + "." + metric.Metric} {
		if value, found := declared[key]; found {
			return Tolerance{Source: ToleranceDeclared, Value: &value}
		}
	}
	if facts.Range == nil {
		return Tolerance{Source: ToleranceNone}
	}
	value := fraction * (facts.Range.Max - facts.Range.Min)
	return Tolerance{Source: ToleranceDefaultSpan, Value: &value}
}

func ParseTolerances(items []string) (map[string]float64, error) {
	declared := map[string]float64{}
	for _, item := range items {
		for _, entry := range strings.Split(item, ",") {
			if err := declare(declared, strings.TrimSpace(entry)); err != nil {
				return nil, err
			}
		}
	}
	return declared, nil
}

func declare(declared map[string]float64, entry string) error {
	at := strings.LastIndex(entry, "=")
	if at <= 0 || !strings.Contains(entry[:at], ".") {
		return fmt.Errorf("%w: %q is not selector.metric=value", ErrTwinTolerance, entry)
	}
	value, err := toleranceValue(entry, entry[at+1:])
	key := strings.TrimSpace(entry[:at])
	if _, twice := declared[key]; twice && err == nil {
		err = fmt.Errorf("%w: %q is declared twice", ErrTwinTolerance, key)
	}
	declared[key] = value
	return err
}

func toleranceValue(entry, raw string) (float64, error) {
	value, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil || math.IsNaN(value) || math.IsInf(value, 0) || value < 0 {
		return 0, fmt.Errorf("%w: %q needs a finite value of at least 0", ErrTwinTolerance, entry)
	}
	return value, nil
}

func CheckTolerances(declared map[string]float64, metrics []BoardMetric) error {
	if unknown := unknownTolerances(declared, metrics); len(unknown) > 0 {
		return fmt.Errorf("%w: %s names no device or kind metric of the board", ErrTwinTolerance, strings.Join(unknown, ", "))
	}
	return nil
}

func unknownTolerances(declared map[string]float64, metrics []BoardMetric) []string {
	known := map[string]bool{}
	for _, metric := range metrics {
		known[metric.Device+"."+metric.Metric], known[metric.Kind+"."+metric.Metric] = true, true
	}
	var unknown []string
	for _, key := range slices.Sorted(maps.Keys(declared)) {
		if !known[key] {
			unknown = append(unknown, key)
		}
	}
	return unknown
}

func TwinImplements(spec *model.DomainSpec, devices []BoardDevice) error {
	implemented := map[string]bool{}
	for i := range spec.Kinds {
		implemented[spec.Kinds[i].Implements] = true
	}
	var wanted []string
	for _, device := range devices {
		if implemented[device.Kind] {
			return nil
		}
		wanted = append(wanted, device.Kind)
	}
	slices.Sort(wanted)
	return fmt.Errorf("%w: domain %s, board kinds %s", ErrTwinKinds, spec.ID, strings.Join(slices.Compact(wanted), ", "))
}

func compareCommands(a, b TwinCommand) int {
	return cmp.Or(cmp.Compare(a.AtNS, b.AtNS), cmp.Compare(a.ID, b.ID))
}
