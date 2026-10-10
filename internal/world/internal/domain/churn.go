package domain

// birthAutonomous creates an entity on the churn schedule.
func (w *World) birthAutonomous(atNS int64) {
	churn := w.Spec.Spec.Entities.Churn
	if churn == nil {
		return
	}
	id := w.reserveBirthID()
	if err := w.addEntity(id, atNS); err == nil {
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
	// Scheduled emissions are dropped when popped (processEmission checks
	// liveness).
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
