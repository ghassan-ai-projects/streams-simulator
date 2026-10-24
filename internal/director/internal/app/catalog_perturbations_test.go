package app

import (
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

func TestThePerturbationCatalogListsEveryDeliveryKindWithItsParameters(t *testing.T) {
	t.Parallel()
	d := shippedDirector(t, t.Context())
	listed, ok := d.PerturbationCatalog()["perturbations"].([]run.PerturbationKind)
	if !ok || len(listed) != len(run.Perturbations()) {
		t.Fatalf("catalog = %v, want the kinds the run applies", d.PerturbationCatalog())
	}
	for _, kind := range listed {
		if kind.Category != run.CategoryDelivery || kind.Description == "" {
			t.Fatalf("kind = %+v, want a described delivery perturbation", kind)
		}
	}
}

func TestDescribingADomainSaysWhichLayerEachFaultActsOn(t *testing.T) {
	t.Parallel()
	d := shippedDirector(t, t.Context())
	described, err := d.CatalogDescribe("aquaculture-pond")
	if err != nil {
		t.Fatal(err)
	}
	categories, _ := described["fault_categories"].(map[string]map[string]string)
	if got := categories["pond"]["aerator_failure"]; got != "world" {
		t.Fatalf("fault_categories = %v, want aerator_failure to be a world fault of the pond kind", categories)
	}
}
