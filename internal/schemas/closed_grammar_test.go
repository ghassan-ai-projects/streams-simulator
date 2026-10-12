package schemas

import (
	"encoding/json"
	"regexp"
	"slices"
	"testing"
)

var proseMembers = []string{"description", "title", "stresses", "not_applicable", "unit", "if_no_action", "if_action_by_deadline"}

var identifierMembers = []string{"name", "id_template", "command_id", "instance", "fmu", "enum_values"}

var referenceMembers = []string{"assign_from_arg", "channel", "channel_a", "channel_b", "confirmation_channels", "effector", "expected_effector",
	"from", "from_arg", "inputs", "observes", "outputs", "reset_on", "scale", "selector", "state", "target", "to", "trigger_state", "watch"}

var openObjectMembers = []string{"args_schema", "args", "params", "inputs", "outputs"}

var expressionLike = regexp.MustCompile(`(?i)(^|_)(expr|expression|formula|equation|condition|script|eval|lambda|code|rule)(_|$)`)

var trivialPattern = regexp.MustCompile(`^\^?(\.\*|\.\+)?\$?$`)

func TestTheDomainSchemaHoldsNoFreeTextExpression(t *testing.T) {
	t.Parallel()
	grammar := walkGrammar(t, DomainSpec())
	for _, problem := range grammar.problems(slices.Concat(proseMembers, identifierMembers, referenceMembers), openObjectMembers) {
		t.Error(problem)
	}
	for _, member := range slices.Concat(proseMembers, identifierMembers, referenceMembers) {
		if !slices.Contains(grammar.freeStrings, member) {
			t.Errorf("allowlisted member %q is no longer a free string; remove it from the list", member)
		}
	}
}

func TestTheGrammarWalkFindsAnExpressionSlippedIn(t *testing.T) {
	t.Parallel()
	for name, property := range map[string]any{
		"rule":    map[string]any{"type": "string"},
		"guard":   map[string]any{"type": []any{"string", "null"}},
		"anytext": map[string]any{"type": "string", "pattern": ".*"},
		"untyped": map[string]any{},
		"bag":     map[string]any{"type": "object"},
	} {
		schema := map[string]any{"type": "object", "properties": map[string]any{name: property}}
		raw, err := json.Marshal(schema)
		if err != nil {
			t.Fatal(err)
		}
		if len(walkGrammar(t, raw).problems(nil, nil)) == 0 {
			t.Errorf("member %q was not reported", name)
		}
	}
}

func walkGrammar(t *testing.T, raw []byte) *grammarWalk {
	t.Helper()
	var schema any
	if err := json.Unmarshal(raw, &schema); err != nil {
		t.Fatal(err)
	}
	grammar := &grammarWalk{}
	grammar.walk(schema, "")
	return grammar
}

type grammarWalk struct {
	freeStrings, openObjects, untyped, members []string
}

func (g *grammarWalk) problems(known, open []string) []string {
	var out []string
	for _, member := range g.freeStrings {
		if !slices.Contains(known, member) {
			out = append(out, "member "+member+" is an unconstrained string: give it an enum or a pattern, or classify it as prose, identifier or reference")
		}
	}
	for _, member := range g.openObjects {
		if !slices.Contains(open, member) {
			out = append(out, "member "+member+" is an object of unchecked members")
		}
	}
	for _, member := range slices.Concat(g.untyped, g.members) {
		if slices.Contains(g.untyped, member) || expressionLike.MatchString(member) {
			out = append(out, "member "+member+" has no type or names an expression")
		}
	}
	return out
}

func (g *grammarWalk) walk(node any, member string) {
	switch typed := node.(type) {
	case map[string]any:
		g.classify(typed, member)
		for key, child := range typed {
			g.descend(key, child, member)
		}
	case []any:
		for _, child := range typed {
			g.walk(child, member)
		}
	}
}

func (g *grammarWalk) descend(key string, child any, member string) {
	if key != "properties" {
		g.walk(child, member)
		return
	}
	for name, property := range child.(map[string]any) {
		g.members = append(g.members, name)
		g.walk(property, name)
	}
}

func (g *grammarWalk) classify(node map[string]any, member string) {
	pattern, _ := node["pattern"].(string)
	constrained := node["enum"] != nil || node["const"] != nil || node["format"] != nil || (pattern != "" && !trivialPattern.MatchString(pattern))
	if allowsString(node["type"]) && !constrained {
		g.freeStrings = append(g.freeStrings, member)
	}
	if node["type"] == "object" && node["properties"] == nil && node["additionalProperties"] != false {
		if _, typedValues := node["additionalProperties"].(map[string]any); !typedValues {
			g.openObjects = append(g.openObjects, member)
		}
	}
	if member != "" && len(node) == 0 {
		g.untyped = append(g.untyped, member)
	}
}

func allowsString(kind any) bool {
	if list, ok := kind.([]any); ok {
		return slices.Contains(list, any("string"))
	}
	return kind == "string"
}
