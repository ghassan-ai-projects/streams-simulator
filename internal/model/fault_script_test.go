package model

import "testing"

func TestAFaultIsCommandedWhenACommandOfTheKindSetsAStateItActsOn(t *testing.T) {
	t.Parallel()
	kind := &Kind{
		Effectors: []Effector{{Name: "set", Effect: Effect{StateDeltas: []StateDelta{{State: "load"}}}}},
		Faults: []Fault{
			{ID: "on_commanded", Affects: []FaultEffect{{State: "load"}}},
			{ID: "on_free", Affects: []FaultEffect{{State: "wear"}}},
			{ID: "both", Affects: []FaultEffect{{State: "wear"}, {State: "load"}}},
			{ID: "none"},
		},
	}
	for id, want := range map[string]bool{"on_commanded": true, "on_free": false, "both": true, "none": false} {
		if got := kind.FaultCommanded(kind.FaultDecl(id)); got != want {
			t.Errorf("FaultCommanded(%s) = %v, want %v", id, got, want)
		}
	}
}

func TestAFaultIsPulseWhenItsOnsetIsAPulse(t *testing.T) {
	t.Parallel()
	pulse, ramp := Fault{Onset: FaultOnset{Shape: ShapePulse}}, Fault{Onset: FaultOnset{Shape: "ramp"}}
	if !pulse.IsPulse() || ramp.IsPulse() {
		t.Fatalf("IsPulse: pulse %v, ramp %v", pulse.IsPulse(), ramp.IsPulse())
	}
}
