package files

import (
	"embed"
	"encoding/json"

	"github.com/ghassan-ai-projects/streams-simulator/internal/component/internal/domain"
)

//go:embed examples/*.composition.json
var builtInExamples embed.FS

func ListExamples() ([]domain.ExampleSummary, error) {
	examples, err := loadExamples()
	return examples.List(), err
}

func FindExample(id string) (json.RawMessage, error) {
	examples, err := loadExamples()
	if err != nil {
		return nil, err
	}
	return examples.Find(id)
}

func loadExamples() (domain.Examples, error) {
	files, err := readGroup(embeddedTree{files: builtInExamples, root: "."}, ".", "examples", ".composition.json")
	if err != nil {
		return domain.Examples{}, err
	}
	return domain.LoadExamples(files)
}
