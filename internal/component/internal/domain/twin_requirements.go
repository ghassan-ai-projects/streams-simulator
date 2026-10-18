package domain

import (
	"encoding/json"
	"maps"
	"regexp"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

const (
	KeywordMemberRequired    = "twin_member_required"
	KeywordMemberUnsupported = "twin_member_unsupported"
	KeywordNamePattern       = "twin_name_pattern"
	KeywordTypeUnsupported   = "twin_type_unsupported"
	KeywordUnitRequired      = "twin_unit_required"
	KeywordUnitUnexpected    = "twin_unit_unexpected"
	KeywordRange             = "twin_range"
	KeywordRiskClass         = "twin_risk_class"
	KeywordNotObject         = "twin_not_object"
	KeywordDecode            = "twin_decode"
)

var riskClasses = []string{"low", "medium", "high"}

func TwinRequirements(doc map[string]any) []model.Issue {
	var list issueList
	list.unsupported(doc, kindMembers, "")
	requireKind(&list, doc)
	for i, metric := range members(doc, "metrics") {
		requireMetric(&list, metric, pointer("metrics", indexPart(i)))
	}
	for i, command := range members(doc, "commands") {
		requireCommand(&list, command, pointer("commands", indexPart(i)))
	}
	return list.items
}

func (l *issueList) unsupported(item map[string]any, allowed []string, at string) {
	for _, key := range slices.Sorted(maps.Keys(item)) {
		if !slices.Contains(allowed, key) {
			l.rule(at+pointer(key), KeywordMemberUnsupported, key, "is not part of the interface a twin reproduces")
		}
	}
}

func (l *issueList) required(item map[string]any, needed []string, at string) {
	for _, name := range needed {
		if _, present := item[name]; !present {
			l.rule(at+pointer(name), KeywordMemberRequired, name, "needs %q: the interface a twin reproduces always states it", name)
		}
	}
}

func requireKind(list *issueList, doc map[string]any) {
	name, isText := doc["kind"].(string)
	switch {
	case !isText:
		list.rule("/kind", KeywordMemberRequired, "kind", "needs %q: a kind file always names its kind", "kind")
	case !kindNamePattern.MatchString(name):
		list.rule("/kind", KeywordNamePattern, "kind", "kind name %q must match %s", name, kindNamePattern)
	}
	if len(members(doc, "metrics")) == 0 {
		list.rule("/metrics", KeywordMemberRequired, "metrics", "kind %q has no metric: a twin needs a reading to observe", name)
	}
}

func requireMetric(list *issueList, item map[string]any, at string) {
	if item == nil {
		list.rule(at, KeywordNotObject, "", "a metric is an object")
		return
	}
	list.unsupported(item, metricMembers, at)
	list.required(item, metricRequired, at)
	requireName(list, item, at, metricNamePattern, "metric")
	requireMetricValue(list, item, at)
}

func requireName(list *issueList, item map[string]any, at string, pattern *regexp.Regexp, what string) {
	if name, isText := item["name"].(string); isText && !pattern.MatchString(name) {
		list.rule(at+"/name", KeywordNamePattern, "name", "%s name %q must match %s", what, name, pattern)
	}
}

func requireMetricValue(list *issueList, item map[string]any, at string) {
	name, _ := item["name"].(string)
	dataType, _ := item["data_type"].(string)
	unit, _ := item["unit"].(string)
	requireType(list, dataType, unit, name, at)
	requireRange(list, item["eu_range"], name, at)
}

func requireType(list *issueList, dataType, unit, name, at string) {
	switch {
	case dataType == "enum" || dataType == "string":
		list.rule(at+"/data_type", KeywordTypeUnsupported, "data_type", "metric %q is %s: a twin reproduces number, integer and boolean metrics", name, dataType)
	case dataType == "boolean" && unit != "":
		list.rule(at+"/unit", KeywordUnitUnexpected, "unit", "metric %q is boolean: a boolean metric has no unit", name)
	case dataType != "" && dataType != "boolean" && unit == "":
		list.rule(at+"/unit", KeywordUnitRequired, "unit", "metric %q is a number or integer: it states its unit", name)
	}
}

func requireRange(list *issueList, value any, name, at string) {
	bounds, isObject := value.(map[string]any)
	low, lowKnown := floatOf(bounds["min"])
	high, highKnown := floatOf(bounds["max"])
	if isObject && lowKnown && highKnown && low >= high {
		list.rule(at+"/eu_range", KeywordRange, "eu_range", "metric %q: eu_range min %v must be below max %v", name, low, high)
	}
}

func floatOf(value any) (float64, bool) {
	switch number := value.(type) {
	case json.Number:
		parsed, err := number.Float64()
		return parsed, err == nil
	case float64:
		return number, true
	case int:
		return float64(number), true
	}
	return 0, false
}

func requireCommand(list *issueList, item map[string]any, at string) {
	if item == nil {
		list.rule(at, KeywordNotObject, "", "a command is an object")
		return
	}
	list.unsupported(item, commandMembers, at)
	list.required(item, commandMembers, at)
	requireName(list, item, at, commandNamePattern, "command")
	requireRiskClass(list, item, at)
}

func requireRiskClass(list *issueList, item map[string]any, at string) {
	risk, present := item["risk_class"]
	text, isText := risk.(string)
	if present && (!isText || !slices.Contains(riskClasses, text)) {
		name, _ := item["name"].(string)
		list.rule(at+"/risk_class", KeywordRiskClass, "risk_class", "command %q: risk_class %v is not low, medium or high, the classes a domain effector holds", name, risk)
	}
}
