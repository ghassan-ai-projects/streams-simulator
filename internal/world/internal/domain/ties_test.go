package domain

import "testing"

func TestAtOneInstantTheLaterIssuedAssignmentWins(t *testing.T) {
	t.Parallel()
	w := followUpWorld(t)
	call(t, w, "run_fan", "first", 100, map[string]any{"speed": 5.0})
	call(t, w, "run_fan", "second", 100, map[string]any{"speed": 7.0})
	if got := w.StateValue("e-1", "fan", at(150)); got != 7 {
		t.Fatalf("fan = %v, want the later set-point 7", got)
	}
}

func TestALeaseRenewedAtItsLapseInstantHolds(t *testing.T) {
	t.Parallel()
	w := followUpWorld(t)
	for i, seconds := range []float64{100, 160, 220, 280} {
		call(t, w, "run_fan", string(rune('a'+i)), seconds, map[string]any{"speed": 5.0, "lease_s": 60.0})
	}
	for _, seconds := range []float64{159, 160, 161, 221, 281, 339} {
		if got := w.StateValue("e-1", "fan", at(seconds)); got != 5 {
			t.Fatalf("fan at %v s = %v, want the renewed 5", seconds, got)
		}
	}
	if got := w.StateValue("e-1", "fan", at(341)); got != 0 {
		t.Fatalf("fan after the last lease = %v, want 0", got)
	}
}
