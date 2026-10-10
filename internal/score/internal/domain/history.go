package domain

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func hasOutcome(outcomes []string, wanted string) bool {
	for _, outcome := range outcomes {
		if outcome == wanted {
			return true
		}
	}
	return false
}

func historyValue(ev Evidence, entity, state string, atNS int64) (float64, bool) {
	var before historySample
	// Prefer the first sample at/after the instant, otherwise the latest before it.
	for _, snap := range ev.History {
		value, ok := entityStateValue(snap.Entity, entity, snap.States, state)
		if ok {
			if snap.TimeNS >= atNS {
				return value, true
			}
			before.retainLatest(value, snap.TimeNS)
		}
	}
	return before.value, before.found
}

// historyValueBefore returns the latest captured sample strictly before t.
// A pre-onset baseline must not read a sample at or after the onset: when
// the fault lands on an emission boundary, that sample already carries the
// fault and the deviation collapses to zero.
func historyValueBefore(ev Evidence, entity, state string, atNS int64) (float64, bool) {
	var before historySample
	for _, snap := range ev.History {
		value, ok := entityStateValue(snap.Entity, entity, snap.States, state)
		if ok && snap.TimeNS < atNS {
			before.retainLatest(value, snap.TimeNS)
		}
	}
	return before.value, before.found
}

func faultFor(ev Evidence, label string) *model.Fault {
	for i := range ev.Domain.Spec.Faults {
		if ev.Domain.Spec.Faults[i].ID == label {
			return &ev.Domain.Spec.Faults[i]
		}
	}
	return nil
}

type historySample struct {
	value float64
	atNS  int64
	found bool
}

func (sample *historySample) retainLatest(value float64, atNS int64) {
	if !sample.found || atNS > sample.atNS {
		sample.value, sample.atNS, sample.found = value, atNS, true
	}
}

func entityStateValue(sampleEntity, entity string, states map[string]float64, state string) (float64, bool) {
	if sampleEntity != entity {
		return 0, false
	}
	value, ok := states[state]
	return value, ok
}
