package domain

import (
	"maps"
	"slices"
)

const (
	twinStresses = "A twin scaffolded from kind files: every metric reads through a library instrument or a plain channel and every command answers by a fixed response, so a consumer meets the device's interface before the device exists."
	plainPeriodS = 5
	stepMS       = 1000
)

type linkPair struct{ driver, driven *twinKind }

func (t *twinBuild) document() map[string]any {
	kinds := t.kindDocuments()
	document := map[string]any{"id": t.choices.ID, "version": t.choices.Version, "title": t.choices.Title, "axes": t.axes(),
		"stresses": twinStresses, "kinds": kinds, "instances": t.instances(), "profiles": twinProfiles(kinds),
		"ground_truth": map[string]any{"negative_class_fraction": 0.4, "pre_degraded_fraction": 0.1, "min_scenarios": 100, "min_per_label": 10}}
	t.addOptional(document)
	return document
}

func (t *twinBuild) addOptional(document map[string]any) {
	if t.choices.Description != "" {
		document["description"] = t.choices.Description
	}
	if t.choices.Site != nil {
		document["site"] = t.site()
	}
	if t.choices.Board != nil {
		document["boards"] = []any{map[string]any{"name": t.choices.Board.Name, "members": t.instanceNames()}}
	}
	if pairs := t.linkPairs(); len(pairs) > 0 {
		document["link_types"], document["links"] = linkTypes(pairs), links(pairs)
	}
}

func (t *twinBuild) site() map[string]any {
	site := map[string]any{"name": t.choices.Site.Name}
	if t.choices.Site.Area != "" {
		site["areas"] = []any{map[string]any{"name": t.choices.Site.Area}}
	}
	return site
}

func (t *twinBuild) instanceNames() []any {
	names := make([]string, len(t.kinds))
	for i, k := range t.kinds {
		names[i] = k.choice.Instance
	}
	return stringList(names)
}

func (t *twinBuild) instances() []any {
	sets := make([]any, len(t.kinds))
	for i, k := range t.kinds {
		set := map[string]any{"kind": k.file.Name, "names": []any{k.choice.Instance}}
		if t.choices.Site != nil && t.choices.Site.Area != "" {
			set["area"] = t.choices.Site.Area
		}
		sets[i] = set
	}
	return sets
}

func twinProfiles(kinds []any) []any {
	weights := map[string]any{}
	for _, id := range faultIDs(kinds) {
		weights[id] = 1
	}
	return []any{
		map[string]any{"name": "nominal", "duration_s": 3600, "description": "The twin at rest: no fault and no command."},
		map[string]any{"name": "correlated_cascade", "description": "Not applicable.",
			"not_applicable": "A scaffolded twin has no peers that one cause trips together, so there is no alarm flood to reproduce."},
		map[string]any{"name": "sensor_pathology", "duration_s": 3600, "fault_weights": weights,
			"description": "A reading goes wrong while the world keeps its real value."},
	}
}

func faultIDs(kinds []any) []string {
	var ids []string
	for _, kind := range kinds {
		for _, fault := range members(kind.(map[string]any), "faults") {
			ids = append(ids, fault["id"].(string))
		}
	}
	return ids
}

func (t *twinBuild) linkPairs() []linkPair {
	var pairs []linkPair
	for _, k := range t.kinds {
		for _, driver := range k.drivers() {
			if pair := (linkPair{driver, k}); !slices.Contains(pairs, pair) {
				pairs = append(pairs, pair)
			}
		}
	}
	return pairs
}

func (k *twinKind) drivers() []*twinKind {
	var drivers []*twinKind
	for _, plan := range k.metrics {
		for _, c := range plan.driven {
			if c.driver != k {
				drivers = append(drivers, c.driver)
			}
		}
	}
	return drivers
}

func linkTypeName(driver, driven *twinKind) string {
	return driver.file.Name + "_drives_" + driven.file.Name
}

func linkTypes(pairs []linkPair) []any {
	types := make([]any, len(pairs))
	for i, pair := range pairs {
		types[i] = map[string]any{"name": linkTypeName(pair.driver, pair.driven), "from": pair.driver.file.Name, "to": pair.driven.file.Name}
	}
	return types
}

func links(pairs []linkPair) []any {
	out := make([]any, len(pairs))
	for i, pair := range pairs {
		out[i] = map[string]any{"type": linkTypeName(pair.driver, pair.driven), "from": pair.driver.choice.Instance, "to": pair.driven.choice.Instance}
	}
	return out
}

func (t *twinBuild) axes() map[string]any {
	return map[string]any{"rate": "low", "cardinality": "small", "value_shape": t.valueShapes(), "cadence": []any{"periodic"},
		"lateness": "bounded", "absence": "signal", "time_reference": "wall", "correlation": t.correlation(),
		"seasonality": []any{"none"}, "actuation": t.actuation(), "consequence": []any{"cost"}, "fidelity": []any{"F1"}}
}

func (t *twinBuild) valueShapes() []any {
	shapes := map[string]bool{}
	for _, k := range t.kinds {
		for _, plan := range k.metrics {
			shapes[shapeOfType[plan.file.DataType]] = true
		}
	}
	return stringList(slices.Sorted(maps.Keys(shapes)))
}

func (t *twinBuild) correlation() []any {
	if t.coupled() {
		return []any{"coupled"}
	}
	return []any{"independent"}
}

func (t *twinBuild) coupled() bool {
	return slices.ContainsFunc(t.kinds, func(k *twinKind) bool {
		return slices.ContainsFunc(k.metrics, func(plan *metricPlan) bool { return len(plan.driven) > 0 })
	})
}

func (t *twinBuild) actuation() string {
	if slices.ContainsFunc(t.kinds, func(k *twinKind) bool { return len(k.commands) > 0 }) {
		return "fast_physical"
	}
	return "observe_only"
}

var shapeOfType = map[string]string{"number": "scalar", "integer": "counter", "boolean": "boolean"}
