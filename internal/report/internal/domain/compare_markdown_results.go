package domain

import (
	"strconv"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func resultLines(c *Comparison) []string {
	scope := c.Results.AlarmSystem.Scope
	lines := []string{"Scope: whole run of each run. Simulated length covered by KPI truth: A " + spanText(scope.ASimulatedLengthNS) + ", B " + spanText(scope.BSimulatedLengthNS) + ".", "", "### KPIs", ""}
	lines = append(lines, listOrNone(len(c.Results.KPIs), table(kpiHeader, kpiCells(c.Results.KPIs)))...)
	lines = append(append(lines, "", "### Alarms", ""), listOrNone(len(c.Results.Alarms), table(alarmHeader, alarmCells(c.Results.Alarms)))...)
	return append(append(lines, "", "### Alarm system", ""), systemLines(c.Results.AlarmSystem)...)
}

func spanText(span *int64) string {
	if span == nil {
		return "none"
	}
	return seconds(*span)
}

var kpiHeader = []string{"Entity", "KPI", "Presence", "A mean", "B mean", "Delta", "Relative", "A min..max", "B min..max"}

func kpiCells(rows []KPIRow) [][]string {
	cells := make([][]string, len(rows))
	for i, row := range rows {
		cells[i] = []string{row.EntityID, row.KPI, row.Presence, statText(row.A, func(s KPIStats) float64 { return s.Mean }),
			statText(row.B, func(s KPIStats) float64 { return s.Mean }), deltaText(row.Delta), relativeText(row.RelativeDelta), rangeText(row.A), rangeText(row.B)}
	}
	return cells
}

func statText(stats *KPIStats, pick func(KPIStats) float64) string {
	if stats == nil {
		return ""
	}
	return decimal(pick(*stats))
}

func deltaText(delta *KPIDelta) string {
	if delta == nil {
		return ""
	}
	return decimal(delta.Mean)
}

func relativeText(relative *KPIRelative) string {
	if relative == nil || relative.Mean == nil {
		return ""
	}
	return strconv.FormatFloat(*relative.Mean*100, 'f', 1, 64) + "%"
}

func rangeText(stats *KPIStats) string {
	if stats == nil {
		return ""
	}
	return decimal(stats.Min) + ".." + decimal(stats.Max)
}

var alarmHeader = []string{"Entity", "Alarm", "Presence", "A activations", "B activations", "Delta", "A active s", "B active s"}

func alarmCells(rows []AlarmRow) [][]string {
	cells := make([][]string, len(rows))
	for i, row := range rows {
		cells[i] = []string{row.EntityID, row.Alarm, row.Presence, activations(row.A), activations(row.B), activationDelta(row.Delta), activeSeconds(row.A), activeSeconds(row.B)}
	}
	return cells
}

func activations(stats *AlarmStats) string {
	if stats == nil {
		return ""
	}
	return strconv.Itoa(stats.Activations)
}

func activationDelta(delta *AlarmDelta) string {
	if delta == nil {
		return ""
	}
	return strconv.Itoa(delta.Activations)
}

func activeSeconds(stats *AlarmStats) string {
	if stats == nil {
		return ""
	}
	return decimal(stats.ActiveS)
}

func systemLines(row SystemRow) []string {
	if row.A == nil && row.B == nil {
		return []string{"None."}
	}
	header := []string{"Run", "Activations", "Peak per 10 min", "Flood intervals", "Chattering", "Stale", "Standing at end"}
	return table(header, append(append(systemRows("A", row.A), systemRows("B", row.B)...), systemRows("Delta", row.Delta)...))
}

func systemRows(name string, kpis *model.AlarmSystemKPIs) [][]string {
	if kpis == nil {
		return nil
	}
	return [][]string{append([]string{name}, systemCells(*kpis)...)}
}

func systemCells(k model.AlarmSystemKPIs) []string {
	values := []int{k.Activations, k.PeakPer10Min, k.FloodIntervals, k.Chattering, k.Stale, k.StandingAtEnd}
	cells := make([]string, len(values))
	for i, value := range values {
		cells[i] = strconv.Itoa(value)
	}
	return cells
}
