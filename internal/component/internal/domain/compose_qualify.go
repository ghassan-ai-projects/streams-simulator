package domain

import (
	"maps"
	"slices"
)

var annotationKeywords = []string{"description", "title", "$comment", "examples"}

func qualifyFaults(kind map[string]any, name string) {
	for _, fault := range members(kind, "faults") {
		fault["id"] = name + "__" + stringOf(fault["id"])
		qualifyDetector(mapOf(mapOf(fault["observability"])["detector"]), name)
		qualifyMember(mapOf(fault["observation"]), name, "channel")
		if effector, named := fault["expected_effector"].(string); named {
			fault["expected_effector"] = name + "__" + effector
		}
	}
}

func qualifyDetector(detector map[string]any, name string) {
	qualifyMember(detector, name, detectorChannels...)
	for _, member := range detectorChannelLists {
		if channels, listed := detector[member].([]any); listed {
			qualifyEach(channels, name)
		}
	}
}

func qualifyEach(channels []any, name string) {
	for i, channel := range channels {
		channels[i] = name + "." + stringOf(channel)
	}
}

func qualifyMember(owner map[string]any, name string, members ...string) {
	for _, member := range members {
		if channel, named := owner[member].(string); named {
			owner[member] = name + "." + channel
		}
	}
}

func qualifyEffectors(kind map[string]any, name string) {
	for _, effector := range members(kind, "effectors") {
		effector["name"] = name + "__" + stringOf(effector["name"])
		channels, _ := effector["confirmation_channels"].([]any)
		qualifyEach(channels, name)
	}
}

func stripDescriptions(kind map[string]any) {
	for _, section := range []string{"state", "dynamics", "flows", "faults", "effectors", "boundaries"} {
		for _, item := range members(kind, section) {
			delete(item, "description")
		}
	}
	for _, effector := range members(kind, "effectors") {
		stripAnnotations(effector["args_schema"])
	}
}

func stripAnnotations(value any) {
	switch typed := value.(type) {
	case map[string]any:
		for _, key := range slices.Sorted(maps.Keys(typed)) {
			stripAnnotation(typed, key)
		}
	case []any:
		for _, item := range typed {
			stripAnnotations(item)
		}
	}
}

func stripAnnotation(object map[string]any, key string) {
	if slices.Contains(annotationKeywords, key) {
		delete(object, key)
		return
	}
	stripAnnotations(object[key])
}

func mapOf(value any) map[string]any {
	out, _ := value.(map[string]any)
	return out
}

func stringOf(value any) string {
	out, _ := value.(string)
	return out
}
