package world

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"math"
)

// updateAvailability advances the producer up/down renewal process. The
// exponential sojourns are in seconds and converted to ns, like every other
// Exp() usage in the world.
func (w *World) updateAvailability(ent *Entity, ch *model.Channel, cs *channelRunState, t int64) error {
	rng := w.substream(ent.ID + "/" + ch.Name + "/availability")
	a := ch.Availability
	if a.Uptime >= 1 {
		cs.availInit = true
		cs.availDown = false
		cs.availUntil = 0
		return nil
	}
	if !cs.availInit {
		cs.availInit = true
		cs.availDown = false
		cs.availUntil = t + int64(rng.Exp(mtbfSeconds(a.Uptime, a.MTTRS))*secondsPerNS)
		return nil
	}
	if t < cs.availUntil {
		return nil
	}
	if cs.availDown {
		cs.availDown = false
		cs.availUntil = t + int64(rng.Exp(mtbfSeconds(a.Uptime, a.MTTRS))*secondsPerNS)
	} else {
		cs.availDown = true
		cs.availUntil = t + int64(rng.Exp(mttrSeconds(a.MTTRS))*secondsPerNS)
	}
	return nil
}

func mttrSeconds(mttr float64) float64 {
	if mttr <= 0 {
		return 1
	}
	return mttr
}

func mtbfSeconds(uptime, mttr float64) float64 {
	if uptime >= 1 {
		return math.Inf(1)
	}
	if uptime <= 0 {
		return mttrSeconds(mttr)
	}
	return mttrSeconds(mttr) * uptime / (1 - uptime)
}

// isConfirmationChannel reports whether the channel independently reports an
// effector's physical effect (the confirmation_channels declarations).
func (w *World) isConfirmationChannel(name string) bool {
	for i := range w.Spec.Spec.Effectors {
		for _, c := range w.Spec.Spec.Effectors[i].ConfirmationChannels {
			if c == name {
				return true
			}
		}
	}
	return false
}

// birthAutonomous creates an entity on the churn schedule.
func (w *World) birthAutonomous(atNS int64) {
	ch := w.Spec.Spec.Entities.Churn
	if ch == nil {
		return
	}
	w.nextIndex++
	id := renderID(w.Spec.Spec.Entities.IDTemplate, w.nextIndex, nil)
	for id == "" || w.entities[id] != nil {
		w.nextIndex++
		id = renderID(w.Spec.Spec.Entities.IDTemplate, w.nextIndex, nil)
	}
	if err := w.addEntity(id, atNS, nil); err == nil {
		// Schedule its death.
		if ch.MeanLifetimeS > 0 {
			rng := w.substream("churn/lifetimes/" + id)
			dieAt := atNS + int64(rng.Exp(ch.MeanLifetimeS)*secondsPerNS)
			w.schedule(kindDeath, id, "", dieAt, nil)
		}
	}
	// Schedule the next birth.
	rng := w.substream("churn/births")
	next := atNS + int64(rng.Exp(3600*secondsPerNS/ch.BirthsPerHour))
	w.schedule(kindBirth, "", "", next, nil)
}

// Retire removes an entity and its scheduled emissions. Used for churn and
// by the run layer for entity.retire commands.
func (w *World) Retire(entityID, reason string, atNS int64) {
	w.retire(entityID, reason, atNS)
}

func (w *World) retire(entityID, reason string, atNS int64) {
	ent := w.entities[entityID]
	if ent == nil || !ent.alive {
		return
	}
	ent.alive = false
	ent.RetiredNS = atNS
	// Scheduled emissions are dropped when popped (processEmission checks
	// liveness).
}
