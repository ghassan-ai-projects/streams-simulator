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

func twoKindSpec(t *testing.T) *domain.Compiled {
	t.Helper()
	spec, err := domain.Load(testsupport.TwoKindDomain())
	if err != nil {
		t.Fatal(err)
	}
	return spec
}

func twoKindWorld(t *testing.T, spec *domain.Compiled, ids ...string) *World {
	t.Helper()
	w, err := New(spec, 7, "w-two", model.DefaultStartTimeNS, Options{InitialEntities: ids, EmitDisabled: true})
	if err != nil {
		t.Fatal(err)
	}
	return w
}

func lagClosed(t, tau, gain, input, x0 float64) float64 {
	return gain*input + (x0-gain*input)*math.Exp(-t/tau)
}

func TestTwoKindWorldMatchesTheClosedFormOfEachKind(t *testing.T) {
	t.Parallel()
	w := twoKindWorld(t, twoKindSpec(t))
	for _, seconds := range []float64{60, 600, 1800, 7200} {
		at := model.DefaultStartTimeNS + int64(seconds*secondsPerNS)
		tank := lagClosed(seconds, 600, 1, 3, 2)
		pump := lagClosed(seconds, 300, 0.5, 2*4, 0)
		if got := w.StateValue("tank-2", "level_true", at); math.Abs(got-tank) > 1e-6 {
			t.Fatalf("tank level at %vs = %v, closed form %v", seconds, got, tank)
		}
		if got := w.StateValue("pump-3", "flow_true", at); math.Abs(got-pump) > 1e-6 {
			t.Fatalf("pump flow at %vs = %v, closed form %v", seconds, got, pump)
		}
	}
}

