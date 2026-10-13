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
	return w.linked(ent, linkRead{in.Link.Type, in.State, in.Link.Aggregate, in.Link.SampleS}, t).value(fallback)
}

func (w *World) followValue(ent *Entity, dyn *model.Dynamics, t int64) (float64, bool) {
	if dyn == nil || dyn.Follow == nil {
		return 0, false
	}
	agg := linkAggregate{form: dyn.Follow.Aggregate}
	for _, source := range dyn.Follow.Sources {
		w.gather(&agg, ent, linkRead{source.Link, source.State, dyn.Follow.Aggregate, dyn.Follow.SampleS}, t)
	}
	return agg.value(0), agg.count > 0
}

type linkRead struct {
	linkType, state, form string
	sampleS               float64
}

func (w *World) linked(ent *Entity, read linkRead, t int64) linkAggregate {
	agg := linkAggregate{form: read.form}
	w.gather(&agg, ent, read, t)
	return agg
}

func (w *World) gather(agg *linkAggregate, ent *Entity, read linkRead, t int64) {
	at := w.sampleInstant(t, read.sampleS)
	for _, id := range w.coupling.sources[ent.ID][read.linkType] {
		if source := w.entities[id]; source != nil && source.presentAt(at) {
			agg.add(w.stateAt(id, read.state, at))
		}
	}
}

func (w *World) sampleInstant(t int64, sampleS float64) int64 {
	if sampleS == 0 {
		return t
	}
	period := int64(sampleS * secondsPerNS)
	elapsed := t - w.StartNS
	return w.StartNS + max((elapsed+period-1)/period-1, 0)*period
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
