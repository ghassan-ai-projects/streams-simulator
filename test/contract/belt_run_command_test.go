package contract

import (
	"errors"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

func TestAStartOrStopCarriesItsRunCommandAndAnotherValueIsRefused(t *testing.T) {
	t.Parallel()
	b := newBeltWorld(t, beltPlant(t, "a"))
	for _, call := range []struct {
		effector, entity string
		run              float64
	}{{"conveyor_drive__start_drive", "D-1", 0}, {"conveyor_drive__stop_drive", "D-1", 1}, {"belt_conveyor__stop_belt", "B-1", 1}, {"belt_conveyor__start_belt", "B-1", 0.5}} {
		if _, err := b.w.InvokeEffector(call.effector, call.entity, "wrong-"+call.effector, map[string]any{"run_command": call.run}, beltNS(0)); !errors.Is(err, world.ErrInvalidInvocation) {
			t.Errorf("%s with run_command %v = %v, want an invalid invocation", call.effector, call.run, err)
		}
	}
	if !b.invoke("belt_conveyor__stop_belt", "B-1", nil) {
		t.Fatal("stop_belt refused")
	}
	if b.advance(1); b.state("B-1", "run_cmd") != 0 {
		t.Fatalf("run_cmd after stop_belt = %v, want the stop to assign 0", b.state("B-1", "run_cmd"))
	}
	if !b.invoke("belt_conveyor__start_belt", "B-1", nil) {
		t.Fatal("start_belt refused")
	}
	if b.advance(2); b.state("B-1", "run_cmd") != 1 {
		t.Fatalf("run_cmd after start_belt = %v, want the start to assign 1", b.state("B-1", "run_cmd"))
	}
}
