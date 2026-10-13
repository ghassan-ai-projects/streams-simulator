package domain

import (
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

const (
	chatterWindowNS = int64(60 * secondsPerNS)
	alarmRateNS     = int64(600 * secondsPerNS)
	floodAlarms     = 10
	staleNS         = int64(24 * 3600 * secondsPerNS)
)

func (w *World) KPITruth() *model.KPITruth {
	if w.kpis == nil || len(w.kpis.order) == 0 {
		return nil
	}
	w.sampleKPIsThrough(w.clockNS)
	truth := &model.KPITruth{FromNS: w.StartNS, ToNS: w.clockNS, Entities: []model.EntityKPIs{}}
	activations := w.collectKPIs(truth)
	truth.AlarmSystem = alarmSystem(truth.Entities, activations)
	return truth
}

func (w *World) collectKPIs(truth *model.KPITruth) []int64 {
	var activations []int64
	for _, id := range w.kpis.order {
		end, retired := w.kpiEnd(id)
		entity, times := w.kpis.entities[id].truth(id, end, retired)
		truth.Entities = append(truth.Entities, entity)
		activations = append(activations, times...)
	}
	return activations
}

func (w *World) kpiEnd(id string) (int64, bool) {
	if ent := w.entities[id]; !ent.alive && ent.retiredNS <= w.clockNS {
		return ent.retiredNS, true
	}
	return w.clockNS, false
}

func (e *entityKPIs) truth(id string, endNS int64, retired bool) (model.EntityKPIs, []int64) {
	entity := model.EntityKPIs{EntityID: id}
	var activations []int64
	for _, track := range e.tracks {
		if !track.alarm {
			entity.KPIs = append(entity.KPIs, track.value())
			continue
		}
		entity.Alarms = append(entity.Alarms, track.alarmKPI(endNS, retired))
		activations = append(activations, track.activations...)
	}
	return entity, activations
}

func alarmSystem(entities []model.EntityKPIs, activations []int64) model.AlarmSystemKPIs {
	slices.Sort(activations)
	system := model.AlarmSystemKPIs{Activations: len(activations), PeakPer10Min: peakRate(activations), FloodIntervals: floodIntervals(activations)}
	for _, entity := range entities {
		for _, alarm := range entity.Alarms {
			system.Chattering += countIf(alarm.Chattering)
			system.Stale += countIf(alarm.LongestActiveS*secondsPerNS >= float64(staleNS))
			system.StandingAtEnd += countIf(alarm.ActiveAtEnd)
		}
	}
	return system
}

func peakRate(sorted []int64) int {
	peak, from := 0, 0
	for to := range sorted {
		for sorted[to]-sorted[from] >= alarmRateNS {
			from++
		}
		peak = max(peak, to-from+1)
	}
	return peak
}

func floodIntervals(sorted []int64) int {
	floods, from := 0, 0
	for to := range sorted {
		if to+1 == len(sorted) || sorted[to+1]/alarmRateNS != sorted[from]/alarmRateNS {
			floods += countIf(to-from+1 > floodAlarms)
			from = to + 1
		}
	}
	return floods
}

func countIf(condition bool) int {
	if condition {
		return 1
	}
	return 0
}
