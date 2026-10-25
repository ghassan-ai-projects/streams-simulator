package domain

import (
	"encoding/json"
	"maps"
	"reflect"
	"slices"
)

type PlanDiff struct {
	Faults        Diff[FaultItem]        `json:"faults"`
	Perturbations Diff[PerturbationItem] `json:"perturbations"`
	Environment   Diff[EnvironmentItem]  `json:"environment"`
	EffectorCalls Diff[CallItem]         `json:"effector_calls"`
}

type Diff[T any] struct {
	Identical []T          `json:"identical"`
	Different []Changed[T] `json:"different"`
	OnlyInA   []T          `json:"only_in_a"`
	OnlyInB   []T          `json:"only_in_b"`
}

type Changed[T any] struct {
	A           T        `json:"a"`
	B           T        `json:"b"`
	Differences []string `json:"differences"`
}

type item interface{ key() string }

type pairing struct {
	partner []int
	same    []bool
	taken   []bool
}

func diffItems[T item](a, b []T) Diff[T] {
	p := pairing{partner: slices.Repeat([]int{-1}, len(a)), same: make([]bool, len(a)), taken: make([]bool, len(b))}
	link(&p, a, b, func(x, y T) bool { return encoded(x) == encoded(y) }, true)
	link(&p, a, b, func(x, y T) bool { return x.key() == y.key() }, false)
	return gatherPairs(a, b, p)
}

func link[T item](p *pairing, a, b []T, match func(x, y T) bool, same bool) {
	for i := range a {
		if p.partner[i] >= 0 {
			continue
		}
		for j := range b {
			if !p.taken[j] && match(a[i], b[j]) {
				p.partner[i], p.taken[j], p.same[i] = j, true, same
				break
			}
		}
	}
}

func gatherPairs[T item](a, b []T, p pairing) Diff[T] {
	diff := Diff[T]{Identical: []T{}, Different: []Changed[T]{}, OnlyInA: []T{}, OnlyInB: []T{}}
	for i, x := range a {
		switch {
		case p.partner[i] < 0:
			diff.OnlyInA = append(diff.OnlyInA, x)
		case p.same[i]:
			diff.Identical = append(diff.Identical, x)
		default:
			diff.Different = append(diff.Different, changedPair(x, b[p.partner[i]]))
		}
	}
	diff.OnlyInB = untaken(b, p.taken)
	return diff
}

func changedPair[T any](a, b T) Changed[T] {
	return Changed[T]{A: a, B: b, Differences: changedFields(a, b)}
}

func untaken[T any](items []T, taken []bool) []T {
	rest := []T{}
	for j, item := range items {
		if !taken[j] {
			rest = append(rest, item)
		}
	}
	return rest
}

func encoded(value any) string {
	raw, _ := json.Marshal(value)
	return string(raw)
}

func changedFields(a, b any) []string {
	left, right := fieldsOf(a), fieldsOf(b)
	names := slices.Concat(slices.Collect(maps.Keys(left)), slices.Collect(maps.Keys(right)))
	slices.Sort(names)
	return slices.DeleteFunc(slices.Compact(names), func(name string) bool { return reflect.DeepEqual(left[name], right[name]) })
}

func fieldsOf(value any) map[string]any {
	fields := map[string]any{}
	_ = json.Unmarshal([]byte(encoded(value)), &fields)
	return fields
}
