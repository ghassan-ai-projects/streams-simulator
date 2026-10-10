package perturb_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/perturb"
)

func shippedLayer(t *testing.T) *perturb.Layer {
	t.Helper()
	spec, err := domain.Load(filepath.Join("..", "..", "domains", "rotating-machinery.domain.json"))
	if err != nil {
		t.Fatal(err)
	}
	return perturb.New("w-1", 7, spec)
}

func event(seq int64) model.SimEvent {
	at := model.FormatTime(model.DefaultStartTimeNS + seq*1e9)
	return model.SimEvent{Seq: seq, WorldID: "w-1", EntityType: "pump", EntityID: "site-a/pump-1",
		Channel: "motor_current", EventTime: at, ObservedTime: at, Value: 30.0, Unit: "A"}
}

func TestCatalogListsEveryNamedPerturbationOnce(t *testing.T) {
	t.Parallel()
	seen := map[string]bool{}
	for _, name := range perturb.Names {
		if seen[name] {
			t.Fatalf("%s listed twice", name)
		}
		seen[name] = true
	}
	for _, name := range []string{perturb.Drop, perturb.DuplicateBurst, perturb.Reorder, perturb.InjectionProbe} {
		if !seen[name] {
			t.Fatalf("catalog lacks %s", name)
		}
	}
	if len(perturb.Names) != 19 {
		t.Fatalf("catalog has %d perturbations, want 19", len(perturb.Names))
	}
}

func TestApplyAdmitsOnlyCatalogNamesWithDeclaredParameters(t *testing.T) {
	t.Parallel()
	layer := shippedLayer(t)
	if _, err := layer.Apply("no_such_perturbation", nil, 0, 0); err == nil || !strings.Contains(err.Error(), "unknown perturbation") {
		t.Fatalf("unknown name: %v", err)
	}
	if _, err := layer.Apply(perturb.Drop, map[string]any{"rate": 2.0}, 0, 0); err == nil {
		t.Fatal("an out-of-range rate must be rejected")
	}
	if _, err := layer.Apply(perturb.Drop, map[string]any{"bogus": 1.0}, 0, 0); err == nil || !strings.Contains(err.Error(), "bogus") {
		t.Fatalf("an undeclared parameter must be named: %v", err)
	}
}

func TestProcessAppliesAnActivePerturbationUntilItIsCleared(t *testing.T) {
	t.Parallel()
	layer := shippedLayer(t)
	id, err := layer.Apply(perturb.Drop, map[string]any{"rate": 1.0}, 0, 0)
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
	if _, err := layer.Apply(perturb.ProducerFlap, nil, 0, until); err != nil {
		t.Fatal(err)
	}
	if held := layer.Process(event(0), model.DefaultStartTimeNS); len(held) != 0 {
		t.Fatalf("a flapping producer holds the record, got %+v", held)
	}
	if released := layer.Flush(until + 1); len(released) != 1 || released[0].Event.Seq != 0 {
		t.Fatalf("flush released %+v", released)
	}
}
