package component_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/component"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestEveryShippedExamplePlantComposesAndMatchesTheRepositoryExample(t *testing.T) {
	t.Parallel()
	shipped := map[string]string{
		"belt-conveyor-drives-a": testsupport.BeltConveyorComposition("a"),
		"belt-conveyor-drives-b": testsupport.BeltConveyorComposition("b"),
		"chiller-pilot":          testsupport.PilotComposition(),
		"factory-utilities":      testsupport.FactoryComposition(),
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
		found, err := component.Example(example.ID)
		if err != nil {
			t.Fatal(err)
		}
		document := found.Composition
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
	plant, err := component.NewComposition(base.Composition, component.NewPlant{ID: "lab", Title: "Lab", Empty: true}, "example chiller-pilot")
	if err != nil || plant["id"] != "lab" || plant["equipment"] != nil {
		t.Fatalf("NewComposition = %v, %v", plant, err)
	}
}

func TestTheShippedExamplesCarryTheRepositoryKindFilesTheirPlantsImplement(t *testing.T) {
	t.Parallel()
	belt := []string{"belt-controller.board.json", "belt_conveyor.kind.json", "bulk_feeder.kind.json", "conveyor_drive.kind.json", "drive-controller.board.json", "feeder-controller.board.json"}
	shipped := map[string][]string{"belt-conveyor-drives-a": belt, "belt-conveyor-drives-b": belt, "chiller-pilot": {}, "factory-utilities": {}}
	for id, want := range shipped {
		found, err := component.Example(id)
		if err != nil {
			t.Fatal(err)
		}
		names := []string{}
		for _, file := range found.KindFiles {
			names = append(names, file.Name)
			repository, err := os.ReadFile(filepath.Join(testsupport.BeltConveyorKindsDir(), file.Name))
			if err != nil || !bytes.Equal(repository, file.Document) {
				t.Fatalf("example %s carries a %s that differs from its repository copy (%v)", id, file.Name, err)
			}
		}
		if !slices.Equal(names, want) {
			t.Fatalf("example %s carries %v, want %v", id, names, want)
		}
	}
}

func TestInstallingAnExamplesKindFilesPutsThemInTheKindsFolder(t *testing.T) {
	t.Parallel()
	kinds := filepath.Join(t.TempDir(), "kinds")
	installed, err := component.InstallExampleKinds("belt-conveyor-drives-a", kinds)
	if err != nil || len(installed.Written) != 6 || installed.Example != "belt-conveyor-drives-a" {
		t.Fatalf("InstallExampleKinds = %+v, %v", installed, err)
	}
	if _, err := os.Stat(filepath.Join(kinds, "feeder-controller.board.json")); err != nil {
		t.Fatal(err)
	}
}
