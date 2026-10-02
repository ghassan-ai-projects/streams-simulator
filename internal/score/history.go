package score

import (
	"encoding/json"
	"fmt"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

func hasOutcome(outcomes []string, wanted string) bool {
	for _, outcome := range outcomes {
		if outcome == wanted {
			return true
		}
	}
	return false
}

func historyValue(r *run.Run, entity, state string, atNS int64) (float64, bool) {
	history := r.History()
	if len(history) == 0 {
		return 0, false
	}
	// Prefer the first sample at/after the requested instant, otherwise the
	// latest sample before it. This is random-access over captured evidence;
	// it never advances the mutable world during scoring.
	var before *float64
	var beforeAt int64
	for _, snap := range history {
		if snap.Entity != entity {
			continue
		}
		value, ok := snap.States[state]
		if !ok {
			continue
		}
		if snap.TimeNS >= atNS {
			return value, true
		}
		if before == nil || snap.TimeNS > beforeAt {
			v := value
			before = &v
			beforeAt = snap.TimeNS
		}
	}
	if before != nil {
		return *before, true
	}
	return 0, false
}

// historyValueBefore returns the latest captured sample strictly before t.
// A pre-onset baseline must not read a sample at or after the onset: when
// the fault lands on an emission boundary, that sample already carries the
// fault and the deviation collapses to zero.
func historyValueBefore(r *run.Run, entity, state string, atNS int64) (float64, bool) {
	history := r.History()
	var best *float64
	var bestAt int64
	for _, snap := range history {
		if snap.Entity != entity {
			continue
		}
		value, ok := snap.States[state]
		if !ok {
			continue
		}
		if snap.TimeNS < atNS && (best == nil || snap.TimeNS > bestAt) {
			v := value
			best = &v
			bestAt = snap.TimeNS
		}
	}
	if best != nil {
		return *best, true
	}
	return 0, false
}

func faultFor(r *run.Run, label string) *model.Fault {
	for i := range r.Domain().Spec.Faults {
		if r.Domain().Spec.Faults[i].ID == label {
			return &r.Domain().Spec.Faults[i]
		}
	}
	return nil
}

// Marshal returns the scorecard as indented JSON.
func (s *Scorecard) Marshal() ([]byte, error) {
	raw, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("score: marshal: %w", err)
	}
	return raw, nil
}
