package domain

import "github.com/ghassan-ai-projects/streams-simulator/internal/model"

type AlarmRow struct {
	EntityID string      `json:"entity_id"`
	Alarm    string      `json:"alarm"`
	Presence string      `json:"presence"`
	A        *AlarmStats `json:"a"`
	B        *AlarmStats `json:"b"`
	Delta    *AlarmDelta `json:"delta,omitempty"`
	Scope    Scope       `json:"scope"`
}

type AlarmStats struct {
	Priority       string  `json:"priority"`
	Activations    int     `json:"activations"`
	Chattering     bool    `json:"chattering"`
	ActiveS        float64 `json:"active_s"`
	LongestActiveS float64 `json:"longest_active_s"`
	ActiveAtEnd    bool    `json:"active_at_end"`
}

type AlarmDelta struct {
	Activations    int     `json:"activations"`
	ActiveS        float64 `json:"active_s"`
	LongestActiveS float64 `json:"longest_active_s"`
}

type SystemRow struct {
	A     *model.AlarmSystemKPIs `json:"a"`
	B     *model.AlarmSystemKPIs `json:"b"`
	Delta *model.AlarmSystemKPIs `json:"delta,omitempty"`
	Scope Scope                  `json:"scope"`
}

func alarmRows(a, b *Report, scope Scope) []AlarmRow {
	rows := []AlarmRow{}
	for _, e := range mergeKeyed(alarmStats(kpisOf(a)), alarmStats(kpisOf(b))) {
		row := AlarmRow{EntityID: e.entity, Alarm: e.state, Presence: presenceOf(e.a != nil, e.b != nil), A: e.a, B: e.b, Scope: scope}
		if e.a != nil && e.b != nil {
			row.Delta = &AlarmDelta{Activations: e.b.Activations - e.a.Activations, ActiveS: e.b.ActiveS - e.a.ActiveS,
				LongestActiveS: e.b.LongestActiveS - e.a.LongestActiveS}
		}
		rows = append(rows, row)
	}
	return rows
}

func alarmStats(kpis *model.KPITruth) map[[2]string]AlarmStats {
	stats := map[[2]string]AlarmStats{}
	if kpis == nil {
		return stats
	}
	for _, entity := range kpis.Entities {
		for _, alarm := range entity.Alarms {
			stats[[2]string{entity.EntityID, alarm.State}] = AlarmStats{Priority: alarm.Priority, Activations: alarm.Activations,
				Chattering: alarm.Chattering, ActiveS: alarm.ActiveS, LongestActiveS: alarm.LongestActiveS, ActiveAtEnd: alarm.ActiveAtEnd}
		}
	}
	return stats
}

func systemRow(a, b *Report, scope Scope) SystemRow {
	row := SystemRow{Scope: scope}
	if kpis := kpisOf(a); kpis != nil {
		row.A = &kpis.AlarmSystem
	}
	if kpis := kpisOf(b); kpis != nil {
		row.B = &kpis.AlarmSystem
	}
	if row.A != nil && row.B != nil {
		row.Delta = systemDelta(*row.A, *row.B)
	}
	return row
}

func systemDelta(a, b model.AlarmSystemKPIs) *model.AlarmSystemKPIs {
	return &model.AlarmSystemKPIs{Activations: b.Activations - a.Activations, PeakPer10Min: b.PeakPer10Min - a.PeakPer10Min,
		FloodIntervals: b.FloodIntervals - a.FloodIntervals, Chattering: b.Chattering - a.Chattering, Stale: b.Stale - a.Stale,
		StandingAtEnd: b.StandingAtEnd - a.StandingAtEnd}
}
