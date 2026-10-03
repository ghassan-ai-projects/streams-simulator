package world

import (
	"container/heap"
	"fmt"
	"strconv"
	"strings"
)

// RenderID renders one entity id from the domain's id template. Exported
// for hosts that must enumerate the domain's entity ids without building a
// world (the audit path).
func RenderID(tmpl string, n int) string {
	return renderID(tmpl, n, nil)
}

// renderID expands the id template.
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

// addEntity creates an entity with initial hidden state and schedules its
// first emissions.
func (w *World) addEntity(id string, atNS int64, params map[string]any) error {
	if _, exists := w.entities[id]; exists {
		return fmt.Errorf("world: entity %q already exists", id)
	}
	ent := w.newEntity(id, atNS)
	w.initializeEntityState(ent, atNS)
	w.entities[id] = ent
	w.entityOrder = append(w.entityOrder, id)
	w.initializeEntityChannels(ent, id, atNS)
	return nil
}

// scheduleChurn seeds autonomous entity births.
func (w *World) scheduleChurn(startNS int64) {
	ch := w.Spec.Spec.Entities.Churn
	if ch == nil || ch.BirthsPerHour <= 0 {
		return
	}
	rng := w.substream("churn/births")
	next := startNS + int64(rng.Exp(3600*secondsPerNS/ch.BirthsPerHour))
	w.schedule(kindBirth, "", "", next, nil)
}

// Schedule pushes an event onto the DES queue with a monotonic tiebreak.
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

func (w *World) newEntity(id string, at int64) *Entity {
	kind := w.Spec.Spec.Entities.EntityType
	if kind == "" {
		kind = strings.Split(w.Spec.Spec.ID, "-")[0]
	}
	return &Entity{ID: id, Type: kind, BornNS: at, States: map[string]*stateValue{}, channels: map[string]*channelRunState{}, alive: true}
}

func (w *World) initializeEntityState(ent *Entity, at int64) {
	for i := range w.Spec.Spec.State {
		state := &w.Spec.Spec.State[i]
		ent.States[state.Name] = &stateValue{x: state.Initial, lastStep: at}
	}
}

func (w *World) initializeEntityChannels(ent *Entity, id string, at int64) {
	for i := range w.Spec.Spec.Channels {
		channel := &w.Spec.Spec.Channels[i]
		ent.channels[channel.Name] = &channelRunState{lastTrigger: w.stateAt(id, channel.Cadence.TriggerState, at)}
		w.scheduleNextEmission(id, channel.Name, channel, at)
	}
}
