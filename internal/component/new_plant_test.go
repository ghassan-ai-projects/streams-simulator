package component_test

import (
	"encoding/json"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/component"
)

func TestAnEmptyPlantFromEveryShippedExampleComposesOnceEquipmentIsAdded(t *testing.T) {
	t.Parallel()
	examples, err := component.Examples()
	if err != nil || len(examples) == 0 {
		t.Fatalf("Examples = %v, %v", examples, err)
	}
	library, err := component.NewLibrary(nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, example := range examples {
		t.Run(example.ID, func(t *testing.T) {
			t.Parallel()
			base, err := component.Example(example.ID)
			if err != nil {
				t.Fatal(err)
			}
			plant, err := component.NewComposition(base, component.NewPlant{ID: "fresh-plant", Title: "Fresh plant", Empty: true}, example.ID)
			if err != nil {
				t.Fatal(err)
			}
			areas, _ := plant["areas"].([]any)
			if len(areas) != 1 {
				t.Fatalf("an empty plant from %s has areas %v, want one", example.ID, plant["areas"])
			}
			area, _ := areas[0].(map[string]any)
			if _, nested := area["parent"]; nested || plant["site"] == nil {
				t.Fatalf("an empty plant from %s has area %v and site %v", example.ID, area, plant["site"])
			}
			plant["equipment"] = []any{map[string]any{"use": "streamsim/pump@1", "names": []any{"P-1"}, "area": area["name"]}}
			raw, err := json.Marshal(plant)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := component.Compose(library, raw, "fresh-plant.composition.json"); err != nil {
				t.Fatalf("an empty plant from %s does not compose: %v", example.ID, err)
			}
		})
	}
}
