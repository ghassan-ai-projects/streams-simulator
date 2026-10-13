package domain

import (
	"cmp"
	"fmt"
	"maps"
	"math"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonvalue"
)

func (b *entryBuild) applySizes() error {
	if err := b.applyPreset(); err != nil {
		return err
	}
	if err := b.checkParameterNames(); err != nil {
		return err
	}
	for _, parameter := range b.equipment.Parameters {
		if err := b.applyParameter(parameter); err != nil {
			return err
		}
	}
	return nil
}

func (b *entryBuild) applyPreset() error {
	b.preset = cmp.Or(b.entry.Preset, b.equipment.DefaultPreset)
	if b.preset == "" {
		return nil
	}
	preset, found := b.equipment.Presets[b.preset]
	if !found {
		return b.refuse("preset", "%s has no preset %q (presets: %v)", b.equipment.ID, b.preset, slices.Sorted(maps.Keys(b.equipment.Presets)))
	}
	return b.writeAll(preset.Values, "preset")
}

func (b *entryBuild) writeAll(values map[string]any, member string) error {
	for _, target := range slices.Sorted(maps.Keys(values)) {
		if err := b.ws.write(target, values[target]); err != nil {
			return b.refuse(member, "%v", err)
		}
	}
	return nil
}

func (b *entryBuild) checkParameterNames() error {
	declared := map[string]bool{}
	for _, parameter := range b.equipment.Parameters {
		declared[parameter.Name] = true
	}
	for _, name := range slices.Sorted(maps.Keys(b.entry.Parameters)) {
		if !declared[name] {
			return b.refuse(relative("parameters", name), "%s has no parameter %q", b.equipment.ID, name)
		}
	}
	return nil
}

func (b *entryBuild) applyParameter(parameter Parameter) error {
	value, given := b.entry.Parameters[parameter.Name]
	if !given {
		value = parameter.Default
	}
	if err := b.admitValue(parameter, value); err != nil {
		return err
	}
	b.parameters[parameter.Name] = value
	if parameter.Type == "enum" {
		return b.writeAll(parameter.Cases[fmt.Sprint(value)], relative("parameters", parameter.Name))
	}
	return b.writeAll(targetValues(parameter, value), relative("parameters", parameter.Name))
}

func targetValues(parameter Parameter, value any) map[string]any {
	values := make(map[string]any, len(parameter.Targets))
	for _, target := range parameter.Targets {
		values[target] = value
	}
	return values
}

func (b *entryBuild) admitChoice(parameter Parameter, value any, at string) error {
	if text, _ := value.(string); !slices.Contains(parameter.Values, text) {
		return b.refuse(at, "%v is not one of %v for parameter %q", value, parameter.Values, parameter.Name)
	}
	return nil
}

func (b *entryBuild) admitValue(parameter Parameter, value any) error {
	at := relative("parameters", parameter.Name)
	if parameter.Type == "enum" {
		return b.admitChoice(parameter, value, at)
	}
	number, numeric := jsonvalue.Number(value)
	switch {
	case !numeric:
		return b.refuse(at, "parameter %q takes a number, got %v", parameter.Name, value)
	case !withinBounds(parameter, number):
		return b.refuse(at, "%v is outside the bounds of parameter %q", number, parameter.Name)
	}
	return b.admitWhole(parameter, number, at)
}

func (b *entryBuild) admitWhole(parameter Parameter, number float64, at string) error {
	if parameter.Type == "integer" && number != math.Trunc(number) {
		return b.refuse(at, "parameter %q takes a whole number, got %v", parameter.Name, number)
	}
	return nil
}
