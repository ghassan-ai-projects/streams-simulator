package domain

import (
	"cmp"
	"math"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

type coupling struct {
	sources map[string]map[string][]string
	readers map[string][]readerEdge
}

type readerEdge struct {
	reader string
	reach  int64
}

type retention struct {
	horizon int64
	floor   int64
}

func newCoupling(spec *domain.Compiled) coupling {
	c := coupling{sources: map[string]map[string][]string{}, readers: map[string][]readerEdge{}}
	for _, link := range spec.Spec.Links {
		if reach, read := sampleReach(spec.InstanceKind(spec.InstanceSetFor(link.To)), link.Type); read {
			c.add(link, reach)
		}
	}
	return c
}

func sampleReach(reader *model.Kind, linkType string) (int64, bool) {
	var reach int64
	read := false
	for _, r := range reader.LinkReadsDetailed() {
		if r.Type != linkType {
			continue
		}
		if r.SampleS == 0 || (read && reach == 0) {
			reach, read = 0, true
			continue
		}
		reach, read = max(reach, int64(r.SampleS*secondsPerNS)), true
	}
	return reach, read
}

func (c coupling) add(link model.Link, reach int64) {
	if c.sources[link.To] == nil {
		c.sources[link.To] = map[string][]string{}
	}
	c.sources[link.To][link.Type] = insertSorted(c.sources[link.To][link.Type], link.From)
	c.readers[link.From] = mergeReader(c.readers[link.From], readerEdge{reader: link.To, reach: reach})
}

func mergeReader(edges []readerEdge, edge readerEdge) []readerEdge {
	at, found := slices.BinarySearchFunc(edges, edge.reader, func(e readerEdge, id string) int { return cmp.Compare(e.reader, id) })
	if !found {
		return slices.Insert(edges, at, edge)
	}
	if edges[at].reach == 0 || edge.reach == 0 {
		edges[at].reach = 0
	} else {
		edges[at].reach = max(edges[at].reach, edge.reach)
	}
	return edges
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
	for _, edge := range w.coupling.readers[source.ID] {
		if reader := w.entities[edge.reader]; reader != nil {
			w.advanceEntity(reader, reader.lastAliveAt(t))
		}
	}
}

func (w *World) readersFloor(source *Entity) int64 {
	floor := int64(math.MaxInt64)
	for _, edge := range w.coupling.readers[source.ID] {
		if reader := w.entities[edge.reader]; reader != nil && !reader.finishedFor(source) {
			floor = min(floor, w.readerNeed(reader, edge.reach))
		}
	}
	return floor
}

func (w *World) readerNeed(reader *Entity, reach int64) int64 {
	if reach == 0 {
		return w.retainedFrom(reader)
	}
	return reader.frontier() - reach - 4*kindSteps(reader.kind)
}

func (e *Entity) frontier() int64 {
	front := e.BornNS
	first := true
	for i := range e.kind.Dynamics {
		if s := e.states[e.kind.Dynamics[i].Target]; s != nil && e.kind.Dynamics[i].Integrated() && (first || s.lastStep < front) {
			front, first = s.lastStep, false
		}
	}
	return front
}

func (w *World) retainedFrom(ent *Entity) int64 {
	oldest := w.readersFloor(ent)
	for i := range ent.kind.Dynamics {
		if s := ent.states[ent.kind.Dynamics[i].Target]; s != nil && ent.kind.Dynamics[i].Integrated() {
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
