package domain

import (
	"container/heap"
	"fmt"
	"strconv"
	"strings"
)

func renderID(tmpl string, n int, params map[string]any) string {
	out := strings.ReplaceAll(tmpl, "{n}", strconv.Itoa(n))
	for {
		start, end := strings.Index(out, "{"), strings.Index(out, "}")
		if start < 0 || end < 0 || end < start {
			return out
		}
		out = out[:start] + entityTemplateValue(params, out[start+1:end]) + out[end+1:]
	}
}

func (w *World) addEntity(id string, set int, atNS int64) error {
	if _, exists := w.entities[id]; exists {
		return fmt.Errorf("world: entity %q already exists", id)
	}
	ent := w.newEntity(id, set, atNS)
	w.initializeEntityState(ent, atNS)
	w.entities[id] = ent
	w.entityOrder = append(w.entityOrder, id)
	w.initializeEntityChannels(ent, id, atNS)
	return nil
}

func (w *World) scheduleChurn(set int, startNS int64) {
	ch := w.Spec.Spec.Instances[set].Churn
	if ch == nil || ch.BirthsPerHour <= 0 {
		return
	}
	rng := w.substream(birthStream(set))
	next := startNS + int64(rng.Exp(3600*secondsPerNS/ch.BirthsPerHour))
	w.schedule(kindBirth, "", "", next, set)
}

func birthStream(set int) string {
	if set == 0 {
		return "churn/births"
	}
	return "churn/births/" + strconv.Itoa(set)
}

func (w *World) schedule(kind eventKind, entity, channel string, atNS int64, payload any) {
	if atNS < 0 {
		atNS = 0
	}
	w.tiebreak++
	heap.Push(&w.queue, &item{
		timeNS:   atNS,
		tiebreak: w.tiebreak,
		kind:     kind,
		entity:   entity,
		channel:  channel,
		payload:  payload,
	})
}

func entityTemplateValue(params map[string]any, name string) string {
	if params != nil {
		if value, ok := params[name]; ok {
			if text, ok := value.(string); ok {
				return text
			}
			return fmt.Sprint(value)
		}
	}
	return "a"
}

func (w *World) newEntity(id string, set int, at int64) *Entity {
	kind := w.Spec.InstanceKind(set)
	return &Entity{ID: id, Type: kind.Name, BornNS: at, kind: kind, set: set, states: map[string]*stateValue{}, channels: map[string]*channelRunState{}, alive: true}
}

func (w *World) initializeEntityState(ent *Entity, at int64) {
	for i := range ent.kind.State {
		state := &ent.kind.State[i]
		ent.states[state.Name] = &stateValue{dyn: ent.kind.DynamicsOf(state.Name), decl: state, x: state.Initial, lastStep: at}
	}
	ent.bindSteps()
}

func (ent *Entity) bindSteps() {
	ent.steps = ent.steps[:0:0]
	for i := range ent.kind.Dynamics {
		dyn := &ent.kind.Dynamics[i]
		if s := ent.states[dyn.Target]; dyn.Integrated() && s != nil {
			ent.steps = append(ent.steps, step{dyn: dyn, state: s})
		}
	}
}

func (w *World) initializeEntityChannels(ent *Entity, id string, at int64) {
	for i := range ent.kind.Channels {
		channel := &ent.kind.Channels[i]
		ent.channels[channel.Name] = &channelRunState{lastTrigger: w.stateAt(id, channel.Cadence.TriggerState, at)}
		w.scheduleNextEmission(id, channel.Name, channel, at)
	}
}
