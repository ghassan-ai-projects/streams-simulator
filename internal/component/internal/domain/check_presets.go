package domain

import (
	"maps"
	"math"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonvalue"
)

func (l *Library) checkPresets(e *Equipment, issues *issueList) {
	if e.DefaultPreset != "" {
		if _, found := e.Presets[e.DefaultPreset]; !found {
			issues.addf("/default_preset", "default preset %q is not one of the presets", e.DefaultPreset)
		}
	}
	claimed := parameterTargets(e)
	for _, name := range slices.Sorted(maps.Keys(e.Presets)) {
		checkPreset(e, name, claimed, issues)
	}
}

func checkPreset(e *Equipment, name string, claimed map[string]string, issues *issueList) {
	workspace := newWorkspace(e)
	values := e.Presets[name].Values
	for _, target := range slices.Sorted(maps.Keys(values)) {
		at := pointer("presets", name, "values", target)
		if err := workspace.write(target, values[target]); err != nil {
			issues.addf(at, "%v", err)
		} else if parameter, taken := claimed[target]; taken {
			issues.addf(at, "target is also written by parameter %q", parameter)
		}
	}
}

func parameterTargets(e *Equipment) map[string]string {
	claimed := map[string]string{}
	for _, parameter := range e.Parameters {
		for _, target := range parameter.Targets {
			claimed[target] = parameter.Name
		}
		for _, name := range slices.Sorted(maps.Keys(parameter.Cases)) {
			for _, target := range slices.Sorted(maps.Keys(parameter.Cases[name])) {
				claimed[target] = parameter.Name
			}
		}
	}
	return claimed
}

func (l *Library) checkParameters(e *Equipment, issues *issueList) {
	for i, parameter := range e.Parameters {
		at := pointer("parameters", indexPart(i))
		if parameter.Type == "enum" {
			checkEnumParameter(e, parameter, at, issues)
		} else {
			checkNumericParameter(e, parameter, at, issues)
		}
	}
}

func checkNumericParameter(e *Equipment, parameter Parameter, at string, issues *issueList) {
	value, numeric := jsonvalue.Number(parameter.Default)
	switch {
	case !numeric:
		issues.addf(at+"/default", "default of %s parameter %q must be a number", parameter.Type, parameter.Name)
	case !withinBounds(parameter, value):
		issues.addf(at+"/default", "default %v is outside the bounds of parameter %q", value, parameter.Name)
	case parameter.Type == "integer" && value != math.Trunc(value):
		issues.addf(at+"/default", "default %v of integer parameter %q is not whole", value, parameter.Name)
	}
	checkTargets(e, parameter, at, issues)
}

func checkTargets(e *Equipment, parameter Parameter, at string, issues *issueList) {
	workspace := newWorkspace(e)
	for _, target := range parameter.Targets {
		if err := workspace.write(target, parameter.Default); err != nil {
			issues.addf(at+"/targets", "%v", err)
		}
	}
}

func withinBounds(parameter Parameter, value float64) bool {
	return (parameter.Min == nil || value >= *parameter.Min) && (parameter.Max == nil || value <= *parameter.Max)
}

func checkEnumParameter(e *Equipment, parameter Parameter, at string, issues *issueList) {
	chosen, _ := parameter.Default.(string)
	if !slices.Contains(parameter.Values, chosen) {
		issues.addf(at+"/default", "default %v is not one of the values of parameter %q", parameter.Default, parameter.Name)
	}
	cases := slices.Sorted(maps.Keys(parameter.Cases))
	if !slices.Equal(cases, slices.Sorted(slices.Values(parameter.Values))) {
		issues.addf(at+"/cases", "cases %v differ from values %v of parameter %q", cases, parameter.Values, parameter.Name)
	}
	for _, value := range cases {
		checkCase(e, parameter, value, at, issues)
	}
}

func checkCase(e *Equipment, parameter Parameter, value, at string, issues *issueList) {
	workspace := newWorkspace(e)
	targets := parameter.Cases[value]
	for _, target := range slices.Sorted(maps.Keys(targets)) {
		if err := workspace.write(target, targets[target]); err != nil {
			issues.addf(at+"/cases/"+value, "%v", err)
		}
	}
}
