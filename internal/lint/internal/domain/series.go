package domain

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

type series struct {
	entity  string
	kind    *model.Kind
	name    string
	state   *model.State
	dynamic *model.Dynamics
	channel *model.Channel
	values  []float64
}

type plant struct {
	id       string
	kind     *model.Kind
	all      []*series
	channels map[string]*series
}

func samplePlants(w *world.World, spec *domain.Compiled, g *grid) []*plant {
	var plants []*plant
	for _, id := range w.EntityIDs() {
		plants = append(plants, newPlant(id, spec.InstanceKind(max(spec.InstanceSetFor(id), 0)), len(g.timesNS)))
	}
	for i, t := range g.timesNS {
		for _, p := range plants {
			p.sample(w, i, t)
		}
	}
	return plants
}

func newPlant(id string, kind *model.Kind, count int) *plant {
	p := &plant{id: id, kind: kind, channels: map[string]*series{}}
	for i := range kind.State {
		state := &kind.State[i]
		p.all = append(p.all, &series{entity: id, kind: kind, name: state.Name, state: state, dynamic: kind.DynamicsOf(state.Name), values: make([]float64, count)})
	}
	for i := range kind.Channels {
		if channel := &kind.Channels[i]; isNumeric(channel) {
			p.channels[channel.Name] = &series{entity: id, kind: kind, name: channel.Name, channel: channel, values: make([]float64, count)}
			p.all = append(p.all, p.channels[channel.Name])
		}
	}
	return p
}

func isNumeric(channel *model.Channel) bool {
	return channel.ValueType == "number" || channel.ValueType == "counter"
}

func (p *plant) sample(w *world.World, i int, t int64) {
	for _, s := range p.all {
		s.values[i] = s.read(w, t)
	}
}

func (s *series) read(w *world.World, t int64) float64 {
	if s.state != nil {
		return w.StateValue(s.entity, s.name, t)
	}
	return w.Reading(s.entity, s.name, t)
}

func (s *series) finding(severity, code string, at int, g *grid, detail string) Finding {
	f := Finding{Severity: severity, Code: code, Entity: s.entity, AtS: g.atS(at), Detail: detail}
	if s.state != nil {
		f.State = s.name
	} else {
		f.Channel = s.name
	}
	return f
}
