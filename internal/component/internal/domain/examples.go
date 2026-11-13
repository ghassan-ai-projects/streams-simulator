package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
)

var ErrUnknownExample = errors.New("component: no such example plant")

type ExampleSummary struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
}

type ExamplePlant struct {
	Composition json.RawMessage
	KindFiles   []KindFile
}

type Examples struct {
	summaries []ExampleSummary
	documents map[string]json.RawMessage
	kindFiles exampleKindFiles
}

func LoadExamples(compositions, kindFiles []File) (Examples, error) {
	pool, err := loadExampleKindFiles(kindFiles)
	if err != nil {
		return Examples{}, err
	}
	examples := Examples{documents: map[string]json.RawMessage{}, kindFiles: pool}
	for _, file := range compositions {
		if err := examples.add(file); err != nil {
			return Examples{}, err
		}
	}
	slices.SortFunc(examples.summaries, func(a, b ExampleSummary) int { return strings.Compare(a.Title, b.Title) })
	return examples, nil
}

func (e *Examples) add(file File) error {
	summary, err := exampleSummary(file)
	if err != nil {
		return err
	}
	e.summaries = append(e.summaries, summary)
	e.documents[summary.ID] = json.RawMessage(file.Raw)
	return nil
}

func exampleSummary(file File) (ExampleSummary, error) {
	var summary ExampleSummary
	if err := json.Unmarshal(file.Raw, &summary); err != nil {
		return ExampleSummary{}, fmt.Errorf("component: example %s: %w", file.Path, err)
	}
	if summary.ID == "" {
		return ExampleSummary{}, fmt.Errorf("component: example %s has no id", file.Path)
	}
	return summary, nil
}

func (e Examples) List() []ExampleSummary {
	return slices.Clone(e.summaries)
}

func (e Examples) Find(id string) (ExamplePlant, error) {
	document, ok := e.documents[id]
	if !ok {
		return ExamplePlant{}, fmt.Errorf("%w: %q", ErrUnknownExample, id)
	}
	kindFiles, err := e.kindFiles.of(id, document)
	if err != nil {
		return ExamplePlant{}, err
	}
	return ExamplePlant{Composition: document, KindFiles: kindFiles}, nil
}
