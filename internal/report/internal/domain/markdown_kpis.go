package domain

import (
	"fmt"
	"strconv"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

var (
	kpiChangeHeaders = []string{"Entity", "KPI", "Unit", "Start", "End", "Change"}
	kpiValueHeaders  = []string{"Entity", "KPI", "Unit", "Mean", "Min", "Max", "Samples"}
)

type kpiEnds map[string]model.KPIOutcome

func kpiSections(truth *Truth, ends kpiEnds) []section {
	if truth == nil || truth.KPIs == nil {
		return nil
	}
	return []section{
		{"KPIs", kpiLines(truth.KPIs.Entities, truth.KPIUnits, ends)},
		{"Alarms", alarmLines(truth.KPIs.Entities)},
		{"Alarm system", alarmSystemLines(truth.KPIs.AlarmSystem)},
	}
}

func kpiLines(entities []model.EntityKPIs, units map[string]string, ends kpiEnds) []string {
	var changes, values [][]string
	for _, entity := range entities {
		for _, kpi := range entity.KPIs {
			if end, ok := ends[kpiKey(entity.EntityID, kpi.State)]; ok {
				changes = append(changes, changeRow(entity.EntityID, end, units))
			} else {
				values = append(values, valueRow(entity.EntityID, kpi, units))
			}
		}
	}
	return listOrNone(len(changes)+len(values), joinTables(optionalTable(kpiChangeHeaders, changes), optionalTable(kpiValueHeaders, values)))
}

func changeRow(entityID string, kpi model.KPIOutcome, units map[string]string) []string {
	return []string{entityID, kpi.State, units[kpi.State], decimal(kpi.Start), decimal(kpi.End), fmt.Sprintf("%+.3f", kpi.End-kpi.Start)}
}

func valueRow(entityID string, kpi model.KPIValue, units map[string]string) []string {
	return []string{entityID, kpi.State, units[kpi.State], decimal(kpi.Mean), decimal(kpi.Min), decimal(kpi.Max), strconv.Itoa(kpi.Samples)}
}

func alarmLines(entities []model.EntityKPIs) []string {
	var alarms [][]string
	for _, entity := range entities {
		for _, alarm := range entity.Alarms {
			alarms = append(alarms, []string{entity.EntityID, alarm.State, alarm.Priority, strconv.Itoa(alarm.Activations), yesNo(alarm.Chattering),
				decimal(alarm.ActiveS), decimal(alarm.LongestActiveS), yesNo(alarm.ActiveAtEnd)})
		}
	}
	return listOrNone(len(alarms), table([]string{"Entity", "Alarm", "Priority", "Activations", "Chattering", "Active s", "Longest s", "Active at end"}, alarms))
}

func asRunKPIs(scoring *OutcomeScoring) kpiEnds {
	if scoring == nil || scoring.Result == nil {
		return nil
	}
	for _, run := range scoring.Result.Runs {
		if run.Mode == model.OutcomeAsRun && run.Facts != nil {
			return kpiEndsOf(run.Facts)
		}
	}
	return nil
}

func kpiEndsOf(facts *model.OutcomeFacts) kpiEnds {
	ends := kpiEnds{}
	for _, entity := range facts.Entities {
		for _, kpi := range entity.KPIs {
			ends[kpiKey(entity.EntityID, kpi.State)] = kpi
		}
	}
	return ends
}

func kpiKey(entityID, state string) string {
	return entityID + "\x00" + state
}

func optionalTable(headers []string, rows [][]string) []string {
	if len(rows) == 0 {
		return nil
	}
	return table(headers, rows)
}

func joinTables(first, second []string) []string {
	if len(first) == 0 || len(second) == 0 {
		return append(first, second...)
	}
	return append(append(first, ""), second...)
}

func alarmSystemLines(system model.AlarmSystemKPIs) []string {
	return table([]string{"Activations", "Peak per 10 min", "Flood intervals", "Chattering", "Stale", "Standing at end"}, [][]string{{
		strconv.Itoa(system.Activations), strconv.Itoa(system.PeakPer10Min), strconv.Itoa(system.FloodIntervals),
		strconv.Itoa(system.Chattering), strconv.Itoa(system.Stale), strconv.Itoa(system.StandingAtEnd)}})
}

func decimal(value float64) string {
	return strconv.FormatFloat(value, 'f', 3, 64)
}
