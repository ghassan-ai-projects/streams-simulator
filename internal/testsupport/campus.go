package testsupport

import (
	"encoding/json"
	"fmt"
	"strings"
)

type campus struct {
	halls  int
	shared map[string]bool
}

func Campus(composition []byte, halls int, shared ...string) ([]byte, error) {
	var doc map[string]any
	if err := json.Unmarshal(composition, &doc); err != nil {
		return nil, fmt.Errorf("testsupport: campus: %w", err)
	}
	newCampus(halls, shared).replicate(doc)
	raw, err := json.Marshal(doc)
	if err != nil {
		return nil, fmt.Errorf("testsupport: campus: %w", err)
	}
	return raw, nil
}

func newCampus(halls int, shared []string) campus {
	c := campus{halls: halls, shared: map[string]bool{}}
	for _, name := range shared {
		c.shared[name] = true
	}
	return c
}

func (c campus) replicate(doc map[string]any) {
	doc["id"], doc["title"] = fmt.Sprintf("%v-campus-%d", doc["id"], c.halls), fmt.Sprintf("%v campus of %d halls", doc["title"], c.halls)
	c.replicateEquipment(doc["equipment"].([]any))
	doc["connections"] = c.replicateConnections(doc["connections"].([]any))
}

func (c campus) hallName(hall int, name string) string {
	if c.shared[name] {
		return name
	}
	return fmt.Sprintf("h%02d/%s", hall, name)
}

func (c campus) replicateEquipment(equipment []any) {
	for _, raw := range equipment {
		entry := raw.(map[string]any)
		var names []any
		for hall := 1; hall <= c.halls; hall++ {
			names = append(names, c.hallNames(entry["names"].([]any), hall)...)
		}
		entry["names"] = names
	}
}

func (c campus) hallNames(names []any, hall int) []any {
	var out []any
	for _, name := range names {
		if renamed := c.hallName(hall, name.(string)); hall == 1 || renamed != name {
			out = append(out, renamed)
		}
	}
	return out
}

func (c campus) replicateConnections(connections []any) []any {
	var out []any
	for _, raw := range connections {
		connection := raw.(map[string]any)
		for hall := 1; hall <= c.halls; hall++ {
			from, to := c.endpoint(hall, connection["from"].(string)), c.endpoint(hall, connection["to"].(string))
			if hall == 1 || from != connection["from"] || to != connection["to"] {
				out = append(out, map[string]any{"from": from, "to": to})
			}
		}
	}
	return out
}

func (c campus) endpoint(hall int, text string) string {
	cut := strings.LastIndex(text, ".")
	return c.hallName(hall, text[:cut]) + text[cut:]
}
