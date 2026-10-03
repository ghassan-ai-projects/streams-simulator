package world

import (
	"math"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/randutil"
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
	cs.advanceAvailability(a, rng, t)
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
	churn := w.Spec.Spec.Entities.Churn
	if churn == nil {
		return
	}
	id := w.reserveBirthID()
	if err := w.addEntity(id, atNS, nil); err == nil {
		w.scheduleLifetime(id, atNS, churn.MeanLifetimeS)
	}
	rng := w.substream("churn/births")
	next := atNS + int64(rng.Exp(3600*secondsPerNS/churn.BirthsPerHour))
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

func (cs *channelRunState) advanceAvailability(a *model.Availability, rng *randutil.SplitMix64, at int64) {
	if !cs.availInit {
		cs.availInit = true
		cs.availDown = false
		cs.availUntil = at + int64(rng.Exp(mtbfSeconds(a.Uptime, a.MTTRS))*secondsPerNS)
		return
	}
	if at < cs.availUntil {
		return
	}
	cs.renewAvailability(a, rng, at)
}

func (cs *channelRunState) renewAvailability(a *model.Availability, rng *randutil.SplitMix64, at int64) {
	if cs.availDown {
		cs.availDown = false
		cs.availUntil = at + int64(rng.Exp(mtbfSeconds(a.Uptime, a.MTTRS))*secondsPerNS)
	} else {
		cs.availDown = true
		cs.availUntil = at + int64(rng.Exp(mttrSeconds(a.MTTRS))*secondsPerNS)
	}
}

func (w *World) reserveBirthID() string {
	w.nextIndex++
	id := renderID(w.Spec.Spec.Entities.IDTemplate, w.nextIndex, nil)
	for id == "" || w.entities[id] != nil {
		w.nextIndex++
		id = renderID(w.Spec.Spec.Entities.IDTemplate, w.nextIndex, nil)
	}
	return id
}

func (w *World) scheduleLifetime(id string, at int64, mean float64) {
	if mean > 0 {
		rng := w.substream("churn/lifetimes/" + id)
		dieAt := at + int64(rng.Exp(mean)*secondsPerNS)
		w.schedule(kindDeath, id, "", dieAt, nil)
	}
}
