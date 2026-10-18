package domain

import (
	"cmp"
	"errors"
	"fmt"
	"maps"
	"math"
	"slices"
	"strconv"
	"strings"
)

var ErrReadingRange = errors.New("component: the reading range is not usable")

func (l *Library) List() []Summary {
	summaries := make([]Summary, 0, len(l.equipment)+len(l.instruments))
	for _, key := range slices.Sorted(maps.Keys(l.equipment)) {
		summaries = append(summaries, summarizeEquipment(l.equipment[key]))
	}
	for _, key := range slices.Sorted(maps.Keys(l.instruments)) {
		summaries = append(summaries, summarizeInstrument(l.instruments[key]))
	}
	slices.SortFunc(summaries, func(a, b Summary) int {
		return cmp.Or(strings.Compare(a.ID, b.ID), cmp.Compare(a.Version, b.Version))
	})
	return summaries
}

func summarizeEquipment(e *Equipment) Summary {
	return Summary{ID: e.ID, Version: e.Version, Kind: KindEquipment, Title: e.Title, Class: e.Class, Tags: orEmpty(e.Tags), Digest: e.Digest}
}

func summarizeInstrument(i *Instrument) Summary {
	return Summary{ID: i.ID, Version: i.Version, Kind: KindInstrument, Title: i.Title, Class: i.UnitFamily, Tags: []string{i.Tag}, Digest: i.Digest,
		Unit: i.Unit, Range: i.Range}
}

func (l *Library) ListForReading(unit string, low, high *float64) ([]Summary, error) {
	eu, err := readingRange(unit, low, high)
	if err != nil {
		return nil, err
	}
	if unit == "" {
		return l.List(), nil
	}
	return l.fitted(l.InstrumentsForUnit(unit), kindMetric{DataType: "number", Unit: unit, EURange: eu}), nil
}

func (l *Library) fitted(found []Summary, metric kindMetric) []Summary {
	for i := range found {
		verdict := l.fit(l.instruments[reference(found[i].ID, found[i].Version)], metric).Fit
		found[i].Fit = &verdict
	}
	slices.SortStableFunc(found, func(a, b Summary) int { return cmp.Compare(fitRank(a), fitRank(b)) })
	return found
}

func fitRank(summary Summary) int {
	if summary.Fit.OK {
		return 0
	}
	return 1
}

func readingRange(unit string, low, high *float64) (*Range, error) {
	switch {
	case low == nil && high == nil:
		return nil, nil
	case unit == "" || low == nil || high == nil:
		return nil, fmt.Errorf("%w: give both a minimum and a maximum, with the unit they are in", ErrReadingRange)
	case notFinite(*low) || notFinite(*high) || *low >= *high:
		return nil, fmt.Errorf("%w: the minimum %v must be a number below the maximum %v", ErrReadingRange, *low, *high)
	}
	return &Range{Min: *low, Max: *high}, nil
}

func (l *Library) InstrumentsForUnit(unit string) []Summary {
	found := []Summary{}
	for _, key := range slices.Sorted(maps.Keys(l.instruments)) {
		if match := l.unitMatch(l.instruments[key], unit); match != "" {
			summary := summarizeInstrument(l.instruments[key])
			summary.Match = match
			found = append(found, summary)
		}
	}
	slices.SortStableFunc(found, func(a, b Summary) int {
		return cmp.Or(cmp.Compare(rankOf(a.Match), rankOf(b.Match)), strings.Compare(a.ID, b.ID), cmp.Compare(a.Version, b.Version))
	})
	return found
}

func (l *Library) unitMatch(instrument *Instrument, unit string) string {
	if _, err := l.vocabulary.convert(instrument.UnitFamily, unit, instrument.Unit); err != nil {
		return ""
	}
	if instrument.Unit == unit {
		return MatchExact
	}
	return MatchConvertible
}

func rankOf(match string) int {
	if match == MatchExact {
		return 0
	}
	return 1
}

func notFinite(number float64) bool { return math.IsNaN(number) || math.IsInf(number, 0) }

func orEmpty(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}

func (l *Library) Describe(ref string) (*Detail, error) {
	key, err := l.resolve(ref)
	if err != nil {
		return nil, err
	}
	if equipment, found := l.equipment[key]; found {
		return &Detail{Entry: summarizeEquipment(equipment), Document: equipment.Document}, nil
	}
	instrument := l.instruments[key]
	return &Detail{Entry: summarizeInstrument(instrument), Document: instrument.Document}, nil
}

func (l *Library) resolve(ref string) (string, error) {
	if strings.Contains(ref, "@") {
		return l.exact(ref)
	}
	var highest int
	for _, summary := range l.List() {
		if summary.ID == ref {
			highest = max(highest, summary.Version)
		}
	}
	if highest == 0 {
		return "", fmt.Errorf("component: %q is not in the library", ref)
	}
	return reference(ref, highest), nil
}

func (l *Library) exact(ref string) (string, error) {
	id, version, _ := strings.Cut(ref, "@")
	major, err := strconv.Atoi(version)
	if err != nil || major < 1 {
		return "", fmt.Errorf("component: %q is not <namespace>/<name>@<major>", ref)
	}
	_, equipment := l.equipment[ref]
	_, instrument := l.instruments[ref]
	if !equipment && !instrument {
		return "", fmt.Errorf("component: %s is not in the library", reference(id, major))
	}
	return ref, nil
}

func (l *Library) Counts() (equipment, instruments int) {
	return len(l.equipment), len(l.instruments)
}
