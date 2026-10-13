package domain

import (
	"container/heap"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

const kpiDefaultStepNS = int64(60 * secondsPerNS)

type kpiTracker struct {
	queue    priorityQueue
	tiebreak uint64
	entities map[string]*entityKPIs
	order    []string
}

type entityKPIs struct {
	stepNS int64
	tracks []*kpiTrack
	doneNS int64
}

func newKPITracker(enabled bool) *kpiTracker {
	if !enabled {
		return nil
	}
	return &kpiTracker{entities: map[string]*entityKPIs{}}
}

func (w *World) trackKPIs(ent *Entity, atNS int64) {
	tracks := kpiTracks(ent.kind)
	if w.kpis == nil || len(tracks) == 0 {
		return
	}
	w.kpis.entities[ent.ID] = &entityKPIs{stepNS: kpiStep(ent.kind), tracks: tracks, doneNS: -1}
	w.kpis.order = append(w.kpis.order, ent.ID)
	w.kpis.schedule(ent.ID, atNS)
}

func kpiTracks(kind *model.Kind) []*kpiTrack {
	var tracks []*kpiTrack
	for _, state := range kind.KPIs {
		tracks = append(tracks, &kpiTrack{state: state, activeSince: -1})
	}
	for i := range kind.Dynamics {
		if dyn := &kind.Dynamics[i]; dyn.F1 != nil && dyn.F1.Form == model.FormAlarm {
			tracks = append(tracks, &kpiTrack{state: dyn.Target, alarm: true, priority: dyn.F1.Alarm.Priority, activeSince: -1})
		}
	}
	return tracks
}

func kpiStep(kind *model.Kind) int64 {
	step := kpiDefaultStepNS
	for i := range kind.Dynamics {
		if dyn := &kind.Dynamics[i]; dyn.F1 != nil && dyn.F1.Form == model.FormAlarm {
			step = min(step, stepNS(dyn))
		}
	}
	return step
}

func (k *kpiTracker) schedule(entity string, atNS int64) {
	k.tiebreak++
	heap.Push(&k.queue, &item{timeNS: atNS, tiebreak: k.tiebreak, entity: entity})
}

func (w *World) sampleKPIsThrough(to int64) {
	for w.kpis != nil && w.kpis.queue.Len() > 0 && w.kpis.queue[0].timeNS <= to {
		sample := heap.Pop(&w.kpis.queue).(*item)
		w.sampleKPIs(sample.entity, sample.timeNS)
	}
}

func (w *World) sampleKPIs(id string, at int64) {
	ent, tracked := w.entities[id], w.kpis.entities[id]
	if !ent.alive && at > ent.retiredNS {
		tracked.doneNS = ent.retiredNS
		return
	}
	for _, track := range tracked.tracks {
		track.sample(at, w.stateOf(ent, track.state, at))
	}
	w.kpis.schedule(id, at+tracked.stepNS)
}
