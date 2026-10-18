package domain

import (
	"slices"
	"strings"
)

// maxStemBytes keeps "<stem>.kind.json" within the 128 bytes of a workspace path segment.
const maxStemBytes = 128 - len(KindFileSuffix)

func foldKind(name string) string {
	return strings.ToLower(name)
}

func DuplicateKinds(kinds []any) []int {
	var repeated []int
	seen := map[string]bool{}
	for i, kind := range kinds {
		entry, _ := kind.(map[string]any)
		name := foldKind(kindNameOf(entry))
		if seen[name] {
			repeated = append(repeated, i)
		}
		seen[name] = true
	}
	return repeated
}

func KindsDuplicated(kinds []any) bool {
	return len(DuplicateKinds(kinds)) > 0
}

func KindFileStem(kind string) string {
	return strings.ReplaceAll(kind[:min(len(kind), maxStemBytes)], ":", "_")
}

func SelectFolded(sources []KindSource, names []string) []KindSource {
	wanted := map[string]bool{}
	for _, name := range names {
		wanted[foldKind(name)] = true
	}
	return slices.DeleteFunc(slices.Clone(sources), func(source KindSource) bool {
		return !wanted[foldKind(kindNameOfSource(source))]
	})
}
