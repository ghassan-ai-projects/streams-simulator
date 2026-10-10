package app

import (
	"encoding/json"
	"maps"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// History returns the world-state history (director-only).
func (r *Run) History() []model.StateSnapshot {
	out := make([]model.StateSnapshot, len(r.history))
	for i, rec := range r.history {
		out[i] = rec
		out[i].States = maps.Clone(rec.States)
	}
	return out
}

// RecordHistory snapshots hidden state for post-hoc analysis.
func (r *Run) RecordHistory(seq int64, atNS int64, entity string, states map[string]float64) {
	r.history = append(r.history, model.StateSnapshot{Seq: seq, TimeNS: atNS, Entity: entity, States: maps.Clone(states)})
}

func cloneVerdict(v *model.Verdict) *model.Verdict {
	if v == nil {
		return nil
	}
	raw, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	out := &model.Verdict{}
	if err := json.Unmarshal(raw, out); err != nil {
		return nil
	}
	return out
}
