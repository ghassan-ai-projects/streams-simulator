package domain

import (
	"maps"
	"slices"
	"strconv"
)

func firstDifference(want, got any, path string) (string, bool) {
	switch expected := want.(type) {
	case map[string]any:
		return differingMember(expected, got, path)
	case []any:
		return differingItem(expected, got, path)
	}
	return path, want != got
}

func differingMember(want map[string]any, got any, path string) (string, bool) {
	actual, isObject := got.(map[string]any)
	if !isObject {
		return path, true
	}
	for _, name := range memberNames(want, actual) {
		if at, differs := memberDifference(want, actual, name, path+"/"+name); differs {
			return at, true
		}
	}
	return path, false
}

func memberDifference(want, got map[string]any, name, path string) (string, bool) {
	expected, inKindFile := want[name]
	found, inKind := got[name]
	if inKindFile != inKind {
		return path, true
	}
	return firstDifference(expected, found, path)
}

func memberNames(a, b map[string]any) []string {
	names := slices.Sorted(maps.Keys(a))
	for _, name := range slices.Sorted(maps.Keys(b)) {
		if _, shared := a[name]; !shared {
			names = append(names, name)
		}
	}
	slices.Sort(names)
	return names
}

func differingItem(want []any, got any, path string) (string, bool) {
	actual, isList := got.([]any)
	if !isList {
		return path, true
	}
	for i := range max(len(want), len(actual)) {
		if at, differs := itemDifference(want, actual, i, path+"/"+strconv.Itoa(i)); differs {
			return at, true
		}
	}
	return path, false
}

func itemDifference(want, got []any, index int, path string) (string, bool) {
	if index >= len(want) || index >= len(got) {
		return path, true
	}
	return firstDifference(want[index], got[index], path)
}
