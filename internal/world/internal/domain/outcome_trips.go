package domain

import "github.com/ghassan-ai-projects/streams-simulator/internal/model"

type tripSpan struct {
	entries   int
	firstNS   int64
	trippedNS int64
}

func isTripped(f1 *model.F1Dyn, value float64) bool { return value == f1.TrippedValue }

func (w *World) traceTrip(s *stateValue, f1 *model.F1Dyn, previous float64, visibleNS int64) {
	if w.outcome == nil || isTripped(f1, previous) == isTripped(f1, s.x) {
		return
	}
	marks := len(s.tripMarks)
	s.tripMarks = append(s.tripMarks[:marks:marks], visibleNS)
}

func (s *stateValue) tripSpanAt(t int64) tripSpan {
	var span tripSpan
	for i := 0; i < len(s.tripMarks) && s.tripMarks[i] <= t; i += 2 {
		end := t
		if i+1 < len(s.tripMarks) && s.tripMarks[i+1] <= t {
			end = s.tripMarks[i+1]
		}
		span.entries, span.trippedNS = span.entries+1, span.trippedNS+end-s.tripMarks[i]
	}
	if span.entries > 0 {
		span.firstNS = s.tripMarks[0]
	}
	return span
}
