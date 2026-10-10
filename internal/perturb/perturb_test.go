package perturb_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/perturb"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func shippedLayer(t *testing.T) *perturb.Layer {
	t.Helper()
	spec, err := domain.Load(testsupport.Domain("rotating-machinery"))
	if err != nil {
		t.Fatal(err)
	}
	layer, err := perturb.New("w-1", 7, spec)
	if err != nil {
		t.Fatal(err)
	}
	return layer
}

func event(seq int64) model.SimEvent {
	at := model.FormatTime(model.DefaultStartTimeNS + seq*1e9)
	return model.SimEvent{Seq: seq, WorldID: "w-1", EntityType: "pump", EntityID: "site-a/pump-1",
		Channel: "motor_current", EventTime: at, ObservedTime: at, Value: 30.0, Unit: "A"}
}

func TestNamesIsTheCatalogAndAModifiableCopy(t *testing.T) {
	t.Parallel()
	names := perturb.Names()
	if len(names) == 0 || names[0] == "" {
		t.Fatalf("catalog = %v", names)
	}
	first := names[0]
	names[0] = "bogus"
	if again := perturb.Names(); again[0] != first {
		t.Fatalf("mutating the result changed the catalog: %v", again)
	}
	if _, err := shippedLayer(t).Apply("bogus", nil, 0, 0); err == nil {
		t.Fatal("a mutated copy must not admit a new name")
	}
}

func TestNewRefusesAMissingSpec(t *testing.T) {
	t.Parallel()
	if _, err := perturb.New("w-1", 1, nil); !errors.Is(err, perturb.ErrNoSpec) {
		t.Fatalf("err = %v, want ErrNoSpec", err)
	}
}

func TestApplyAdmitsOnlyCatalogNamesWithDeclaredParameters(t *testing.T) {
	t.Parallel()
	layer := shippedLayer(t)
	if _, err := layer.Apply("no_such_perturbation", nil, 0, 0); err == nil || !strings.Contains(err.Error(), "unknown perturbation") {
		t.Fatalf("unknown name: %v", err)
	}
	if _, err := layer.Apply("drop", map[string]any{"rate": 2.0}, 0, 0); err == nil {
		t.Fatal("an out-of-range rate must be rejected")
	}
	if _, err := layer.Apply("drop", map[string]any{"bogus": 1.0}, 0, 0); err == nil || !strings.Contains(err.Error(), "bogus") {
		t.Fatalf("an undeclared parameter must be named: %v", err)
	}
}

func TestProcessAppliesAnActivePerturbationUntilItIsCleared(t *testing.T) {
	t.Parallel()
	layer := shippedLayer(t)
	id, err := layer.Apply("drop", map[string]any{"rate": 1.0}, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	dropped := layer.Process(event(0), model.DefaultStartTimeNS)
	if len(dropped) != 1 || dropped[0].Delivered || dropped[0].Reason != model.DeliveryDroppedByPerturb {
		t.Fatalf("active drop: %+v", dropped)
	}
	if err := layer.Clear(id); err != nil {
		t.Fatal(err)
	}
	passed := layer.Process(event(1), model.DefaultStartTimeNS+1e9)
	if len(passed) != 1 || !passed[0].Delivered || passed[0].Reason != model.DeliveryOK {
		t.Fatalf("after clear: %+v", passed)
	}
	if err := layer.Clear(id); err == nil || !strings.Contains(err.Error(), "unknown perturbation id") {
		t.Fatalf("second clear: %v", err)
	}
}

func TestFlushReleasesRecordsHeldByAFlap(t *testing.T) {
	t.Parallel()
	layer := shippedLayer(t)
	until := model.DefaultStartTimeNS + 30e9
	if _, err := layer.Apply("producer_flap", nil, 0, until); err != nil {
		t.Fatal(err)
	}
	if held := layer.Process(event(0), model.DefaultStartTimeNS); len(held) != 0 {
		t.Fatalf("a flapping producer holds the record, got %+v", held)
	}
	if released := layer.Flush(until + 1); len(released) != 1 || released[0].Event.Seq != 0 {
		t.Fatalf("flush released %+v", released)
	}
}
