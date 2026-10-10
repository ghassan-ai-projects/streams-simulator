package world_test

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

func newPondWorld(t *testing.T, opts world.Options) *world.World {
	t.Helper()
	spec, err := domain.Load(filepath.Join("..", "..", "domains", "aquaculture-pond.domain.json"))
	if err != nil {
		t.Fatal(err)
	}
	w, err := world.New(spec, 5, "w-contract", model.DefaultStartTimeNS, opts)
	if err != nil {
		t.Fatal(err)
	}
	return w
}

func TestNewRefusesAMissingSpec(t *testing.T) {
	t.Parallel()
	if _, err := world.New(nil, 1, "w", model.DefaultStartTimeNS, world.Options{}); !errors.Is(err, world.ErrNoSpec) {
		t.Fatalf("err = %v, want ErrNoSpec", err)
	}
}

func TestWorldExposesItsIdentityAndEntities(t *testing.T) {
	t.Parallel()
	w := newPondWorld(t, world.Options{})
	if w.ID != "w-contract" || w.StartNS != model.DefaultStartTimeNS || w.Spec == nil {
		t.Fatalf("identity = %q %d %v", w.ID, w.StartNS, w.Spec)
	}
	ids := w.EntityIDs()
	if len(ids) != 8 || ids[0] != world.RenderID(w.Spec.Spec.Entities.IDTemplate, 1) {
		t.Fatalf("entities = %v", ids)
	}
	if e := w.Entity(ids[0]); e == nil || e.ID != ids[0] || e.Type != "pond" {
		t.Fatalf("entity = %+v", e)
	}
	if w.Entity("no-such-entity") != nil {
		t.Fatal("an unknown entity must be nil")
	}
}

func TestAdvanceEmitsNativeEventsInOrderAndRefusesTimeTravel(t *testing.T) {
	t.Parallel()
	w := newPondWorld(t, world.Options{})
	var seqs []int64
	w.SetEmitter(func(ev model.SimEvent) { seqs = append(seqs, ev.Seq) })
	emitted, _, err := w.Advance(model.DefaultStartTimeNS + 3600e9)
	if err != nil {
		t.Fatal(err)
	}
	if emitted == 0 || int64(emitted) != w.EmittedCount() || len(seqs) != emitted {
		t.Fatalf("emitted %d, count %d, observed %d", emitted, w.EmittedCount(), len(seqs))
	}
	for i, seq := range seqs {
		if seq != int64(i) {
			t.Fatalf("event %d has seq %d", i, seq)
		}
	}
	if _, _, err := w.Advance(model.DefaultStartTimeNS); err == nil {
		t.Fatal("moving the clock backwards must be refused")
	}
	if w.Clock() != model.DefaultStartTimeNS+3600e9 {
		t.Fatalf("clock = %d", w.Clock())
	}
}

func TestFaultsAndEffectorsActThroughTheFacade(t *testing.T) {
	t.Parallel()
	w := newPondWorld(t, world.Options{ForceEffectorOK: true})
	pond := w.EntityIDs()[0]
	at := model.DefaultStartTimeNS + 60e9
	id, err := w.InjectFault(pond, "aerator_failure", at, nil)
	if err != nil {
		t.Fatal(err)
	}
	if faults := w.ListFaults(); len(faults) != 1 || faults[0].FaultID != id || w.ActiveFaultsCount() != 1 {
		t.Fatalf("faults = %+v", faults)
	}
	if err := w.ClearFault(id, at+60e9); err != nil {
		t.Fatal(err)
	}
	if _, _, err := w.Advance(at + 120e9); err != nil || w.ActiveFaultsCount() != 0 {
		t.Fatalf("after the clear time: %v, active %d", err, w.ActiveFaultsCount())
	}
	result, err := w.InvokeEffector("start_aerator", pond, "cmd-1", map[string]any{"pond_id": pond}, at)
	if err != nil || !result.Accepted || result.Mode != world.ModeOK {
		t.Fatalf("invoke: %v, %+v", err, result)
	}
	if calls := w.EffectorCalls(); len(calls) != 1 || calls[0].CommandID != "cmd-1" || calls[0].EntityID != pond {
		t.Fatalf("calls = %+v", calls)
	}
}

func TestEntitiesJoinAndLeaveAndTheirStateIsReadable(t *testing.T) {
	t.Parallel()
	w := newPondWorld(t, world.Options{})
	initial := w.InitialEntityIDs()
	if len(initial) != 8 || w.NextEventNS() == 0 {
		t.Fatalf("initial = %v, next event %d", initial, w.NextEventNS())
	}
	at := model.DefaultStartTimeNS + 60e9
	if err := w.AddEntity("site-a/pond-99", at); err != nil {
		t.Fatal(err)
	}
	if err := w.AddEntity("site-a/pond-99", at); err == nil {
		t.Fatal("adding an existing entity must fail")
	}
	if got := w.EntityIDs(); len(got) != 9 || len(w.InitialEntityIDs()) != 8 {
		t.Fatalf("after add: entities %d, initial %d", len(got), len(w.InitialEntityIDs()))
	}
	w.Retire("site-a/pond-99", "test", at)
	if len(w.EntityIDs()) != 8 {
		t.Fatalf("after retire: %v", w.EntityIDs())
	}
	pond := initial[0]
	if w.StateValue(pond, "dissolved_oxygen_true", at) <= 0 {
		t.Fatal("hidden state must be readable")
	}
	if w.Reading(pond, "pond.dissolved_oxygen", at) <= 0 {
		t.Fatal("noise-free reading must be positive")
	}
}

func TestFailureModeOverrideAndKickCountAreVisible(t *testing.T) {
	t.Parallel()
	w := newPondWorld(t, world.Options{})
	pond := w.EntityIDs()[0]
	w.SetFailureMode(world.ModeReject)
	at := model.DefaultStartTimeNS + 60e9
	result, err := w.InvokeEffector("start_aerator", pond, "cmd-2", map[string]any{"pond_id": pond}, at)
	if err != nil || result.Mode != world.ModeReject || result.Accepted {
		t.Fatalf("forced reject: %v, %+v", err, result)
	}
	w.SetFailureMode("")
	if _, err := w.InvokeEffector("start_aerator", pond, "cmd-3", map[string]any{"pond_id": pond}, at); err != nil {
		t.Fatal(err)
	}
	if w.PendingKicks() < 0 {
		t.Fatal("kick count must not be negative")
	}
}
