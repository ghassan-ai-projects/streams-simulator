package domain

import (
	"encoding/json"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func wantInterval(t *testing.T, got model.FaultInterval, id, fault string, injected, onset int64, cleared *int64, by string) {
	t.Helper()
	if got.FaultID != id || got.EntityID != "e-1" || got.Fault != fault || got.InjectedAtNS != injected || got.OnsetNS != onset {
		t.Fatalf("interval = %+v, want %s of %s injected at %d with onset %d", got, id, fault, injected, onset)
	}
	switch {
	case cleared == nil && (got.ClearedNS != nil || got.ClearedBy != nil):
		t.Fatalf("interval = %+v, want it still active", got)
	case cleared != nil && (got.ClearedNS == nil || *got.ClearedNS != *cleared || got.ClearedBy == nil || *got.ClearedBy != by):
		t.Fatalf("interval = %+v, want cleared at %d by %s", got, *cleared, by)
	}
}

func instant(ns int64) *int64 { return &ns }

func TestAFaultNeverClearedIsAnOpenIntervalWithNullEnds(t *testing.T) {
	t.Parallel()
	w := constWorld(t, nil, model.DefaultStartTimeNS)
	if _, err := w.InjectFault("e-1", "f1", at(10), nil); err != nil {
		t.Fatal(err)
	}
	if _, _, err := w.Advance(at(60)); err != nil {
		t.Fatal(err)
	}
	for _, intervals := range [][]model.FaultInterval{w.FaultIntervals(), w.FinalFaultIntervals()} {
		if len(intervals) != 1 {
			t.Fatalf("intervals = %+v, want one", intervals)
		}
		wantInterval(t, intervals[0], "f-0", "f1", at(0), at(10), nil, "")
	}
	raw, err := json.Marshal(w.FaultIntervals()[0])
	if err != nil || string(raw) != `{"fault_id":"f-0","entity_id":"e-1","fault":"f1","injected_at_ns":`+itoa(at(0))+`,"onset_ns":`+itoa(at(10))+`,"cleared_ns":null,"cleared_by":null}` {
		t.Fatalf("json = %s (%v), want explicit nulls", raw, err)
	}
}

func itoa(ns int64) string {
	raw, _ := json.Marshal(ns)
	return string(raw)
}

func TestAClearCommandEndsTheFaultAtItsAtNSNotAtTheCommandTime(t *testing.T) {
	t.Parallel()
	w := constWorld(t, nil, model.DefaultStartTimeNS)
	id, err := w.InjectFault("e-1", "f1", at(10), nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := w.Advance(at(20)); err != nil {
		t.Fatal(err)
	}
	if err := w.ClearFault(id, at(50)); err != nil {
		t.Fatal(err)
	}
	wantInterval(t, w.FaultIntervals()[0], id, "f1", at(0), at(10), instant(at(50)), model.ClearedByCommand)
	wantInterval(t, w.FinalFaultIntervals()[0], id, "f1", at(0), at(10), nil, "")
	if _, _, err := w.Advance(at(80)); err != nil {
		t.Fatal(err)
	}
	wantInterval(t, w.FinalFaultIntervals()[0], id, "f1", at(0), at(10), instant(at(50)), model.ClearedByCommand)
}

func TestAClearCommandWithoutATimeEndsTheFaultAtTheClock(t *testing.T) {
	t.Parallel()
	w := constWorld(t, nil, model.DefaultStartTimeNS)
	id, _ := w.InjectFault("e-1", "f1", at(10), nil)
	if _, _, err := w.Advance(at(30)); err != nil {
		t.Fatal(err)
	}
	if err := w.ClearFault(id, 0); err != nil {
		t.Fatal(err)
	}
	wantInterval(t, w.FinalFaultIntervals()[0], id, "f1", at(0), at(10), instant(at(30)), model.ClearedByCommand)
}

func TestAPulseEndsItselfAtOnsetPlusItsDuration(t *testing.T) {
	t.Parallel()
	w := constWorld(t, sagSpec(t, false), model.DefaultStartTimeNS)
	id, err := w.InjectFault("e-1", "long_sag", at(600.4), nil)
	if err != nil {
		t.Fatal(err)
	}
	wantInterval(t, w.FaultIntervals()[0], id, "long_sag", at(0), at(600.4), instant(at(600.7)), model.ClearedByPulseEnd)
	wantInterval(t, w.FinalFaultIntervals()[0], id, "long_sag", at(0), at(600.4), nil, "")
	if _, _, err := w.Advance(at(700)); err != nil {
		t.Fatal(err)
	}
	wantInterval(t, w.FinalFaultIntervals()[0], id, "long_sag", at(0), at(600.4), instant(at(600.7)), model.ClearedByPulseEnd)
}

func TestAClearCommandThatShortensAPulseTakesTheEndAndALaterOneDoesNot(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name  string
		clear float64
		end   float64
		by    string
	}{
		{"earlier than the pulse end", 600.5, 600.5, model.ClearedByCommand},
		{"after the pulse end", 650, 600.7, model.ClearedByPulseEnd},
		{"at the pulse end", 600.7, 600.7, model.ClearedByPulseEnd},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			w := constWorld(t, sagSpec(t, false), model.DefaultStartTimeNS)
			id, _ := w.InjectFault("e-1", "long_sag", at(600.4), nil)
			if err := w.ClearFault(id, at(tc.clear)); err != nil {
				t.Fatal(err)
			}
			wantInterval(t, w.FaultIntervals()[0], id, "long_sag", at(0), at(600.4), instant(at(tc.end)), tc.by)
		})
	}
}

func TestAResponderEndsTheFaultItWasToldAboutAtItsArrival(t *testing.T) {
	t.Parallel()
	w := followUpWorld(t)
	if _, err := w.InjectFault("e-1", "f1", at(0), nil); err != nil {
		t.Fatal(err)
	}
	call(t, w, "work_order", "miss", 100, map[string]any{"cause": "pump_wear"})
	wantInterval(t, w.FaultIntervals()[0], "f-0", "f1", at(0), at(0), nil, "")
	call(t, w, "work_order", "match", 200, map[string]any{"cause": "f1"})
	wantInterval(t, w.FaultIntervals()[0], "f-0", "f1", at(0), at(0), instant(at(1100)), model.ClearedByResponder)
	if _, _, err := w.Advance(at(1200)); err != nil {
		t.Fatal(err)
	}
	wantInterval(t, w.FinalFaultIntervals()[0], "f-0", "f1", at(0), at(0), instant(at(1100)), model.ClearedByResponder)
}

func TestFaultIntervalsListEveryFaultInInjectionOrder(t *testing.T) {
	t.Parallel()
	w := constWorld(t, sagSpec(t, false), model.DefaultStartTimeNS)
	first, _ := w.InjectFault("e-1", "long_sag", at(5), nil)
	second, _ := w.InjectFault("e-1", "short_sag", at(7), nil)
	got := w.FaultIntervals()
	if len(got) != 2 || got[0].FaultID != first || got[1].FaultID != second || got[1].Fault != "short_sag" {
		t.Fatalf("intervals = %+v, want %s then %s", got, first, second)
	}
}
