package files

import (
	"embed"
	"encoding/json"

	"github.com/ghassan-ai-projects/streams-simulator/internal/component/internal/domain"
)

//go:embed examples/*.composition.json
var builtInExamples embed.FS

var builtInExampleTree = embeddedTree{files: builtInExamples, root: "."}

func ListExamples() ([]domain.ExampleSummary, error) {
	examples, err := loadExamples(builtInExampleTree)
	return examples.List(), err
}

func FindExample(id string) (json.RawMessage, error) {
	return findExample(builtInExampleTree, id)
}

func findExample(source tree, id string) (json.RawMessage, error) {
	examples, err := loadExamples(source)
	if err != nil {
		return nil, err
	}
	return examples.Find(id)
}

func loadExamples(source tree) (domain.Examples, error) {
	files, err := readGroup(source, ".", "examples", ".composition.json")
	if err != nil {
		return domain.Examples{}, err
	}
	return domain.LoadExamples(files)
}
