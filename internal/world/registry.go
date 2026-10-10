package world

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
func (w *World) AddEntity(id string, atNS int64) error {
	return w.addEntity(id, atNS)
}

// Entity returns the entity by id, or nil.
func (w *World) Entity(id string) *Entity { return w.entities[id] }
