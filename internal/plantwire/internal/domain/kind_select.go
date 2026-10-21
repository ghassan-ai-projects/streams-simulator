package domain

import (
	"encoding/json"
	"path"
	"slices"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func ImplementedKinds(spec *model.DomainSpec) map[string]bool {
	wanted := map[string]bool{}
	for i := range spec.Kinds {
		if spec.Kinds[i].Implements != "" {
			wanted[spec.Kinds[i].Implements] = true
		}
	}
	return wanted
}

func SelectKindSources(sources []KindSource, wanted map[string]bool) []KindSource {
	return slices.DeleteFunc(slices.Clone(sources), func(source KindSource) bool { return !wanted[kindNameOfSource(source)] })
}

func kindNameOfSource(source KindSource) string {
	var header struct {
		Kind string `json:"kind"`
	}
	if source.Err == nil && json.Unmarshal(source.Raw, &header) == nil && header.Kind != "" {
		return header.Kind
	}
	return strings.TrimSuffix(path.Base(source.Path), KindFileSuffix)
}

func BoardTypes(spec *model.DomainSpec) map[string]bool {
	wanted := map[string]bool{}
	for i := range spec.Kinds {
		if spec.Kinds[i].BoardType != "" {
			wanted[spec.Kinds[i].BoardType] = true
		}
	}
	return wanted
}

func SelectBoardSources(sources []KindSource, wanted map[string]bool) []KindSource {
	return slices.DeleteFunc(slices.Clone(sources), func(source KindSource) bool { return !wanted[boardTypeOfSource(source)] })
}

func boardTypeOfSource(source KindSource) string {
	var header struct {
		BoardType string `json:"board_type"`
	}
	if source.Err == nil && json.Unmarshal(source.Raw, &header) == nil && header.BoardType != "" {
		return header.BoardType
	}
	return strings.TrimSuffix(path.Base(source.Path), BoardFileSuffix)
}
