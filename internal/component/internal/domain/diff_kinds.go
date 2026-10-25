package domain

import (
	"fmt"
	"reflect"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func diffKinds(a, b *diffSide, tally *UnchangedSummary) []KindDiff {
	left, right := a.views(), b.views()
	var kinds []KindDiff
	for _, name := range kindNames(left, right) {
		kinds = append(kinds, diffOneKind(name, left, right, a, b, tally)...)
	}
	return kinds
}

func diffOneKind(name string, left, right map[string]kindView, a, b *diffSide, tally *UnchangedSummary) []KindDiff {
	before, hadBefore := left[name]
	after, hasAfter := right[name]
	switch {
	case !hadBefore:
		return []KindDiff{wholeKind(name, StatusAdded, kindView{}, after, b.spec)}
	case !hasAfter:
		return []KindDiff{wholeKind(name, StatusRemoved, before, kindView{}, a.spec)}
	}
	return diffKind(name, before, after, b.spec, tally)
}

func wholeKind(name, status string, before, after kindView, spec *model.DomainSpec) KindDiff {
	code, shape := CodeKindAdded, after
	if status == StatusRemoved {
		code, shape = CodeKindRemoved, before
	}
	change := newChange(code, name, SectionKind, "kind", nil, nil)
	counts := map[string]any{"readings": len(shape.kind.Channels), "dynamics": len(shape.kind.Dynamics), "commands": len(shape.kind.Effectors),
		"faults": len(shape.kind.Faults), "states": len(shape.kind.State)}
	change.Before, change.After = pickBy(status == StatusRemoved, counts), pickBy(status == StatusAdded, counts)
	change.Sentence = fmt.Sprintf("kind %s %s with %d instance(s).", name, status, summaryOf(shape.instances).Count)
	return KindDiff{Kind: name, Status: status, Changes: []Change{change}, Relations: relationsOf(spec, name),
		Instances: KindInstances{Before: summaryOf(before.instances), After: summaryOf(after.instances)}}
}

func pickBy(keep bool, value map[string]any) any {
	if keep {
		return value
	}
	return nil
}

func diffKind(name string, before, after kindView, spec *model.DomainSpec, tally *UnchangedSummary) []KindDiff {
	changes := diffSections(name, before, after, tally)
	changes = append(changes, diffInstances(name, before, after, tally)...)
	changes = append(changes, diffMembers(name, before, after)...)
	if len(changes) == 0 {
		tally.Kinds++
		return nil
	}
	return []KindDiff{{Kind: name, Status: StatusChanged, Changes: changes, Relations: relationsOf(spec, name),
		Instances: KindInstances{Before: summaryOf(before.instances), After: summaryOf(after.instances)}}}
}

func diffSections(name string, before, after kindView, tally *UnchangedSummary) []Change {
	var changes []Change
	left, right := before.groups(), after.groups()
	counters := []*int{&tally.Readings, &tally.Dynamics, &tally.Commands, &tally.Faults, &tally.States}
	for i := range left {
		found, unchanged := diffGroup(name, left[i], right[i])
		changes = append(changes, found...)
		*counters[i] += unchanged
	}
	return changes
}

func diffInstances(name string, before, after kindView, tally *UnchangedSummary) []Change {
	left, right := generic(before.instances), generic(after.instances)
	if reflect.DeepEqual(left, right) {
		tally.InstanceSets += len(before.instances)
		return nil
	}
	change := newChange(CodeInstancesChanged, name+"/"+SectionInstances, SectionInstances, SectionInstances, left, right)
	was, now := summaryOf(before.instances), summaryOf(after.instances)
	change.Sentence = fmt.Sprintf("%s/instances: %d instance(s) %v became %d %v.", name, was.Count, namesOf(was), now.Count, namesOf(now))
	return []Change{change}
}

func namesOf(summary InstanceSummary) []string {
	return slices.Concat(summary.Names, summary.IDTemplates)
}

func summaryOf(sets []model.InstanceSet) InstanceSummary {
	summary := InstanceSummary{Names: []string{}, IDTemplates: []string{}}
	for _, set := range sets {
		if len(set.Names) > 0 {
			summary.Count += len(set.Names)
			summary.Names = append(summary.Names, set.Names...)
			continue
		}
		summary.Count += set.Count.Default
		summary.IDTemplates = append(summary.IDTemplates, set.IDTemplate)
	}
	return summary
}

var coveredMembers = []string{"name", "state", "dynamics", "channels", "effectors", "faults"}

func diffMembers(name string, before, after kindView) []Change {
	return memberChanges(genericMap(before.kind), genericMap(after.kind), coveredMembers, CodeKindChanged, SectionKind, name+"/"+SectionKind+"/")
}
