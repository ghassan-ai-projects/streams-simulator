package domain

import (
	"math"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

type coupling struct {
	sources map[string]map[string][]string
	readers map[string][]string
}

type retention struct {
	horizon int64
	floor   int64
}

func newCoupling(spec *domain.Compiled) coupling {
	c := coupling{sources: map[string]map[string][]string{}, readers: map[string][]string{}}
	for _, link := range spec.Spec.Links {
		if slices.Contains(spec.InstanceKind(spec.InstanceSetFor(link.To)).LinkReads(), link.Type) {
			c.add(link)
		}
	}
	return c
}

func (c coupling) add(link model.Link) {
	if c.sources[link.To] == nil {
		c.sources[link.To] = map[string][]string{}
	}
	c.sources[link.To][link.Type] = insertSorted(c.sources[link.To][link.Type], link.From)
	c.readers[link.From] = insertSorted(c.readers[link.From], link.To)
}

func insertSorted(ids []string, id string) []string {
	at, found := slices.BinarySearch(ids, id)
	if found {
		return ids
	}
	return slices.Insert(ids, at, id)
}

func kindSteps(kind *model.Kind) int64 {
	var steps int64
	for i := range kind.Dynamics {
		steps += stepNS(&kind.Dynamics[i])
	}
	return steps
}

func (w *World) advanceReaders(source *Entity, t int64) {
	if t <= source.readersAt || len(w.coupling.readers[source.ID]) == 0 {
		return
	}
	source.readersAt = t
	for _, id := range w.coupling.readers[source.ID] {
		if reader := w.entities[id]; reader != nil {
			w.advanceEntity(reader, reader.lastAliveAt(t))
		}
	}
}

func (w *World) readersFloor(source *Entity) int64 {
	floor := int64(math.MaxInt64)
	for _, id := range w.coupling.readers[source.ID] {
		if reader := w.entities[id]; reader != nil && !reader.finishedFor(source) {
			floor = min(floor, w.retainedFrom(reader))
		}
	}
	return floor
}

func (w *World) retainedFrom(ent *Entity) int64 {
	oldest := w.readersFloor(ent)
	for i := range ent.kind.Dynamics {
		if s := ent.states[ent.kind.Dynamics[i].Target]; s != nil && ent.kind.Dynamics[i].Tier != "F0" {
			oldest = min(oldest, s.oldest())
		}
	}
	return oldest
}

func (s *stateValue) oldest() int64 {
	if len(s.history) > 0 {
		return s.history[0].lastStep
	}
	return s.lastStep
}
