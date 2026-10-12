package domain

import "github.com/ghassan-ai-projects/streams-simulator/internal/model"

type storeExchange struct {
	inputs     []float64
	capacities []float64
	links      []storeLink
}

type storeLink struct {
	from, to    int
	conductance float64
}

func (w *World) stepStores(ent *Entity, group *storeGroup, at, dt int64) []float64 {
	exchange := w.exchangeAt(ent, group, at)
	return rk4Levels(group.levels(), float64(dt)/secondsPerNS, exchange.derivative)
}

func (w *World) exchangeAt(ent *Entity, group *storeGroup, at int64) *storeExchange {
	exchange := &storeExchange{}
	for _, dyn := range group.dynamics {
		exchange.inputs = append(exchange.inputs, w.inputValue(ent, dyn, at))
		exchange.capacities = append(exchange.capacities, dyn.F1.Capacity)
	}
	for i := range ent.kind.Flows {
		flow := &ent.kind.Flows[i]
		exchange.links = append(exchange.links, storeLink{from: group.index(flow.From), to: group.index(flow.To),
			conductance: flow.Conductance * w.flowScale(ent, flow, at)})
	}
	return exchange
}

func (w *World) flowScale(ent *Entity, flow *model.Flow, at int64) float64 {
	scale := 1.0
	for _, name := range flow.Scale {
		scale *= w.stateAt(ent.ID, name, at)
	}
	return scale
}

func (exchange *storeExchange) derivative(levels []float64) []float64 {
	rates := append([]float64(nil), exchange.inputs...)
	for _, link := range exchange.links {
		moved := link.conductance * (levels[link.from] - levelOr(levels, link.to))
		rates[link.from] -= moved
		if link.to >= 0 {
			rates[link.to] += moved
		}
	}
	for i := range rates {
		rates[i] /= exchange.capacities[i]
	}
	return rates
}

func levelOr(levels []float64, index int) float64 {
	if index < 0 {
		return 0
	}
	return levels[index]
}

func rk4Levels(x []float64, h float64, f func([]float64) []float64) []float64 {
	k1 := f(x)
	k2 := f(offset(x, k1, h/2))
	k3 := f(offset(x, k2, h/2))
	k4 := f(offset(x, k3, h))
	next := make([]float64, len(x))
	for i := range x {
		next[i] = x[i] + h/6*(k1[i]+2*k2[i]+2*k3[i]+k4[i])
	}
	return next
}

func offset(x, k []float64, h float64) []float64 {
	out := make([]float64, len(x))
	for i := range x {
		out[i] = x[i] + h*k[i]
	}
	return out
}
