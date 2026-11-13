package model

import (
	"slices"
	"strings"
)

func (e *Effector) EntityArguments() []string {
	required, _ := e.ArgsSchema["required"].([]any)
	var names []string
	for _, item := range required {
		if name, _ := item.(string); e.isEntityArgument(name) {
			names = append(names, name)
		}
	}
	slices.Sort(names)
	return names
}

func (e *Effector) isEntityArgument(name string) bool {
	return strings.HasSuffix(name, "_id") && e.argumentType(name) == "string"
}

func (e *Effector) argumentType(name string) string {
	properties, _ := e.ArgsSchema["properties"].(map[string]any)
	property, _ := properties[name].(map[string]any)
	kind, _ := property["type"].(string)
	return kind
}
