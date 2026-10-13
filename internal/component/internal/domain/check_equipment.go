package domain

import (
	"regexp"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

var qualifiable = regexp.MustCompile(`^[a-z][a-z0-9]*(_[a-z0-9]+)*$`)

const (
	maxKindName   = 20
	maxMemberName = 40
)

type kindIndex struct {
	states    map[string]string
	effectors map[string]bool
	points    map[string]bool
	faults    []map[string]any
	effs      []map[string]any
}

func newKindIndex(e *Equipment) *kindIndex {
	kind, _ := e.Document["kind"].(map[string]any)
	index := &kindIndex{states: map[string]string{}, effectors: map[string]bool{}, points: map[string]bool{},
		faults: members(kind, "faults"), effs: members(kind, "effectors")}
	for _, state := range members(kind, "state") {
		name, _ := state["name"].(string)
		index.states[name], _ = state["unit"].(string)
	}
	index.effectors = nameSet(index.effs)
	for _, point := range e.MeasurementPoints {
		index.points[point.Name] = true
	}
	return index
}

func nameSet(items []map[string]any) map[string]bool {
	names := map[string]bool{}
	for _, item := range items {
		name, _ := item["name"].(string)
		names[name] = true
	}
	return names
}

func members(kind map[string]any, section string) []map[string]any {
	items, _ := kind[section].([]any)
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		entry, _ := item.(map[string]any)
		out = append(out, entry)
	}
	return out
}

func defaultKindName(id string) string {
	_, name, _ := strings.Cut(id, "/")
	return strings.ReplaceAll(name, "-", "_")
}

func (l *Library) checkEquipment(e *Equipment) error {
	var issues issueList
	checkSections(mapOf(e.Document["kind"]), &issues)
	if issues.empty() {
		l.checkMembers(e, &issues)
	}
	return issues.violation(model.ContractComponent, e.File)
}

func (l *Library) checkMembers(e *Equipment, issues *issueList) {
	index := newKindIndex(e)
	l.checkNames(e, index, issues)
	l.checkPoints(e, index, issues)
	l.checkPorts(e, index, issues)
	checkReferences(mapOf(e.Document["kind"]), index, issues)
	l.checkPresets(e, issues)
	l.checkParameters(e, issues)
}

var kindSections = []string{"state", "dynamics", "flows", "faults", "effectors", "boundaries"}

func checkSections(kind map[string]any, issues *issueList) {
	for _, section := range kindSections {
		items, _ := kind[section].([]any)
		for i, item := range items {
			if _, object := item.(map[string]any); !object {
				issues.addf(pointer("kind", section, indexPart(i)), "must be an object")
			}
		}
	}
}

func (l *Library) checkNames(e *Equipment, index *kindIndex, issues *issueList) {
	if !qualifiable.MatchString(defaultKindName(e.ID)) {
		issues.addf("/id", "the name of %s, with hyphens as underscores, must match %s", e.ID, qualifiable)
	}
	checkMemberNames(index.faults, "faults", "id", issues)
	checkMemberNames(index.effs, "effectors", "name", issues)
	checkUnique(e, issues)
}

func checkMemberNames(items []map[string]any, section, key string, issues *issueList) {
	for i, item := range items {
		name, _ := item[key].(string)
		if !qualifiable.MatchString(name) || len(name) > maxMemberName {
			issues.addf(pointer("kind", section, indexPart(i), key), "%q must match %s and be at most %d characters", name, qualifiable, maxMemberName)
		}
	}
}

func checkUnique(e *Equipment, issues *issueList) {
	var points, ports, parameters []string
	for _, point := range e.MeasurementPoints {
		points = append(points, point.Name)
	}
	for _, port := range e.Ports {
		ports = append(ports, port.Name)
	}
	for _, parameter := range e.Parameters {
		parameters = append(parameters, parameter.Name)
	}
	checkDistinct(points, "measurement_points", "name", "measurement point", issues)
	checkDistinct(ports, "ports", "name", "port", issues)
	checkDistinct(parameters, "parameters", "name", "parameter", issues)
}

func checkDistinct(names []string, section, key, noun string, issues *issueList) {
	seen := map[string]bool{}
	for i, name := range names {
		if seen[name] {
			issues.addf(pointer(section, indexPart(i), key), "%s %q is declared twice", noun, name)
		}
		seen[name] = true
	}
}

func (l *Library) checkPoints(e *Equipment, index *kindIndex, issues *issueList) {
	for i, point := range e.MeasurementPoints {
		at := func(member string) string { return pointer("measurement_points", indexPart(i), member) }
		l.checkPointUnits(point, index, at, issues)
		l.checkPointInstrument(point, at, issues)
		if point.Range.Min >= point.Range.Max {
			issues.addf(at("range"), "range min %v must be below max %v", point.Range.Min, point.Range.Max)
		}
	}
}

func (l *Library) checkPointUnits(point MeasurementPoint, index *kindIndex, at func(string) string, issues *issueList) {
	l.checkStateUnit(point, index, at("state"), issues)
	l.checkTag(point.Tag, point.UnitFamily, at, issues)
}

func (l *Library) checkStateUnit(point MeasurementPoint, index *kindIndex, at string, issues *issueList) {
	unit, declared := index.states[point.State]
	switch {
	case !declared:
		issues.addf(at, "state %q is not declared by the kind", point.State)
	case !l.vocabulary.hasUnit(point.UnitFamily, unit):
		issues.addf(at, "state %q has unit %q, which is not a unit of family %q", point.State, unit, point.UnitFamily)
	}
}

func (l *Library) checkTag(tag, family string, at func(string) string, issues *issueList) {
	if _, known := l.vocabulary.UnitFamilies[family]; !known {
		issues.addf(at("unit_family"), "unit family %q is not in the vocabulary", family)
	}
	if _, known := l.vocabulary.TagLetters[tag]; !known {
		issues.addf(at("tag"), "tag %q is not in the vocabulary", tag)
	} else if !l.vocabulary.tagMeasures(tag, family) {
		issues.addf(at("tag"), "tag %q does not measure unit family %q", tag, family)
	}
}

func (l *Library) checkPointInstrument(point MeasurementPoint, at func(string) string, issues *issueList) {
	instrument, found := l.instruments[point.DefaultInstrument]
	switch {
	case !found:
		issues.addf(at("default_instrument"), "default instrument %s is not in the library", point.DefaultInstrument)
	case instrument.Tag != point.Tag || instrument.UnitFamily != point.UnitFamily:
		issues.addf(at("default_instrument"), "default instrument %s measures %s/%s, the point %s/%s",
			point.DefaultInstrument, instrument.Tag, instrument.UnitFamily, point.Tag, point.UnitFamily)
	}
}

func (l *Library) checkPorts(e *Equipment, index *kindIndex, issues *issueList) {
	for i, port := range e.Ports {
		l.checkPort(port, index, pointer("ports", indexPart(i)), issues)
	}
}

func (l *Library) checkPort(port Port, index *kindIndex, at string, issues *issueList) {
	portType, known := l.vocabulary.PortTypes[port.Type]
	unit, declared := index.states[port.State]
	switch {
	case !known:
		issues.addf(at+"/type", "port type %q is not in the vocabulary", port.Type)
	case !declared:
		issues.addf(at+"/state", "state %q is not declared by the kind", port.State)
	case !l.vocabulary.hasUnit(portType.UnitFamily, unit):
		issues.addf(at+"/state", "state %q has unit %q, which is not a unit of family %q fixed by port type %q", port.State, unit, portType.UnitFamily, port.Type)
	}
}
