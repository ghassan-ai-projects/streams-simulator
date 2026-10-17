package domain

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strconv"
)

type ArgShape string

const (
	ShapeText ArgShape = "text"
	ShapePath ArgShape = "path"
	ShapeRun  ArgShape = "run"
	ShapeNum  ArgShape = "number"
	ShapeFlag ArgShape = "switch"
)

type CLIArg struct {
	Key      string
	Flag     string
	Shape    ArgShape
	Class    FileClass
	RunFile  string
	Required bool
}

type ArgvPart struct {
	Literal string
	Path    string
}

var (
	cliComponentList = Action{Name: "component.list", Kind: KindCLI, Command: []string{"component", "list"}, Library: true}
	cliComponentShow = Action{Name: "component.show", Kind: KindCLI, Command: []string{"component", "show"}, Library: true,
		CLIArgs: []CLIArg{{Key: "ref", Shape: ShapeText, Required: true}}}
	cliExamples = Action{Name: "component.examples", Kind: KindCLI, Command: []string{"component", "examples"}}
	cliExample  = Action{Name: "component.example", Kind: KindCLI, Command: []string{"component", "example"},
		CLIArgs: []CLIArg{{Key: "id", Shape: ShapeText, Required: true}}}
	cliNewPlant = Action{Name: "component.new", Kind: KindCLI, Command: []string{"component", "new"},
		CLIArgs: []CLIArg{{Key: "id", Shape: ShapeText, Required: true}, {Key: "title", Flag: "--title", Shape: ShapeText, Required: true},
			{Key: "from", Flag: "--from", Shape: ShapePath, Class: ClassComposition}, {Key: "example", Flag: "--example", Shape: ShapeText},
			{Key: "empty", Flag: "--empty", Shape: ShapeFlag}}}
	cliComponentValidate = Action{Name: "component.validate", Kind: KindCLI, Command: []string{"component", "validate"}, Library: true}
	cliKindValidate      = Action{Name: "kind.validate", Kind: KindCLI, Command: []string{"kind", "validate"},
		CLIArgs: []CLIArg{{Key: "path", Shape: ShapePath, Class: ClassKind, Required: true}}}
	cliKindTwin = Action{Name: "kind.twin", Kind: KindCLI, Command: []string{"kind", "twin"}, Library: true, Kinds: true, Mutating: true, Etag: true,
		CLIArgs: []CLIArg{{Key: "choices", Flag: "--choices", Shape: ShapePath, Class: ClassTwin, Required: true},
			{Key: "out", Flag: "--out", Shape: ShapePath, Class: ClassDomain, Required: true}}}
	cliCompose = Action{Name: "domain.compose", Kind: KindCLI, Command: []string{"domain", "compose"}, Library: true, Kinds: true, Mutating: true, Etag: true,
		CLIArgs: []CLIArg{{Key: "composition", Shape: ShapePath, Class: ClassComposition, Required: true},
			{Key: "out", Flag: "--out", Shape: ShapePath, Class: ClassDomain, Required: true}}}
	cliValidate = Action{Name: "domain.validate", Kind: KindCLI, Command: []string{"domain", "validate"}, Kinds: true,
		CLIArgs: []CLIArg{{Key: "path", Shape: ShapePath, Class: ClassDomain, Required: true}}}
	cliLint = Action{Name: "domain.lint", Kind: KindCLI, Command: []string{"domain", "lint"},
		CLIArgs: []CLIArg{{Key: "path", Shape: ShapePath, Class: ClassDomain, Required: true}, {Key: "hours", Flag: "--hours", Shape: ShapeNum}}}
	cliVerify = Action{Name: "run.verify", Kind: KindCLI, Command: []string{"verify"}, Catalog: true,
		CLIArgs: []CLIArg{{Key: "run", Shape: ShapeRun, RunFile: "run.json", Required: true}}}
	cliReport = Action{Name: "run.report", Kind: KindCLI, Command: []string{"report"},
		CLIArgs: []CLIArg{{Key: "run", Shape: ShapeRun, Required: true}}}
	cliScenario = Action{Name: "scenario.export", Kind: KindCLI, Command: []string{"scenario", "from-run"}, Catalog: true, Mutating: true, Etag: true,
		CLIArgs: []CLIArg{{Key: "run", Shape: ShapeRun, RunFile: "run.json", Required: true},
			{Key: "out", Flag: "--out", Shape: ShapePath, Class: ClassScenario, Required: true}}}
)

func (a Action) Argv(args map[string]any) ([]ArgvPart, error) {
	if err := a.closedArguments(args); err != nil {
		return nil, err
	}
	parts := literals(a.Command)
	for _, spec := range a.CLIArgs {
		rendered, err := a.argument(spec, args)
		if err != nil {
			return nil, err
		}
		parts = append(parts, rendered...)
	}
	return parts, nil
}

func (a Action) argument(spec CLIArg, args map[string]any) ([]ArgvPart, error) {
	value, present := args[spec.Key]
	switch {
	case !present && spec.Required:
		return nil, fmt.Errorf("%w: %s needs %q", ErrBadArguments, a.Name, spec.Key)
	case !present:
		return nil, nil
	}
	parts, err := spec.parts(value)
	if err != nil {
		return nil, fmt.Errorf("%w: %s %q: %w", ErrBadArguments, a.Name, spec.Key, err)
	}
	return parts, nil
}

func (spec CLIArg) parts(value any) ([]ArgvPart, error) {
	if spec.Shape == ShapeFlag {
		return spec.renderSwitch(value)
	}
	rendered, err := spec.render(value)
	if err != nil {
		return nil, err
	}
	return append(literals(flagOf(spec)), rendered), nil
}

func (a Action) closedArguments(args map[string]any) error {
	for _, key := range slices.Sorted(maps.Keys(args)) {
		if !a.takes(key) {
			return fmt.Errorf("%w: %s takes no argument %q", ErrBadArguments, a.Name, key)
		}
	}
	return nil
}

func (a Action) takes(key string) bool {
	if a.Etag && key == etagKey {
		return true
	}
	for _, spec := range a.CLIArgs {
		if spec.Key == key {
			return true
		}
	}
	return false
}

func renderNumber(value any) (ArgvPart, error) {
	switch number := value.(type) {
	case float64:
		return ArgvPart{Literal: strconv.FormatFloat(number, 'f', -1, 64)}, nil
	case json.Number:
		if _, err := number.Float64(); err == nil {
			return ArgvPart{Literal: number.String()}, nil
		}
	}
	return ArgvPart{}, fmt.Errorf("want a number, got %v", value)
}

func (spec CLIArg) render(value any) (ArgvPart, error) {
	switch spec.Shape {
	case ShapeNum:
		return renderNumber(value)
	case ShapePath, ShapeRun:
		return spec.renderPath(value)
	case ShapeText:
		return spec.renderText(value)
	}
	return ArgvPart{}, fmt.Errorf("unknown shape %q", spec.Shape)
}
