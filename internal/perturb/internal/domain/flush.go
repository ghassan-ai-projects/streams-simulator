package domain

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// Flush returns records held by windowing perturbations (reorder windows,
// producer flaps, gross backfill) at the end of an advance. Callers must
// process them exactly once, at the boundary where the perturbation's
// window closes.
func (l *Layer) Flush(atNS int64) []Delivered {
	var out []Delivered
	for _, id := range l.activeIDs() {
		out = append(out, l.active[id].flushRecords(atNS)...)
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
	for _, n := range catalogNames {
		if n == name {
			return true
		}
	}
	return false
}

func paramFloat(params map[string]any, key string, def float64) float64 {
	if value, ok := asFloat(params[key]); ok {
		return value
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

func (a *applied) flushRecords(atNS int64) []Delivered {
	switch a.Name {
	case Reorder:
		return a.flushReorder(atNS)
	case ProducerFlap:
		return a.flushBuffered(atNS, true)
	case GrossBackfill:
		return a.flushBuffered(atNS, false)
	}
	return nil
}

func (a *applied) flushReorder(atNS int64) []Delivered {
	flushed := a.window.flush(atNS)
	for i := range flushed {
		if flushed[i].Delivered && flushed[i].Reason == model.DeliveryOK {
			flushed[i].Reason = model.DeliveryReordered
		}
	}
	return flushed
}

func (a *applied) flushBuffered(atNS int64, birth bool) []Delivered {
	if len(a.buffer) == 0 || a.UntilNS <= 0 || atNS < a.UntilNS {
		return nil
	}
	var out []Delivered
	for _, record := range a.buffer {
		out = append(out, recoveredRecord(record, atNS, birth))
	}
	a.buffer = nil
	return out
}

func recoveredRecord(record Delivered, atNS int64, birth bool) Delivered {
	// Producer recovery republishes a birth burst; backfill retains birth flags.
	event := record.Event
	if birth {
		event.Birth = true
	}
	event.ObservedTime = model.FormatTime(atNS)
	record.Event = event
	record.Reason = model.DeliveryDelayed
	return record
}
