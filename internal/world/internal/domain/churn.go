package domain

func (w *World) birthAutonomous(set int, atNS int64) {
	churn := w.Spec.Spec.Instances[set].Churn
	if churn == nil {
		return
	}
	id := w.reserveBirthID(set)
	if err := w.addEntity(id, set, atNS); err == nil {
		w.scheduleLifetime(id, atNS, churn.MeanLifetimeS)
	}
	rng := w.substream(birthStream(set))
	next := atNS + int64(rng.Exp(3600*secondsPerNS/churn.BirthsPerHour))
	w.schedule(kindBirth, "", "", next, set)
}

func (w *World) Retire(entityID, reason string, atNS int64) {
	w.retire(entityID, reason, atNS)
}

func (w *World) retire(entityID, reason string, atNS int64) {
	ent := w.entities[entityID]
	if ent == nil || !ent.alive {
		return
	}
	ent.alive = false
	ent.retiredNS = atNS

}

func (w *World) reserveBirthID(set int) string {
	template := w.Spec.Spec.Instances[set].IDTemplate
	w.nextIndex[set]++
	id := renderID(template, w.nextIndex[set], nil)
	for id == "" || w.entities[id] != nil {
		w.nextIndex[set]++
		id = renderID(template, w.nextIndex[set], nil)
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
