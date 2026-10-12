package domain

import "fmt"

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

func (w *World) PeerIDs(entityID string) []string {
	ent := w.entities[entityID]
	if ent == nil {
		return nil
	}
	var out []string
	for _, id := range w.EntityIDs() {
		if w.entities[id].kind == ent.kind {
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
// go through birthAutonomous, which schedules the death). An entity cannot
// be born before the world clock: the world would have to simulate a history
// that already passed.
func (w *World) AddEntity(id string, atNS int64) error {
	if atNS < w.clockNS {
		return fmt.Errorf("world: entity %q cannot be born at %d, before the world clock %d", id, atNS, w.clockNS)
	}
	set := w.instanceSetOf(id)
	if set < 0 {
		return fmt.Errorf("world: entity id %q fits no instance set of the domain", id)
	}
	return w.addEntity(id, set, atNS)
}

// Entity returns a snapshot of the entity by id, or nil. The caller gets a
// copy of the identity fields: writing to it cannot change the world.
func (w *World) Entity(id string) *Entity {
	live := w.entities[id]
	if live == nil {
		return nil
	}
	return &Entity{ID: live.ID, Type: live.Type, BornNS: live.BornNS, alive: live.alive}
}
