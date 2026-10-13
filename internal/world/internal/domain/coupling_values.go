package domain

import "github.com/ghassan-ai-projects/streams-simulator/internal/model"

const runningThreshold = 0.5

type linkAggregate struct {
	form    string
	count   int
	running int
	sum     float64
	low     float64
	high    float64
}

func (w *World) inputState(ent *Entity, in model.F1Input, t int64) float64 {
	if in.Link == nil {
		return w.stateAt(ent.ID, in.State, t)
	}
	var fallback float64
	if in.Link.Default != nil {
		fallback = *in.Link.Default
	}
	return w.linked(ent, in.Link.Type, in.State, in.Link.Aggregate, t).value(fallback)
}

func (w *World) followValue(ent *Entity, dyn *model.Dynamics, t int64) (float64, bool) {
	if dyn == nil || dyn.Follow == nil {
		return 0, false
	}
	agg := w.linked(ent, dyn.Follow.Link, dyn.Follow.State, dyn.Follow.Aggregate, t)
	return agg.value(0), agg.count > 0
}

func (w *World) linked(ent *Entity, linkType, state, form string, t int64) linkAggregate {
	agg := linkAggregate{form: form}
	for _, id := range w.coupling.sources[ent.ID][linkType] {
		if source := w.entities[id]; source != nil && source.presentAt(t) {
			agg.add(w.stateAt(id, state, t))
		}
	}
	return agg
}

func (a *linkAggregate) add(v float64) {
	if a.count == 0 {
		a.low, a.high = v, v
	}
	a.count++
	a.sum += v
	a.low, a.high = min(a.low, v), max(a.high, v)
	if v >= runningThreshold {
		a.running++
	}
}

func (a linkAggregate) value(fallback float64) float64 {
	switch a.form {
	case model.AggregateCountRunning:
		return float64(a.running)
	case model.AggregateSum:
		return a.sum
	}
	if a.count == 0 {
		return fallback
	}
	return a.spread()
}

func (a linkAggregate) spread() float64 {
	switch a.form {
	case model.AggregateMean:
		return a.sum / float64(a.count)
	case model.AggregateMin:
		return a.low
	}
	return a.high
}
