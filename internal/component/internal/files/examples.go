package files

import (
	"embed"

	"github.com/ghassan-ai-projects/streams-simulator/internal/component/internal/domain"
)

//go:embed examples/*.composition.json examples/kinds/*.kind.json examples/kinds/*.board.json
var builtInExamples embed.FS

var builtInExampleTree = embeddedTree{files: builtInExamples, root: "."}

func ListExamples() ([]domain.ExampleSummary, error) {
	examples, err := loadExamples(builtInExampleTree)
	return examples.List(), err
}

func FindExample(id string) (domain.ExamplePlant, error) {
	return findExample(builtInExampleTree, id)
}

func findExample(source tree, id string) (domain.ExamplePlant, error) {
	examples, err := loadExamples(source)
	if err != nil {
		return domain.ExamplePlant{}, err
	}
	return examples.Find(id)
}

func loadExamples(source tree) (domain.Examples, error) {
	compositions, err := readGroup(source, ".", "examples", ".composition.json")
	if err != nil {
		return domain.Examples{}, err
	}
	kindFiles, err := readGroup(source, "examples", "kinds", ".json")
	if err != nil {
		return domain.Examples{}, err
	}
	return domain.LoadExamples(compositions, kindFiles)
}
