package world

import (
	"container/heap"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/randutil"
)

// NextEventNS is the time of the next scheduled event.
func (w *World) NextEventNS() int64 {
	if w.queue.Len() == 0 {
		return 0
	}
	return w.queue[0].timeNS
}

// Advance processes every event scheduled at or before to, then sets the
// clock to to, and reports how many native events were emitted and how many
// effect kicks started. Moving the clock backwards is refused.
func (w *World) Advance(to int64) (emitted, effectsApplied int, err error) {
	if to < w.clockNS {
		return 0, 0, fmt.Errorf("world: clock would move backwards (%d -> %d)", w.clockNS, to)
	}
	w.emittedThisAdvance = 0
	w.effectsAppliedThis = 0
	w.processScheduledEvents(to)
	w.clockNS = to
	return w.emittedThisAdvance, w.effectsAppliedThis, nil
}

// Clock is the current world time.
func (w *World) Clock() int64 { return w.clockNS }

// SetEmitter installs the event sink. Events flow world -> emitter.
func (w *World) SetEmitter(emitter func(model.SimEvent)) {
	w.emitter = emitter
}

// EmittedCount is the total number of native events produced.
func (w *World) EmittedCount() int64 { return w.seq }

// substream returns the cached PRNG for a named substream under this world's
// seed. Names are "<world_id>/<entity>/<channel>/<purpose>"; the world id
// prefix keeps distinct worlds independent even under the same seed.
func (w *World) substream(name string) *randutil.SplitMix64 {
	full := w.ID + "/" + name
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
		w.clockNS = event.timeNS
		w.processScheduledEvent(event)
	}
}

func (w *World) processScheduledEvent(event *item) {
	switch event.kind {
	case kindEmission:
		w.processEmission(event.entity, event.channel, event.timeNS)
	case kindEffectStart:
		if kick, ok := event.payload.(*kick); ok && !kick.applied {
			kick.applied = true
			w.effectsAppliedThis++
		}
	case kindBirth:
		w.birthAutonomous(event.timeNS)
	case kindDeath:
		w.retire(event.entity, "lifetime", event.timeNS)
	}
}
