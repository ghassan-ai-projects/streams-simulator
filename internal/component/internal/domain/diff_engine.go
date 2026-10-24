package domain

import (
	"encoding/json"
	"maps"
	"reflect"
	"slices"
	"strings"
)

type entry struct {
	key    string
	fields map[string]any
	raw    any
	groups []group
}

type group struct {
	noun, segment, section string
	entries                []entry
}

func (g group) byKey() map[string]entry {
	index := make(map[string]entry, len(g.entries))
	for _, one := range g.entries {
		index[one.key] = one
	}
	return index
}

func diffGroup(parent string, before, after group) ([]Change, int) {
	left, right := before.byKey(), after.byKey()
	var changes []Change
	unchanged := 0
	for _, key := range unionKeys(left, right) {
		found := diffPair(subjectOf(parent, before, key), after, left, right, key)
		if len(found) == 0 {
			unchanged++
		}
		changes = append(changes, found...)
	}
	return changes, unchanged
}

func unionKeys[V any](left, right map[string]V) []string {
	union := maps.Clone(left)
	maps.Copy(union, right)
	return slices.Sorted(maps.Keys(union))
}

func subjectOf(parent string, shape group, key string) string {
	if parent == "" {
		return shape.segment + "/" + key
	}
	return parent + "/" + shape.segment + "/" + key
}

func diffPair(subject string, shape group, left, right map[string]entry, key string) []Change {
	before, hadBefore := left[key]
	after, hasAfter := right[key]
	switch {
	case !hadBefore:
		return []Change{newChange(shape.noun+"_added", subject, shape.section, shape.noun, nil, after.fields)}
	case !hasAfter:
		return []Change{newChange(shape.noun+"_removed", subject, shape.section, shape.noun, before.fields, nil)}
	}
	return diffEntry(shape, subject, before, after)
}

func diffEntry(shape group, subject string, before, after entry) []Change {
	changes := fieldChanges(shape, subject, before.fields, after.fields)
	for i := range before.groups {
		found, _ := diffGroup(subject, before.groups[i], after.groups[i])
		changes = append(changes, found...)
	}
	if len(changes) == 0 && !reflect.DeepEqual(before.raw, after.raw) {
		changes = append(changes, newChange(shape.noun+"_changed", subject, shape.section, "definition", before.raw, after.raw))
	}
	return changes
}

func fieldChanges(shape group, subject string, before, after map[string]any) []Change {
	var changes []Change
	for _, field := range unionKeys(before, after) {
		if !reflect.DeepEqual(before[field], after[field]) {
			changes = append(changes, newChange(fieldCode(shape.noun, field), subject, shape.section, field, before[field], after[field]))
		}
	}
	return changes
}

var fieldCodes = map[string]string{
	"instrument": CodeInstrumentChanged, "unit": CodeUnitChanged, "range": CodeRangeChanged, "resolution": CodeResolutionChanged,
	"noise": CodeNoiseChanged, "cadence": CodeCadenceChanged, "availability": CodeAvailabilityChanged,
	"gain": CodeGainChanged, "coef": CodeGainChanged, "time_constant_s": CodeTimeConstantChanged, "dead_time_s": CodeDeadTimeChanged,
	"clamp": CodeClampChanged, "safe_stop": CodeSafeStopChanged, "risk_class": CodeRiskClassChanged, "sets": CodeSetsChanged,
	"lease": CodeLeaseChanged, "latency_ms_mean": CodeLatencyChanged, "latency_ms_sigma": CodeLatencyChanged,
	"failure_modes": CodeFailureModesChanged, "state_changes": CodeStateChangesChanged,
}

func fieldCode(noun, field string) string {
	_, last, nested := strings.Cut(field, ".")
	if !nested {
		last = field
	}
	if code, named := fieldCodes[last]; named {
		return code
	}
	return noun + "_changed"
}

func generic(value any) any {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	var out any
	if json.Unmarshal(raw, &out) != nil {
		return nil
	}
	return out
}

func genericMap(value any) map[string]any {
	out, _ := generic(value).(map[string]any)
	if out == nil {
		out = map[string]any{}
	}
	return out
}
