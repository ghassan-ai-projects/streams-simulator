package domain

import (
	"maps"
	"math"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

const (
	PresenceBoth  = "both"
	PresenceOnlyA = "only_in_a"
	PresenceOnlyB = "only_in_b"
)

type Results struct {
	KPIs        []KPIRow   `json:"kpis"`
	Alarms      []AlarmRow `json:"alarms"`
	AlarmSystem SystemRow  `json:"alarm_system"`
}

type Scope struct {
	Kind               string `json:"kind"`
	ASimulatedLengthNS *int64 `json:"a_simulated_length_ns"`
	BSimulatedLengthNS *int64 `json:"b_simulated_length_ns"`
}

type KPIRow struct {
	EntityID      string       `json:"entity_id"`
	KPI           string       `json:"kpi"`
	Presence      string       `json:"presence"`
	A             *KPIStats    `json:"a"`
	B             *KPIStats    `json:"b"`
	Delta         *KPIDelta    `json:"delta,omitempty"`
	RelativeDelta *KPIRelative `json:"relative_delta,omitempty"`
	Scope         Scope        `json:"scope"`
}

type KPIStats struct {
	Mean    float64 `json:"mean"`
	Min     float64 `json:"min"`
	Max     float64 `json:"max"`
	Samples int     `json:"samples"`
}

type KPIDelta struct {
	Mean float64 `json:"mean"`
	Min  float64 `json:"min"`
	Max  float64 `json:"max"`
}

type KPIRelative struct {
	Mean *float64 `json:"mean"`
	Min  *float64 `json:"min"`
	Max  *float64 `json:"max"`
}

type entry[V any] struct {
	entity, state string
	a, b          *V
}

func resultsOf(a, b *Report) Results {
	scope := scopeOf(kpisOf(a), kpisOf(b))
	return Results{KPIs: kpiComparisons(a, b, scope), Alarms: alarmRows(a, b, scope), AlarmSystem: systemRow(a, b, scope)}
}

func scopeOf(a, b *model.KPITruth) Scope {
	return Scope{Kind: ScopeWholeRun, ASimulatedLengthNS: spanOf(a), BSimulatedLengthNS: spanOf(b)}
}

func spanOf(kpis *model.KPITruth) *int64 {
	if kpis == nil {
		return nil
	}
	span := kpis.ToNS - kpis.FromNS
	return &span
}

func kpiComparisons(a, b *Report, scope Scope) []KPIRow {
	rows := []KPIRow{}
	for _, e := range mergeKeyed(kpiStats(kpisOf(a)), kpiStats(kpisOf(b))) {
		row := KPIRow{EntityID: e.entity, KPI: e.state, Presence: presenceOf(e.a != nil, e.b != nil), A: e.a, B: e.b, Scope: scope}
		if e.a != nil && e.b != nil {
			row.Delta, row.RelativeDelta = kpiDelta(*e.a, *e.b), kpiRelative(*e.a, *e.b)
		}
		rows = append(rows, row)
	}
	return rows
}

func kpiStats(kpis *model.KPITruth) map[[2]string]KPIStats {
	stats := map[[2]string]KPIStats{}
	if kpis == nil {
		return stats
	}
	for _, entity := range kpis.Entities {
		for _, kpi := range entity.KPIs {
			stats[[2]string{entity.EntityID, kpi.State}] = KPIStats{Mean: kpi.Mean, Min: kpi.Min, Max: kpi.Max, Samples: kpi.Samples}
		}
	}
	return stats
}

func kpiDelta(a, b KPIStats) *KPIDelta {
	return &KPIDelta{Mean: b.Mean - a.Mean, Min: b.Min - a.Min, Max: b.Max - a.Max}
}

func kpiRelative(a, b KPIStats) *KPIRelative {
	return &KPIRelative{Mean: relative(a.Mean, b.Mean), Min: relative(a.Min, b.Min), Max: relative(a.Max, b.Max)}
}

func relative(a, b float64) *float64 {
	if a == 0 {
		return nil
	}
	change := (b - a) / math.Abs(a)
	return &change
}

func presenceOf(inA, inB bool) string {
	switch {
	case inA && inB:
		return PresenceBoth
	case inA:
		return PresenceOnlyA
	}
	return PresenceOnlyB
}

func mergeKeyed[V any](a, b map[[2]string]V) []entry[V] {
	keys := slices.Concat(slices.Collect(maps.Keys(a)), slices.Collect(maps.Keys(b)))
	slices.SortFunc(keys, func(x, y [2]string) int { return slices.Compare(x[:], y[:]) })
	entries := []entry[V]{}
	for _, key := range slices.Compact(keys) {
		entries = append(entries, entry[V]{entity: key[0], state: key[1], a: lookup(a, key), b: lookup(b, key)})
	}
	return entries
}

func lookup[V any](values map[[2]string]V, key [2]string) *V {
	value, found := values[key]
	if !found {
		return nil
	}
	return &value
}
