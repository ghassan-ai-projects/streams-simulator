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
	out := tmpl
	out = strings.ReplaceAll(out, "{n}", strconv.Itoa(n))
	for {
		start := strings.Index(out, "{")
		end := strings.Index(out, "}")
		if start < 0 || end < 0 || end < start {
			break
		}
		name := out[start+1 : end]
		val := "a"
		if params != nil {
			if v, ok := params[name]; ok {
				switch x := v.(type) {
				case string:
					val = x
				default:
					val = fmt.Sprint(x)
				}
			}
		}
		out = out[:start] + val + out[end+1:]
	}
	return out
}

// addEntity creates an entity with initial hidden state and schedules its
// first emissions.
func (w *World) addEntity(id string, atNS int64, params map[string]any) error {
	if _, exists := w.entities[id]; exists {
		return fmt.Errorf("world: entity %q already exists", id)
	}
	et := w.Spec.Spec.Entities.EntityType
	if et == "" {
		et = strings.Split(w.Spec.Spec.ID, "-")[0]
	}
	ent := &Entity{
		ID:       id,
		Type:     et,
		BornNS:   atNS,
		States:   map[string]*stateValue{},
		channels: map[string]*channelRunState{},
		alive:    true,
	}
	for i := range w.Spec.Spec.State {
		st := &w.Spec.Spec.State[i]
		ent.States[st.Name] = &stateValue{
			x:        st.Initial,
			lastStep: atNS,
		}
	}
	w.entities[id] = ent
	w.entityOrder = append(w.entityOrder, id)

	// Generate any template ids from params for this entity.
	if tmpl := w.Spec.Spec.Entities.IDTemplate; tmpl != "" {
		// re-render a fresh n if this is an autonomous birth
		_ = tmpl
	}
	_ = params

	for i := range w.Spec.Spec.Channels {
		ch := &w.Spec.Spec.Channels[i]
		cs := &channelRunState{lastTrigger: w.stateAt(id, ch.Cadence.TriggerState, atNS)}
		ent.channels[ch.Name] = cs
		next := w.nextEmission(id, ch, atNS)
		if next > 0 {
			w.schedule(kindEmission, id, ch.Name, next, nil)
		}
	}
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
