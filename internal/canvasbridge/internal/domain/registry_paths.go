package domain

import (
	"fmt"
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
