package domain

import (
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonschema"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/schemas"
)

type NewPlant struct {
	ID    string
	Title string
	Empty bool
}

var leftOutOfAnEmptyPlant = []string{"description", "equipment", "connections"}

func NewComposition(base []byte, plant NewPlant, source string) (map[string]any, error) {
	contract, err := compileContract(model.ContractComposition, schemas.Composition())
	if err != nil {
		return nil, err
	}
	var doc map[string]any
	if err := contract.decode(base, source, &doc); err != nil {
		return nil, err
	}
	doc["id"], doc["title"] = plant.ID, plant.Title
	if plant.Empty {
		emptyPlant(doc)
	}
	return doc, contract.checkNew(doc, source, plant.Empty)
}

func emptyPlant(doc map[string]any) {
	for _, key := range leftOutOfAnEmptyPlant {
		delete(doc, key)
	}
	doc["profiles"] = emptyProfiles(doc["profiles"])
}

func nominalProfile() map[string]any {
	return map[string]any{"name": "nominal", "duration_s": float64(86400)}
}

func emptyProfiles(raw any) []any {
	listed, _ := raw.([]any)
	kept := []any{}
	for _, entry := range listed {
		if profile, ok := entry.(map[string]any); ok && profile["not_applicable"] == nil {
			delete(profile, "fault_weights")
			kept = append(kept, profile)
		}
	}
	if len(kept) == 0 {
		kept = append(kept, nominalProfile())
	}
	return kept
}

func (c contractSchema) checkNew(doc map[string]any, source string, empty bool) error {
	var kept []jsonschema.Error
	for _, issue := range c.schema.Validate(doc) {
		if !empty || !strings.HasPrefix(issue.Path, "/equipment") {
			kept = append(kept, issue)
		}
	}
	if len(kept) > 0 {
		return violation(c.contract, source, model.SchemaIssues(kept))
	}
	return nil
}