func TestTwoKindFaultsFollowTheirOwnEnvelopes(t *testing.T) {
	t.Parallel()
	w := twoKindWorld(t, twoKindSpec(t))
	start := model.DefaultStartTimeNS
	if _, err := w.InjectFault("tank-1", "tank_leak", start, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := w.InjectFault("pump-1", "pump_wear", start, nil); err != nil {
		t.Fatal(err)
	}
	for _, seconds := range []float64{600, 1800, 5400} {
		hours := seconds / 3600
		at := start + int64(seconds*secondsPerNS)
		tank := lagClosed(seconds, 600, 1, 3, 2) - 1.0*math.Min(1.0*hours, 1)
		pump := lagClosed(seconds, 300, 0.5, 8, 0) - 2.0*math.Min(2.0*hours, 1)
		if got := w.StateValue("tank-1", "level_true", at); math.Abs(got-tank) > 1e-6 {
			t.Fatalf("leaking tank at %vs = %v, closed form %v", seconds, got, tank)
		}
		if got := w.StateValue("pump-1", "flow_true", at); math.Abs(got-pump) > 1e-6 {
			t.Fatalf("worn pump at %vs = %v, closed form %v", seconds, got, pump)
		}
	}
}

func TestAFaultOrEffectorReachesOnlyEntitiesOfItsKind(t *testing.T) {
	t.Parallel()
	w := twoKindWorld(t, twoKindSpec(t))
	start := model.DefaultStartTimeNS
	if _, err := w.InjectFault("pump-1", "tank_leak", start, nil); err == nil || !strings.Contains(err.Error(), `kind "pump"`) {
		t.Fatalf("a tank fault on a pump: %v, want a refusal naming the pump kind", err)
	}
	if _, err := w.InjectFault("tank-1", "pump_wear", start, nil); err == nil || !strings.Contains(err.Error(), `kind "tank"`) {
		t.Fatalf("a pump fault on a tank: %v, want a refusal naming the tank kind", err)
	}
	args := map[string]any{"tank_id": "pump-1"}
	if _, err := w.InvokeEffector("refill", "pump-1", "c-1", args, start); !errors.Is(err, ErrInvalidInvocation) {
		t.Fatalf("a tank effector on a pump: %v, want ErrInvalidInvocation", err)
	}
	if _, err := w.InvokeEffector("refill", "tank-1", "c-2", map[string]any{"tank_id": "tank-1"}, start); err != nil {
		t.Fatalf("a tank effector on a tank: %v", err)
	}
}

func TestEveryEntityEmitsOnlyTheChannelsOfItsKind(t *testing.T) {
	t.Parallel()
	w, err := New(twoKindSpec(t), 7, "w-two", model.DefaultStartTimeNS, Options{})
	if err != nil {
		t.Fatal(err)
	}
	counts := map[string]int{}
	w.SetEmitter(func(e model.SimEvent) { counts[fmt.Sprintf("%s|%s|%s", e.EntityType, e.EntityID, e.Channel)]++ })
	if _, _, err := w.Advance(model.DefaultStartTimeNS + 3600*secondsPerNS); err != nil {
		t.Fatal(err)
	}
	if len(counts) != 7 {
		t.Fatalf("streams = %v, want 3 tanks and 4 pumps", counts)
	}
	for key, n := range counts {
		want := map[bool]int{true: 60, false: 120}[key[:4] == "tank"]
		if (key[:4] != "tank" && key[:4] != "pump") || n < want-1 || n > want+1 {
			t.Fatalf("%s emitted %d events, want about %d", key, n, want)
		}
	}
	if counts["tank|tank-1|tank.level"] == 0 || counts["pump|pump-4|pump.flow"] == 0 {
		t.Fatalf("streams = %v", counts)
	}
}

func TestPeersAreTheEntitiesOfTheSameKind(t *testing.T) {
	t.Parallel()
	w := twoKindWorld(t, twoKindSpec(t))
	if got := w.PeerIDs("tank-1"); !slices.Equal(got, []string{"tank-1", "tank-2", "tank-3"}) {
		t.Fatalf("tank peers = %v", got)
	}
	if got := w.PeerIDs("pump-2"); len(got) != 4 || got[0] != "pump-1" {
		t.Fatalf("pump peers = %v", got)
	}
}

func TestEachInstanceSetBirthsEntitiesOfItsOwnKind(t *testing.T) {
	t.Parallel()
	spec := twoKindSpec(t)
	spec.Spec.Instances[1].Churn = &model.Churn{BirthsPerHour: 3600, MeanLifetimeS: 1e6}
	w := twoKindWorld(t, spec)
	if _, _, err := w.Advance(model.DefaultStartTimeNS + 600*secondsPerNS); err != nil {
		t.Fatal(err)
	}
	born := w.EntityIDs()[7:]
	if len(born) < 100 || born[0] != "pump-5" || w.Entity(born[0]).Type != "pump" {
		t.Fatalf("born = %v, want pumps numbered after the initial four", born)
	}
	if len(w.PeerIDs("tank-1")) != 3 {
		t.Fatalf("tank peers = %v, want the three initial tanks", w.PeerIDs("tank-1"))
	}
}

func TestEntitiesOfOneKindDrawTheSameStreamWhetherOrNotAnotherKindExists(t *testing.T) {
	t.Parallel()
	spec := twoKindSpec(t)
	both := collectStreams(t, spec)
	tanks := collectStreams(t, spec, "tank-1", "tank-2", "tank-3")
	pumps := collectStreams(t, spec, "pump-1", "pump-2", "pump-3", "pump-4")
	for _, stream := range [][]string{tanks, pumps} {
		for _, line := range stream {
			if !slices.Contains(both, line) {
				t.Fatalf("event %s of the single-kind world is missing from the two-kind world", line)
			}
		}
	}
	if len(both) != len(tanks)+len(pumps) {
		t.Fatalf("two-kind world holds %d events, single-kind worlds %d + %d", len(both), len(tanks), len(pumps))
	}
}

func collectStreams(t *testing.T, spec *domain.Compiled, ids ...string) []string {
	t.Helper()
	w, err := New(spec, 7, "w-two", model.DefaultStartTimeNS, Options{InitialEntities: ids})
	if err != nil {
		t.Fatal(err)
	}
	var lines []string
	w.SetEmitter(func(e model.SimEvent) {
		lines = append(lines, fmt.Sprintf("%s|%s|%s|%v", e.EntityID, e.Channel, e.EventTime, e.Value))
	})
	if _, _, err := w.Advance(model.DefaultStartTimeNS + 1800*secondsPerNS); err != nil {
		t.Fatal(err)
	}
	return lines
}

func TestAddEntityPlacesAnIdInItsInstanceSetAndRefusesOneThatFitsNone(t *testing.T) {
	t.Parallel()
	w := twoKindWorld(t, twoKindSpec(t))
	at := model.DefaultStartTimeNS
	if err := w.AddEntity("pump-9", at); err != nil || w.Entity("pump-9").Type != "pump" {
		t.Fatalf("AddEntity(pump-9) = %v, want a pump", err)
	}
	if err := w.AddEntity("valve-9", at); err == nil || !strings.Contains(err.Error(), `"valve-9" fits no instance set`) {
		t.Fatalf("AddEntity(valve-9) = %v, want a refusal naming the id", err)
	}
	if w.Entity("valve-9") != nil {
		t.Fatal("a refused id became an entity")
	}
}
