package domain

import (
	"maps"
	"slices"
	"strings"
)

var noChoiceHint = map[string]string{"commands": " (a command that is a safe stop has a fixed response)"}

func (t *twinBuild) loadKinds(files map[string]map[string]any) error {
	for i, choice := range t.choices.Kinds {
		k, err := t.loadKind(i, choice, files)
		if err != nil {
			return err
		}
		t.kinds = append(t.kinds, k)
		t.byName[choice.Kind] = k
	}
	return nil
}

func (t *twinBuild) loadKind(i int, choice kindChoice, files map[string]map[string]any) (*twinKind, error) {
	at := pointer("kinds", indexPart(i), "kind")
	if t.byName[choice.Kind] != nil {
		return nil, t.refuse(ErrNameCollision, at, "kind %q is chosen twice", choice.Kind)
	}
	doc, found := files[choice.Kind]
	if !found {
		return nil, t.refuse(ErrNoKindFile, at, "no kind file defines kind %q (%d kind files loaded)", choice.Kind, len(files))
	}
	file, issues := readKind(doc)
	if len(issues) > 0 {
		return nil, t.refuseKindFile(choice.Kind, at, issues)
	}
	return &twinKind{index: i, choice: choice, file: file, actuated: map[string]bool{}}, nil
}

func (t *twinBuild) checkInstances() error {
	seen := map[string]bool{}
	for i, choice := range t.choices.Kinds {
		if seen[choice.Instance] {
			return t.refuse(ErrNameCollision, pointer("kinds", indexPart(i), "instance"), "instance %q is named by two kinds", choice.Instance)
		}
		seen[choice.Instance] = true
	}
	return nil
}

func (t *twinBuild) planKinds() error {
	for _, k := range t.kinds {
		for _, step := range []func(*twinKind) error{t.planMetrics, t.planCommands, t.checkActuators} {
			if err := step(k); err != nil {
				return err
			}
		}
	}
	return t.checkDistinctNames()
}

func (t *twinBuild) checkDistinctNames() error {
	seen := map[string]string{}
	for _, k := range t.kinds {
		for _, name := range k.names() {
			if first, taken := seen[name]; taken {
				return t.refuse(ErrNameCollision, pointer("kinds", indexPart(k.index), "kind"), "kind %q and kind %q both declare %s: names are unique across a domain", first, k.file.Name, name)
			}
			seen[name] = k.file.Name
		}
	}
	return nil
}

func (k *twinKind) names() []string {
	names := make([]string, 0, len(k.file.Metrics)+len(k.file.Commands))
	for _, metric := range k.file.Metrics {
		names = append(names, "metric "+metric.Name)
	}
	for _, command := range k.file.Commands {
		names = append(names, "command "+command.Name)
	}
	return names
}

func (t *twinBuild) requireDecided(k *twinKind, member string, want, have []string) error {
	for _, name := range have {
		if !slices.Contains(want, name) {
			return t.refuse(ErrUnknownChoice, pointer("kinds", indexPart(k.index), member, name),
				"kind %q: the kind file asks no choice for %q%s", k.file.Name, name, noChoiceHint[member])
		}
	}
	if missing := slices.DeleteFunc(slices.Clone(want), func(name string) bool { return slices.Contains(have, name) }); len(missing) > 0 {
		return t.refuse(ErrIncompleteChoices, pointer("kinds", indexPart(k.index), member), "kind %q: choose for %s", k.file.Name, strings.Join(missing, ", "))
	}
	return nil
}

func (t *twinBuild) planMetrics(k *twinKind) error {
	if err := t.requireDecided(k, "metrics", k.file.metricNames(), slices.Sorted(maps.Keys(k.choice.Metrics))); err != nil {
		return err
	}
	for _, metric := range k.file.Metrics {
		plan, err := t.planMetric(k, metric)
		if err != nil {
			return err
		}
		k.metrics = append(k.metrics, plan)
	}
	return nil
}

func (t *twinBuild) planMetric(k *twinKind, metric kindMetric) (*metricPlan, error) {
	plan := &metricPlan{file: metric, choice: k.choice.Metrics[metric.Name]}
	at := pointer("kinds", indexPart(k.index), "metrics", metric.Name)
	if err := t.checkInitial(plan, at+"/initial"); err != nil {
		return nil, err
	}
	if err := t.attachInstrument(plan, at+"/instrument"); err != nil {
		return nil, err
	}
	return plan, nil
}

func (t *twinBuild) checkInitial(plan *metricPlan, at string) error {
	value, eu := plan.choice.Initial, plan.file.EURange
	if value == nil || eu == nil || (*value >= eu.Min && *value <= eu.Max) {
		return nil
	}
	return t.refuse(ErrValueOutOfRange, at, "initial %v lies outside the eu_range %v to %v of metric %q", *value, eu.Min, eu.Max, plan.file.Name)
}
