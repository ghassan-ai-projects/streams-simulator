package world

import (
	"container/heap"
	"fmt"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/randutil"
	"math"
)

// NextEventNS is the time of the next scheduled event.
func (w *World) NextEventNS() int64 {
	if w.queue.Len() == 0 {
		return 0
	}
	return w.queue[0].timeNS
}

// PendingEvents is the number of scheduled events.
func (w *World) PendingEvents() int { return w.queue.Len() }

// Advance processes every event scheduled at or before to, then sets the
// clock to to. Moving the clock backwards is refused.
func (w *World) Advance(to int64) (int, int, error) {
	if to < w.ClockNS {
		return 0, 0, fmt.Errorf("world: clock would move backwards (%d -> %d)", w.ClockNS, to)
	}
	w.emittedThisAdvance = 0
	w.effectsAppliedThis = 0
	for w.queue.Len() > 0 {
		it := w.queue[0]
		if it.timeNS > to {
			break
		}
		heap.Pop(&w.queue)
		w.ClockNS = it.timeNS
		switch it.kind {
		case kindEmission:
			w.processEmission(it.entity, it.channel, it.timeNS)
		case kindEffectStart:
			if k, ok := it.payload.(*kick); ok && !k.applied {
				k.applied = true
				w.effectsAppliedThis++
			}
		case kindBirth:
			w.birthAutonomous(it.timeNS)
		case kindDeath:
			w.retire(it.entity, "lifetime", it.timeNS)
		}
	}
	w.ClockNS = to
	return w.emittedThisAdvance, w.effectsAppliedThis, nil
}

// Clock is the current world time.
func (w *World) Clock() int64 { return w.ClockNS }

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
	s := randutil.Substream(w.Seed, full)
	w.subs[full] = s
	return s
}

// EntityIDs returns the live entity ids in creation order.
func (w *World) EntityIDs() []string {
	var out []string
	for _, id := range w.entityOrder {
		if w.entities[id] != nil && w.entities[id].alive {
			out = append(out, id)
		}
	}
	return out
}

// InitialEntityIDs returns the entities created at world construction, in
// order (before any churn or entity.add).
func (w *World) InitialEntityIDs() []string {
	var out []string
	for _, id := range w.entityOrder {
		if w.entities[id] != nil && w.entities[id].BornNS == w.StartNS {
			out = append(out, id)
		}
	}
	return out
}

// AddEntity creates an entity at the given time (entity.add; churn births
// go through birthAutonomous, which schedules the death).
func (w *World) AddEntity(id string, atNS int64, params map[string]any) error {
	return w.addEntity(id, atNS, params)
}

// Entity returns the entity by id, or nil.
func (w *World) Entity(id string) *Entity { return w.entities[id] }

// DynamicsFor returns the dynamics declaration for a state, or nil.
func (w *World) DynamicsFor(state string) *model.Dynamics {
	for i := range w.Spec.Spec.Dynamics {
		if w.Spec.Spec.Dynamics[i].Target == state {
			return &w.Spec.Spec.Dynamics[i]
		}
	}
	return nil
}

// StateValue exposes a hidden state to the director only (solver and truth).
func (w *World) StateValue(entity, state string, t int64) float64 {
	return w.stateAt(entity, state, t)
}

// quantize rounds v to the channel's declared resolution.
func quantize(v, resolution float64) float64 {
	if resolution <= 0 {
		return v
	}
	return math.Round(v/resolution) * resolution
}
