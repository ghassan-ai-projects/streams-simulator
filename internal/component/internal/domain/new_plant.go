package domain

import (
	"encoding/json"
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

func emptyProfiles(raw any) []any {
	listed, _ := raw.([]any)
	kept, named := []any{}, map[string]bool{}
	for _, entry := range listed {
		if profile, ok := entry.(map[string]any); ok && profile["not_applicable"] == nil {
			delete(profile, "fault_weights")
			kept = append(kept, profile)
			named[profile["name"].(string)] = true
		}
	}
	return append(kept, missingRequiredProfiles(named)...)
}

func missingRequiredProfiles(named map[string]bool) []any {
	var missing []any
	for _, name := range requiredProfileNames() {
		if !named[name] {
			missing = append(missing, map[string]any{"name": name, "duration_s": float64(86400)})
		}
	}
	return missing
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

type profileClause struct {
	Contains struct {
		Properties struct {
			Name struct {
				Const string `json:"const"`
			} `json:"name"`
		} `json:"properties"`
	} `json:"contains"`
}

type profileRules struct {
	Properties struct {
		Profiles struct {
			AllOf []profileClause `json:"allOf"`
		} `json:"profiles"`
	} `json:"properties"`
}

func requiredProfileNames() []string {
	var spec profileRules
	if json.Unmarshal(schemas.DomainSpec(), &spec) != nil {
		return nil
	}
	var names []string
	for _, clause := range spec.Properties.Profiles.AllOf {
		names = append(names, clause.Contains.Properties.Name.Const)
	}
	return names
}
