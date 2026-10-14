package domain

import (
	"math"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func (w *World) FaultIntervals() []model.FaultInterval {
	return w.faultIntervals(math.MaxInt64)
}

func (w *World) FinalFaultIntervals() []model.FaultInterval {
	return w.faultIntervals(w.clockNS)
}

func (w *World) faultIntervals(through int64) []model.FaultInterval {
	intervals := make([]model.FaultInterval, 0, len(w.faultOrder))
	for _, id := range w.faultOrder {
		intervals = append(intervals, w.faultsByID[id].interval(id, through))
	}
	return intervals
}

func (f *activeFault) interval(id string, through int64) model.FaultInterval {
	interval := model.FaultInterval{FaultID: id, EntityID: f.entity, Fault: f.fault.ID, InjectedAtNS: f.injectedNS, OnsetNS: f.onsetNS}
	if f.clearedNS > 0 && f.clearedNS <= through {
		cleared, by := f.clearedNS, f.clearedBy
		interval.ClearedNS, interval.ClearedBy = &cleared, &by
	}
	return interval
}
