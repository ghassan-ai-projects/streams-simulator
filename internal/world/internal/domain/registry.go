package domain

import (
	"fmt"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

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
		if w.entities[id].Type == ent.Type {
			out = append(out, id)
		}
	}
	return out
}

func (w *World) InitialEntityIDs() []string {
	var out []string
	for _, id := range w.entityOrder {
		if w.entities[id] != nil && w.entities[id].BornNS == w.StartNS {
			out = append(out, id)
		}
	}
	return out
}

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

func (w *World) Entity(id string) *Entity {
	live := w.entities[id]
	if live == nil {
		return nil
	}
	return &Entity{ID: live.ID, Type: live.Type, BornNS: live.BornNS, alive: live.alive}
}

func (w *World) StateNamesOf(entityID string) []string {
	ent := w.entities[entityID]
	if ent == nil {
		return nil
	}
	names := make([]string, len(ent.kind.State))
	for i := range ent.kind.State {
		names[i] = ent.kind.State[i].Name
	}
	slices.Sort(names)
	return names
}

func (w *World) EffectorOf(entityID, name string) *model.Effector {
	ent := w.entities[entityID]
	if ent == nil {
		return nil
	}
	return ent.kind.EffectorDecl(name)
}
