package domain

import (
	"maps"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

var ignoredFields = []string{"basis", "recorded_in", "equals_binary_suggestion", "statement"}

func newEntry(key string, view any, raw any, nested []string, groups ...group) entry {
	fields := genericMap(view)
	for _, name := range ignoredFields {
		delete(fields, name)
	}
	for _, name := range nested {
		delete(fields, name)
	}
	flatten(fields, "effect", "acknowledgement")
	return entry{key: key, fields: fields, raw: generic(raw), groups: groups}
}

func flatten(fields map[string]any, names ...string) {
	for _, name := range names {
		inner, isObject := fields[name].(map[string]any)
		if !isObject {
			continue
		}
		delete(fields, name)
		for _, key := range slices.Sorted(maps.Keys(inner)) {
			fields[name+"."+key] = inner[key]
		}
	}
}

func (v kindView) groups() []group {
	return []group{v.readings(), v.dynamics(), v.commands(), v.faults(), v.states()}
}

func (v kindView) readings() group {
	shape := group{noun: "reading", segment: SectionReadings, section: SectionReadings}
	for i, reading := range v.assumed.Readings {
		shape.entries = append(shape.entries, newEntry(reading.Channel, reading, v.kind.Channels[i], nil))
	}
	return shape
}

func (v kindView) dynamics() group {
	shape := group{noun: "dynamics", segment: SectionDynamics, section: SectionDynamics}
	for i, dynamics := range v.assumed.Dynamics {
		shape.entries = append(shape.entries, newEntry(dynamics.Target, dynamics, v.kind.Dynamics[i], []string{"inputs"}, inputGroup(dynamics.Inputs)))
	}
	return shape
}

func inputGroup(inputs []InputValue) group {
	shape := group{noun: "input", segment: "inputs", section: SectionDynamics}
	for _, input := range inputs {
		key := input.State
		if input.Link != "" {
			key += "@" + input.Link
		}
		shape.entries = append(shape.entries, newEntry(key, input, input, nil))
	}
	return shape
}

func (v kindView) commands() group {
	shape := group{noun: "command", segment: SectionCommands, section: SectionCommands}
	for i, command := range v.assumed.Commands {
		shape.entries = append(shape.entries, newEntry(command.Name, command, v.kind.Effectors[i], []string{"arguments"}, argumentGroup(command.Arguments)))
	}
	return shape
}

func argumentGroup(arguments []ArgumentValue) group {
	shape := group{noun: "argument", segment: "arguments", section: SectionCommands}
	for _, argument := range arguments {
		shape.entries = append(shape.entries, newEntry(argument.Name, argument, argument, nil))
	}
	return shape
}

func (v kindView) faults() group {
	shape := group{noun: "fault", segment: SectionFaults, section: SectionFaults}
	for i, fault := range v.assumed.Faults {
		shape.entries = append(shape.entries, newEntry(fault.ID, fault, v.kind.Faults[i], nil))
	}
	return shape
}

func (v kindView) states() group {
	shape := group{noun: "state", segment: SectionStates, section: SectionStates}
	for _, state := range v.kind.State {
		shape.entries = append(shape.entries, newEntry(state.Name, state, state, nil))
	}
	return shape
}

type kindView struct {
	kind      *model.Kind
	assumed   KindAssumptions
	instances []model.InstanceSet
}

func (s *diffSide) views() map[string]kindView {
	views := make(map[string]kindView, len(s.spec.Kinds))
	for i := range s.spec.Kinds {
		kind := &s.spec.Kinds[i]
		views[kind.Name] = kindView{kind: kind, assumed: s.assumptions.Kinds[i], instances: s.instancesOf(kind.Name)}
	}
	return views
}

func (s *diffSide) instancesOf(kind string) []model.InstanceSet {
	var sets []model.InstanceSet
	for _, set := range s.spec.Instances {
		if set.Kind == kind {
			sets = append(sets, set)
		}
	}
	return sets
}

func kindNames(left, right map[string]kindView) []string {
	return unionKeys(left, right)
}
