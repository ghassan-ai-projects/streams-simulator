package domain

import (
	"fmt"
	"sort"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func catalogEntry(spec *model.DomainSpec) Entry {
	return Entry{ID: spec.ID, Title: spec.Title, Version: spec.Version, Axes: axisVector(spec), Stresses: spec.Stresses}
}

func coverageCounts() map[string]map[string]int {
	by := map[string]map[string]int{}
	for _, axis := range []string{"rate", "cardinality", "value_shape", "cadence", "lateness", "absence", "time_reference", "correlation", "seasonality", "actuation", "consequence", "fidelity"} {
		by[axis] = map[string]int{}
	}
	return by
}

func addDomainCoverage(by map[string]map[string]int, s *model.DomainSpec) {
	addCoverage(by["rate"], []string{s.Axes.Rate})
	addCoverage(by["cardinality"], []string{s.Axes.Cardinality})
	addCoverage(by["value_shape"], s.Axes.ValueShape)
	addCoverage(by["cadence"], s.Axes.Cadence)
	addCoverage(by["lateness"], []string{s.Axes.Lateness})
	addCoverage(by["absence"], []string{s.Axes.Absence})
	addCoverage(by["time_reference"], []string{s.Axes.TimeRef})
	addCoverage(by["correlation"], s.Axes.Correlation)
	addCoverage(by["seasonality"], s.Axes.Seasonality)
	addCoverage(by["actuation"], []string{s.Axes.Actuation})
	addCoverage(by["consequence"], s.Axes.Consequence)
	addCoverage(by["fidelity"], s.Axes.Fidelity)
}

func addCoverage(counts map[string]int, values []string) {
	for _, value := range values {
		if value != "" {
			counts[value]++
		}
	}
}

func thinCoverage(by map[string]map[string]int) []string {
	var thin []string
	for axis, counts := range by {
		for value, n := range counts {
			if n < 2 {
				thin = append(thin, fmt.Sprintf("%s=%s(%d)", axis, value, n))
			}
		}
	}
	sort.Strings(thin)
	return thin
}
