package domain

import (
	"cmp"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"
)

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
	return Summary{ID: e.ID, Version: e.Version, Kind: KindEquipment, Title: e.Title, Class: e.Class, Tags: orEmpty(e.Tags), Digest: e.Digest, File: e.File}
}

func summarizeInstrument(i *Instrument) Summary {
	return Summary{ID: i.ID, Version: i.Version, Kind: KindInstrument, Title: i.Title, Class: i.UnitFamily, Tags: []string{i.Tag}, Digest: i.Digest, File: i.File}
}

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
