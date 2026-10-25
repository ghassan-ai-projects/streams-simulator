package run_test

import (
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/perturb"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

func TestTheRunDescribesTheDeliveryPerturbationsItWillApply(t *testing.T) {
	t.Parallel()
	kinds := run.Perturbations()
	names := perturb.Names()
	if len(kinds) != len(names) {
		t.Fatalf("%d kinds described, want the %d the delivery layer applies", len(kinds), len(names))
	}
	for i, kind := range kinds {
		if kind.Kind != names[i] || kind.Category != run.CategoryDelivery {
			t.Fatalf("kind %d = %+v, want %s as a delivery perturbation", i, kind, names[i])
		}
	}
}
