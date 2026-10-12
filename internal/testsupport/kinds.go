package testsupport

import (
	"bytes"
	"encoding/json"
	"strings"
)

var kindSections = []string{"state", "dynamics", "channels", "faults", "effectors"}

func AsKinds(flat []byte) []byte {
	doc := decodeFlat(flat)
	entities, _ := doc["entities"].(map[string]any)
	name := kindName(doc, entities)
	doc["kinds"] = []any{kindOf(doc, name)}
	doc["instances"] = []any{instancesOf(entities, name)}
	delete(doc, "entities")
	out, err := json.Marshal(doc)
	if err != nil {
		panic("testsupport: kinded domain: " + err.Error())
	}
	return out
}

func decodeFlat(flat []byte) map[string]any {
	var doc map[string]any
	decoder := json.NewDecoder(bytes.NewReader(flat))
	decoder.UseNumber()
	if err := decoder.Decode(&doc); err != nil {
		panic("testsupport: flat domain: " + err.Error())
	}
	return doc
}

func kindName(doc, entities map[string]any) string {
	if name, ok := entities["entity_type"].(string); ok {
		return name
	}
	id, _ := doc["id"].(string)
	return strings.Split(id, "-")[0]
}

func kindOf(doc map[string]any, name string) map[string]any {
	kind := map[string]any{"name": name}
	for _, section := range kindSections {
		if value, ok := doc[section]; ok {
			kind[section] = value
			delete(doc, section)
		}
	}
	return kind
}

func instancesOf(entities map[string]any, name string) map[string]any {
	instances := map[string]any{"kind": name}
	for _, member := range []string{"id_template", "count", "churn"} {
		if value, ok := entities[member]; ok {
			instances[member] = value
		}
	}
	return instances
}
