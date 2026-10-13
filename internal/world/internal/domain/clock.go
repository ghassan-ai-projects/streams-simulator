package domain

import (
	"container/heap"
	"errors"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/randutil"
)

func (w *World) NextEventNS() int64 {
	if w.queue.Len() == 0 {
		return 0
	}
	return w.queue[0].timeNS
}

var ErrClockBackwards = errors.New("world: clock would move backwards")

func (w *World) Advance(to int64) (emitted, effectsApplied int, err error) {
	if to < w.clockNS {
		return 0, 0, fmt.Errorf("%w (%d -> %d)", ErrClockBackwards, w.clockNS, to)
	}
	w.emittedThisAdvance = 0
	w.effectsAppliedThis = 0
	w.processScheduledEvents(to)
	w.clockNS = to
	return w.emittedThisAdvance, w.effectsAppliedThis, nil
}

func (w *World) Clock() int64 { return w.clockNS }

func (w *World) StreamKey() string { return w.streamKey }

func (w *World) SetEmitter(emitter func(model.SimEvent)) {
	w.emitter = emitter
}

func (w *World) EmittedCount() int64 { return w.seq }

func (w *World) substream(name string) *randutil.SplitMix64 {
	full := w.streamKey + "/" + name
	if s, ok := w.subs[full]; ok {
		return s
	}
	s := randutil.Substream(w.seed, full)
	w.subs[full] = s
	return s
}

func (w *World) processScheduledEvents(to int64) {
	for w.queue.Len() > 0 {
		event := w.queue[0]
		if event.timeNS > to {
			break
		}
		heap.Pop(&w.queue)
		w.sampleKPIsThrough(event.timeNS - 1)
		w.clockNS = event.timeNS
		w.processScheduledEvent(event)
	}
	w.sampleKPIsThrough(to)
}

func (w *World) processScheduledEvent(event *item) {
	switch event.kind {
	case kindEmission:
		w.processEmission(event.entity, event.channel, event.timeNS)
	case kindEffectStart:
		w.startEffect(event)
	case kindBirth:
		w.birthAutonomous(event.payload.(int), event.timeNS)
	case kindDeath:
		w.retire(event.entity, "lifetime", event.timeNS)
	default:
		w.processChannelEvent(event)
	}
}

func (w *World) processChannelEvent(event *item) {
	switch event.kind {
	case kindEventCheck:
		w.processEventCheck(event.entity, event.channel, event.timeNS)
	case kindObservationCapture:
		w.captureFrozen(event.payload.(*activeFault))
	}
}

func (w *World) startEffect(event *item) {
	if kick, ok := event.payload.(*kick); ok && !kick.applied {
		kick.applied = true
		w.effectsAppliedThis++
	}
}
