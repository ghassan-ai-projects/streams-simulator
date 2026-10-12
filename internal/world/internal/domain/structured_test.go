package domain

import (
	"errors"
	"fmt"
	"math"
	"slices"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func structuredSpec(t *testing.T) *domain.Compiled {
	t.Helper()
	spec, err := domain.Load(testsupport.StructuredDomain())
	if err != nil {
		t.Fatal(err)
	}
	return spec
}

func structuredWorld(t *testing.T, opts Options) *World {
	t.Helper()
	w, err := New(structuredSpec(t), 7, "w-structured", model.DefaultStartTimeNS, opts)
	if err != nil {
		t.Fatal(err)
	}
	return w
}

func TestTheInitialPopulationListsNamesBeforeTemplatedIds(t *testing.T) {
	t.Parallel()
	w := structuredWorld(t, Options{EmitDisabled: true})
	want := []string{"tank-1", "tank-2", "T-main", "pump-1", "pump-2", "pump-3", "P-slow", "air_header"}
	if got := w.EntityIDs(); !slices.Equal(got, want) {
		t.Fatalf("entities = %v, want %v", got, want)
	}
	if w.Entity("air_header").Type != "header" || w.Entity("T-main").Type != "tank" {
		t.Fatal("a named instance did not take the kind of its set")
	}
	if got := w.PeerIDs("tank-1"); !slices.Equal(got, []string{"tank-1", "tank-2", "T-main"}) {
		t.Fatalf("tank peers = %v, want the named tank among them", got)
	}
}

func TestAnAddedStateFollowsItsClosedFormOnlyOnTheOverriddenInstance(t *testing.T) {
	t.Parallel()
	w := structuredWorld(t, Options{EmitDisabled: true})
	for _, seconds := range []float64{60, 900, 3600} {
		at := model.DefaultStartTimeNS + int64(seconds*secondsPerNS)
		want := lagClosed(seconds, 900, 1, 2*3, 20)
		if got := w.StateValue("T-main", "temp", at); math.Abs(got-want) > 1e-6 {
			t.Fatalf("temp at %vs = %v, closed form %v", seconds, got, want)
		}
		level := lagClosed(seconds, 600, 1, 3, 2)
		if got := w.StateValue("T-main", "level_true", at); math.Abs(got-level) > 1e-6 {
			t.Fatalf("level at %vs = %v, closed form %v", seconds, got, level)
		}
	}
	if !slices.Equal(w.StateNamesOf("tank-1"), []string{"inflow", "level_true"}) || !slices.Equal(w.StateNamesOf("T-main"), []string{"inflow", "level_true", "temp"}) {
		t.Fatalf("state names: tank %v named %v", w.StateNamesOf("tank-1"), w.StateNamesOf("T-main"))
	}
}

func TestANarrowedStateBoundClampsOnlyTheOverriddenInstance(t *testing.T) {
	t.Parallel()
	w := structuredWorld(t, Options{EmitDisabled: true})
	start := model.DefaultStartTimeNS
	for _, id := range []string{"tank-1", "T-main"} {
		if _, err := w.InjectFault(id, "tank_leak", start, map[string]any{"severity": 5.0}); err != nil {
			t.Fatal(err)
		}
	}
	at := start + 3600*secondsPerNS
	free := lagClosed(3600, 600, 1, 3, 2) - 5.0
	if got := w.StateValue("tank-1", "level_true", at); math.Abs(got-free) > 1e-6 || free >= 0.5 {
		t.Fatalf("unbounded tank = %v, closed form %v", got, free)
	}
	if got := w.StateValue("T-main", "level_true", at); got != 0.5 {
		t.Fatalf("narrowed tank = %v, want the narrowed minimum 0.5", got)
	}
}

func TestAnAddedChannelEmitsOnlyFromTheOverriddenInstance(t *testing.T) {
	t.Parallel()
	w := structuredWorld(t, Options{})
	counts := map[string]int{}
	w.SetEmitter(func(e model.SimEvent) { counts[e.EntityID+"|"+e.Channel]++ })
	if _, _, err := w.Advance(model.DefaultStartTimeNS + 3600*secondsPerNS); err != nil {
		t.Fatal(err)
	}
	if n := counts["T-main|tank.temp"]; n < 59 || n > 61 {
		t.Fatalf("added channel emitted %d events in an hour, want about 60", n)
	}
	if counts["tank-1|tank.temp"] != 0 || counts["T-main|tank.level"] == 0 || counts["air_header|header.pressure"] == 0 || counts["pump-1|header.pressure"] != 0 {
		t.Fatalf("streams = %v", counts)
	}
}

func TestAnAddedFaultAndEffectorReachOnlyTheOverriddenInstance(t *testing.T) {
	t.Parallel()
	w := structuredWorld(t, Options{EmitDisabled: true})
	start := model.DefaultStartTimeNS
	if _, err := w.InjectFault("T-main", "tank_overheat", start, nil); err != nil {
		t.Fatalf("added fault on its instance: %v", err)
	}
	if _, err := w.InjectFault("tank-1", "tank_overheat", start, nil); err == nil || !strings.Contains(err.Error(), "not declared by the kind") {
		t.Fatalf("added fault on a plain tank: %v", err)
	}
	args := map[string]any{"tank_id": "T-main"}
	if _, err := w.InvokeEffector("chill", "T-main", "c-1", args, start); err != nil {
		t.Fatalf("added effector on its instance: %v", err)
	}
	if _, err := w.InvokeEffector("chill", "tank-1", "c-2", map[string]any{"tank_id": "tank-1"}, start); !errors.Is(err, ErrInvalidInvocation) {
		t.Fatalf("added effector on a plain tank: %v, want ErrInvalidInvocation", err)
	}
}

func TestANarrowedArgumentBoundIsEnforcedOnlyForTheOverriddenInstance(t *testing.T) {
	t.Parallel()
	w := structuredWorld(t, Options{EmitDisabled: true})
	start := model.DefaultStartTimeNS
	call := func(id string, speed float64) error {
		args := map[string]any{"pump_id": id, "speed": speed}
		_, err := w.InvokeEffector("boost", id, fmt.Sprintf("c-%s-%v", id, speed), args, start)
		return err
	}
	if err := call("P-slow", 40); err != nil {
		t.Fatalf("speed 40 on P-slow: %v", err)
	}
	if err := call("P-slow", 60); !errors.Is(err, ErrInvalidInvocation) {
		t.Fatalf("speed 60 on P-slow: %v, want ErrInvalidInvocation", err)
	}
	if err := call("pump-1", 60); err != nil {
		t.Fatalf("speed 60 on pump-1: %v", err)
	}
	if w.EffectorOf("P-slow", "boost") == nil || w.EffectorOf("P-slow", "chill") != nil || w.EffectorOf("nobody", "boost") != nil {
		t.Fatal("EffectorOf did not follow the entity's kind")
	}
}

func TestAnAddedEffectorsConfirmationChannelReportsTheCounterfactualUnderSilentNoEffect(t *testing.T) {
	t.Parallel()
	w := structuredWorld(t, Options{EmitDisabled: true, ForceFailureMode: ModeSilentNoEffect})
	start := model.DefaultStartTimeNS
	if _, err := w.InvokeEffector("chill", "T-main", "c-1", map[string]any{"tank_id": "T-main"}, start); err != nil {
		t.Fatal(err)
	}
	at := start + 3600*secondsPerNS
	state := w.StateValue("T-main", "temp", at)
	if want := lagClosed(3600, 900, 1, 2*3, 20); math.Abs(state-want) > 1e-6 {
		t.Fatalf("temp = %v, closed form without the effect %v", state, want)
	}
	if gap := state - w.Reading("T-main", "tank.temp", at); math.Abs(gap-5) > 0.05 {
		t.Fatalf("confirmation channel sits %v below the state, want the counterfactual 5", gap)
	}
}
