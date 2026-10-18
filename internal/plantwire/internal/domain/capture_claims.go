package domain

import (
	"fmt"
	"maps"
	"slices"
)

var boundKeywords = []string{"minimum", "maximum", "exclusiveMinimum", "exclusiveMaximum", "multipleOf", "minLength", "maxLength",
	"minItems", "maxItems", "minProperties", "maxProperties"}

func claimsOf(entry map[string]any) []string {
	claims := []string{}
	commands, _ := entry["commands"].([]any)
	for i, command := range commands {
		fields, _ := command.(map[string]any)
		base := fmt.Sprintf("/commands/%d", i)
		claims = append(claims, commandClaims(base, fields)...)
		claims = append(claims, boundClaims(base+"/args_schema", fields["args_schema"])...)
	}
	return claims
}

func commandClaims(base string, command map[string]any) []string {
	var claims []string
	for _, member := range []string{"risk_class", "safe_stop"} {
		if _, present := command[member]; present {
			claims = append(claims, base+"/"+member)
		}
	}
	return claims
}

func boundClaims(pointer string, schema any) []string {
	fields, isObject := schema.(map[string]any)
	if !isObject {
		return nil
	}
	var claims []string
	for _, name := range slices.Sorted(maps.Keys(fields)) {
		claims = append(claims, memberClaims(pointer+"/"+pointerToken(name), name, fields[name])...)
	}
	return claims
}

func memberClaims(pointer, name string, value any) []string {
	switch {
	case slices.Contains(boundKeywords, name):
		return []string{pointer}
	case name == "enum" || name == "const":
		return nil
	case name == "properties":
		return propertyClaims(pointer, value)
	}
	return nestedClaims(pointer, value)
}

func nestedClaims(pointer string, value any) []string {
	switch nested := value.(type) {
	case map[string]any:
		return boundClaims(pointer, nested)
	case []any:
		return listClaims(pointer, nested)
	}
	return nil
}

func listClaims(pointer string, items []any) []string {
	var claims []string
	for i, item := range items {
		claims = append(claims, boundClaims(fmt.Sprintf("%s/%d", pointer, i), item)...)
	}
	return claims
}

func propertyClaims(pointer string, value any) []string {
	properties, _ := value.(map[string]any)
	var claims []string
	for _, name := range slices.Sorted(maps.Keys(properties)) {
		claims = append(claims, boundClaims(pointer+"/"+pointerToken(name), properties[name])...)
	}
	return claims
}
