package domain

import (
	"fmt"
	"slices"
	"strings"
)

func (spec CLIArg) renderText(value any) (ArgvPart, error) {
	text, ok := value.(string)
	if !ok || text == "" || strings.HasPrefix(text, "-") {
		return ArgvPart{}, fmt.Errorf("want a non-empty value that is not a flag, got %v", value)
	}
	return ArgvPart{Literal: text}, nil
}

func (spec CLIArg) renderPath(value any) (ArgvPart, error) {
	text, ok := value.(string)
	if !ok {
		return ArgvPart{}, fmt.Errorf("want a workspace path, got %v", value)
	}
	if spec.Shape == ShapeRun {
		return runPart(text, spec.RunFile)
	}
	path, err := ParsePath(text)
	if err != nil || path.Class != spec.Class {
		return ArgvPart{}, fmt.Errorf("want a %s file of the workspace, got %q", spec.Class, text)
	}
	return ArgvPart{Path: path.Rel}, nil
}

func (spec CLIArg) renderSwitch(value any) ([]ArgvPart, error) {
	on, ok := value.(bool)
	if !ok {
		return nil, fmt.Errorf("want true or false, got %v", value)
	}
	if !on {
		return nil, nil
	}
	return literals(flagOf(spec)), nil
}

func runPart(runID, file string) (ArgvPart, error) {
	if badSegment(runID) {
		return ArgvPart{}, fmt.Errorf("want a run id, got %q", runID)
	}
	if file == "" {
		return ArgvPart{Path: "runs/" + runID}, nil
	}
	return ArgvPart{Path: "runs/" + runID + "/" + file}, nil
}

func literals(words []string) []ArgvPart {
	parts := make([]ArgvPart, 0, len(words))
	for _, word := range words {
		parts = append(parts, ArgvPart{Literal: word})
	}
	return parts
}

func flagOf(spec CLIArg) []string {
	if spec.Flag == "" {
		return nil
	}
	return []string{spec.Flag}
}

const maxStems = 64

func (spec CLIArg) renderStems(value any) ([]ArgvPart, error) {
	stems, err := stemsOf(value)
	if err != nil {
		return nil, err
	}
	parts := make([]ArgvPart, len(stems))
	for i, stem := range stems {
		path, err := StemPath(spec.Class, stem)
		if err != nil {
			return nil, err
		}
		parts[i] = ArgvPart{Path: path.Rel, Existing: true}
	}
	return parts, nil
}

func stemsOf(value any) ([]string, error) {
	switch list := value.(type) {
	case []string:
		return distinctStems(slices.Clone(list))
	case []any:
		stems, err := anyStems(list)
		if err != nil {
			return nil, err
		}
		return distinctStems(stems)
	default:
		return nil, fmt.Errorf("want a list of file names without folder or suffix, got %v", value)
	}
}

func anyStems(list []any) ([]string, error) {
	var stems []string
	for _, item := range list {
		text, ok := item.(string)
		if !ok {
			return nil, fmt.Errorf("want file names without folder or suffix, got %v", item)
		}
		stems = append(stems, text)
	}
	return stems, nil
}

func distinctStems(stems []string) ([]string, error) {
	var distinct []string
	for _, stem := range stems {
		if !slices.Contains(distinct, stem) {
			distinct = append(distinct, stem)
		}
	}
	if len(distinct) == 0 || len(distinct) > maxStems {
		return nil, fmt.Errorf("want between 1 and %d different file names, got %d", maxStems, len(distinct))
	}
	return distinct, nil
}
