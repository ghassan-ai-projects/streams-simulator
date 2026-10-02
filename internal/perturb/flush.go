package perturb

import (
	"encoding/json"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// Flush returns records held by windowing perturbations (reorder windows,
// producer flaps, gross backfill) at the end of an advance. Callers must
// process them exactly once, at the boundary where the perturbation's
// window closes.
func (l *Layer) Flush(atNS int64) []Delivered {
	var out []Delivered
	for _, id := range l.ActiveIDs() {
		a := l.active[id]
		switch a.Name {
		case Reorder:
			flushed := a.window.flush(atNS)
			for i := range flushed {
				if flushed[i].Delivered && flushed[i].Reason == model.DeliveryOK {
					flushed[i].Reason = model.DeliveryReordered
				}
			}
			out = append(out, flushed...)
		case ProducerFlap:
			// Birth burst: republish every held event at one observed time,
			// with event times spread across the outage.
			if len(a.buffer) > 0 && a.UntilNS > 0 && atNS >= a.UntilNS {
				recovery := atNS
				for i := range a.buffer {
					r := a.buffer[i]
					ev := r.Event
					ev.Birth = true
					ev.ObservedTime = model.FormatTime(recovery)
					r.Event = ev
					r.Reason = model.DeliveryDelayed
					out = append(out, r)
				}
				a.buffer = nil
			}
		case GrossBackfill:
			if len(a.buffer) > 0 && a.UntilNS > 0 && atNS >= a.UntilNS {
				recovery := atNS
				for i := range a.buffer {
					r := a.buffer[i]
					ev := r.Event
					ev.ObservedTime = model.FormatTime(recovery)
					r.Event = ev
					r.Reason = model.DeliveryDelayed
					out = append(out, r)
				}
				a.buffer = nil
			}
		}
	}
	if len(l.pending) > 0 {
		out = append(l.pending, out...)
		l.pending = nil
	}
	return out
}

func mapRecs(recs []Delivered, f func(Delivered) []Delivered) []Delivered {
	var out []Delivered
	for _, r := range recs {
		out = append(out, f(r)...)
	}
	return out
}

func isName(name string) bool {
	for _, n := range Names {
		if n == name {
			return true
		}
	}
	return false
}

func paramFloat(params map[string]any, key string, def float64) float64 {
	if params == nil {
		return def
	}
	switch v := params[key].(type) {
	case float64:
		return v
	case int64:
		return float64(v)
	case int:
		return float64(v)
	case json.Number:
		f, err := v.Float64()
		if err == nil {
			return f
		}
	}
	return def
}

func paramInt(params map[string]any, key string, def int) int {
	return int(paramFloat(params, key, float64(def)))
}

func paramStr(params map[string]any, key, def string) string {
	if params == nil {
		return def
	}
	if s, ok := params[key].(string); ok {
		return s
	}
	return def
}
