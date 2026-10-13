package domain

import (
	"strconv"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func kpiSections(kpis *model.KPITruth) []section {
	if kpis == nil {
		return nil
	}
	values, alarms := kpiRows(kpis.Entities)
	return []section{
		{"KPIs", listOrNone(len(values), table([]string{"Entity", "KPI", "Mean", "Min", "Max", "Samples"}, values))},
		{"Alarms", listOrNone(len(alarms), table([]string{"Entity", "Alarm", "Priority", "Activations", "Chattering", "Active s", "Longest s", "Active at end"}, alarms))},
		{"Alarm system", alarmSystemLines(kpis.AlarmSystem)},
	}
}

func kpiRows(entities []model.EntityKPIs) (values, alarms [][]string) {
	for _, entity := range entities {
		for _, kpi := range entity.KPIs {
			values = append(values, []string{entity.EntityID, kpi.State, decimal(kpi.Mean), decimal(kpi.Min), decimal(kpi.Max), strconv.Itoa(kpi.Samples)})
		}
		for _, alarm := range entity.Alarms {
			alarms = append(alarms, []string{entity.EntityID, alarm.State, alarm.Priority, strconv.Itoa(alarm.Activations), yesNo(alarm.Chattering),
				decimal(alarm.ActiveS), decimal(alarm.LongestActiveS), yesNo(alarm.ActiveAtEnd)})
		}
	}
	return values, alarms
}

func alarmSystemLines(system model.AlarmSystemKPIs) []string {
	return table([]string{"Activations", "Peak per 10 min", "Flood intervals", "Chattering", "Stale", "Standing at end"}, [][]string{{
		strconv.Itoa(system.Activations), strconv.Itoa(system.PeakPer10Min), strconv.Itoa(system.FloodIntervals),
		strconv.Itoa(system.Chattering), strconv.Itoa(system.Stale), strconv.Itoa(system.StandingAtEnd)}})
}

func decimal(value float64) string {
	return strconv.FormatFloat(value, 'f', 3, 64)
}
