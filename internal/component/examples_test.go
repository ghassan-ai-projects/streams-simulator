package component_test

import (
	"bytes"
	"errors"
	"os"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/component"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestEveryShippedExamplePlantComposesAndMatchesTheRepositoryExample(t *testing.T) {
	t.Parallel()
	shipped := map[string]string{
		"chiller-pilot":     testsupport.PilotComposition(),
		"factory-utilities": testsupport.FactoryComposition(),
	}
	examples, err := component.Examples()
	if err != nil || len(examples) != len(shipped) {
		t.Fatalf("Examples = %v, %v", examples, err)
	}
	library, err := component.NewLibrary(nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, example := range examples {
		document, err := component.Example(example.ID)
		if err != nil {
			t.Fatal(err)
		}
		repository, err := os.ReadFile(shipped[example.ID])
		if err != nil || !bytes.Equal(repository, document) {
			t.Fatalf("example %s differs from its repository copy (%v)", example.ID, err)
		}
		if _, err := component.Compose(library, document, example.ID); err != nil {
			t.Fatalf("example %s does not compose: %v", example.ID, err)
		}
	}
	if _, err := component.Example("nope"); !errors.Is(err, component.ErrUnknownExample) {
		t.Fatalf("Example(nope) error = %v", err)
	}
}

func TestANewPlantStartsFromAShippedExample(t *testing.T) {
	t.Parallel()
	base, err := component.Example("chiller-pilot")
	if err != nil {
		t.Fatal(err)
	}
	plant, err := component.NewComposition(base, component.NewPlant{ID: "lab", Title: "Lab", Empty: true}, "example chiller-pilot")
	if err != nil || plant["id"] != "lab" || plant["equipment"] != nil {
		t.Fatalf("NewComposition = %v, %v", plant, err)
	}
}
