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

type Examples struct {
	summaries []ExampleSummary
	documents map[string]json.RawMessage
}

func LoadExamples(files []File) (Examples, error) {
	examples := Examples{documents: map[string]json.RawMessage{}}
	for _, file := range files {
		summary, err := exampleSummary(file)
		if err != nil {
			return Examples{}, err
		}
		examples.summaries = append(examples.summaries, summary)
		examples.documents[summary.ID] = json.RawMessage(file.Raw)
	}
	slices.SortFunc(examples.summaries, func(a, b ExampleSummary) int { return strings.Compare(a.Title, b.Title) })
	return examples, nil
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

func (e Examples) Find(id string) (json.RawMessage, error) {
	document, ok := e.documents[id]
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrUnknownExample, id)
	}
	return document, nil
}
