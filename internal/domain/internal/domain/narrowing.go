package domain

import (
	"fmt"
	"maps"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func (checks domainChecks) checkNarrowing(base *model.Kind, n *model.Narrowing) error {
	if n == nil {
		return nil
	}
	scope := checks.within(checks.at + "/narrow")
	for _, check := range []func(*model.Kind, *model.Narrowing) error{scope.narrowStates, scope.narrowChannels, scope.narrowEffectors} {
		if err := check(base, n); err != nil {
			return err
		}
	}
	return nil
}

func (checks domainChecks) narrowStates(base *model.Kind, n *model.Narrowing) error {
	return narrowEach(checks, "state", n.State, func(item *model.StateNarrowing) string { return item.Name }, base.StateDecl, stateNarrowingViolation)
}

func narrowEach[D, T any](checks domainChecks, what string, items []T, name func(*T) string, declared func(string) *D, violation func(*D, *T) (string, string)) error {
	seen := map[string]bool{}
	return checks.each(what, len(items), func(at domainChecks, i int) error {
		decl := declared(name(&items[i]))
		if seen[name(&items[i])] {
			return at.bad("/name", "narrowing names %s %q twice", what, name(&items[i]))
		}
		seen[name(&items[i])] = true
		if decl == nil {
			return at.bad("/name", "narrowing names %s %q, which the kind does not declare", what, name(&items[i]))
		}
		return at.rejectViolation(violation(decl, &items[i]))
	})
}

func (checks domainChecks) rejectViolation(member, message string) error {
	if message == "" {
		return nil
	}
	return checks.bad(member, "%s", message)
}

func stateNarrowingViolation(decl *model.State, n *model.StateNarrowing) (string, string) {
	switch {
	case n.Min >= n.Max:
		return "/max", fmt.Sprintf("state %q narrowing needs min below max", n.Name)
	case decl.Min != nil && n.Min < *decl.Min:
		return "/min", fmt.Sprintf("state %q min %v widens the kind's min %v", n.Name, n.Min, *decl.Min)
	case decl.Max != nil && n.Max > *decl.Max:
		return "/max", fmt.Sprintf("state %q max %v widens the kind's max %v", n.Name, n.Max, *decl.Max)
	case decl.Initial < n.Min || decl.Initial > n.Max:
		return "/min", fmt.Sprintf("state %q initial %v lies outside the narrowed bounds", n.Name, decl.Initial)
	}
	return "", ""
}

func (checks domainChecks) narrowChannels(base *model.Kind, n *model.Narrowing) error {
	return narrowEach(checks, "channels", n.Channels, func(item *model.ChannelNarrowing) string { return item.Name }, base.ChannelDecl, rangeViolation)
}

func rangeViolation(decl *model.Channel, n *model.ChannelNarrowing) (string, string) {
	switch {
	case n.Range.Min >= n.Range.Max:
		return "/range/max", fmt.Sprintf("channel %q range needs min below max", n.Name)
	case decl.Range != nil && n.Range.Min < decl.Range.Min:
		return "/range/min", fmt.Sprintf("channel %q range min %v widens the kind's %v", n.Name, n.Range.Min, decl.Range.Min)
	case decl.Range != nil && n.Range.Max > decl.Range.Max:
		return "/range/max", fmt.Sprintf("channel %q range max %v widens the kind's %v", n.Name, n.Range.Max, decl.Range.Max)
	}
	return "", ""
}

func (checks domainChecks) narrowEffectors(base *model.Kind, n *model.Narrowing) error {
	seen := map[string]bool{}
	return checks.each("effectors", len(n.Effectors), func(at domainChecks, i int) error {
		item := &n.Effectors[i]
		decl := base.EffectorDecl(item.Name)
		if seen[item.Name] {
			return at.bad("/name", "narrowing names effector %q twice", item.Name)
		}
		seen[item.Name] = true
		if decl == nil {
			return at.bad("/name", "narrowing names effector %q, which the kind does not declare", item.Name)
		}
		return at.narrowArguments(decl, item)
	})
}

func (checks domainChecks) narrowArguments(decl *model.Effector, n *model.EffectorNarrowing) error {
	for _, arg := range slices.Sorted(maps.Keys(n.Args)) {
		member, message := argumentViolation(decl, arg, n.Args[arg])
		if err := checks.rejectViolation("/args"+memberPointer(arg)+member, message); err != nil {
			return err
		}
	}
	return nil
}

func argumentViolation(decl *model.Effector, arg string, b model.ArgBounds) (string, string) {
	property := argumentProperty(decl, arg)
	if property == nil {
		return "", fmt.Sprintf("effector %q narrows argument %q, which its args_schema does not declare as a number", decl.Name, arg)
	}
	return boundsViolation(property, arg, b)
}

func boundsViolation(property map[string]any, arg string, b model.ArgBounds) (string, string) {
	lo, hasLo := bound(property, "minimum")
	hi, hasHi := bound(property, "maximum")
	switch {
	case b.Minimum != nil && hasLo && *b.Minimum < lo:
		return "/minimum", fmt.Sprintf("argument %q minimum %v widens the declared %v", arg, *b.Minimum, lo)
	case b.Maximum != nil && hasHi && *b.Maximum > hi:
		return "/maximum", fmt.Sprintf("argument %q maximum %v widens the declared %v", arg, *b.Maximum, hi)
	}
	return emptyBounds(arg, narrowedBound(b.Minimum, lo, hasLo), narrowedBound(b.Maximum, hi, hasHi))
}

func narrowedBound(narrowed *float64, declared float64, hasDeclared bool) *float64 {
	if narrowed == nil && hasDeclared {
		return &declared
	}
	return narrowed
}

func emptyBounds(arg string, lo, hi *float64) (string, string) {
	if lo != nil && hi != nil && *lo > *hi {
		return "/maximum", fmt.Sprintf("argument %q bounds leave no value: minimum %v above maximum %v", arg, *lo, *hi)
	}
	return "", ""
}

func argumentProperty(decl *model.Effector, arg string) map[string]any {
	properties, _ := decl.ArgsSchema["properties"].(map[string]any)
	property, _ := properties[arg].(map[string]any)
	if kind, _ := property["type"].(string); kind != "number" && kind != "integer" {
		return nil
	}
	return property
}

func bound(property map[string]any, key string) (float64, bool) {
	value, ok := property[key].(float64)
	return value, ok
}

func applyNarrowing(kind *model.Kind, n *model.Narrowing) {
	for _, item := range n.State {
		state := &kind.State[stateIndex(kind, item.Name)]
		lower, upper := item.Min, item.Max
		state.Min, state.Max = &lower, &upper
	}
	narrowChannelRanges(kind, n.Channels)
	narrowEffectorSchemas(kind, n.Effectors)
}

func narrowChannelRanges(kind *model.Kind, items []model.ChannelNarrowing) {
	for _, item := range items {
		for i := range kind.Channels {
			if kind.Channels[i].Name == item.Name {
				narrowed := item.Range
				kind.Channels[i].Range = &narrowed
			}
		}
	}
}

func stateIndex(kind *model.Kind, name string) int {
	return slices.IndexFunc(kind.State, func(s model.State) bool { return s.Name == name })
}

func narrowEffectorSchemas(kind *model.Kind, items []model.EffectorNarrowing) {
	for _, item := range items {
		for i := range kind.Effectors {
			if kind.Effectors[i].Name == item.Name {
				kind.Effectors[i].ArgsSchema = narrowedSchema(kind.Effectors[i].ArgsSchema, item.Args)
			}
		}
	}
}

func narrowedSchema(schema map[string]any, args map[string]model.ArgBounds) map[string]any {
	out := maps.Clone(schema)
	properties := maps.Clone(schema["properties"].(map[string]any))
	for _, arg := range slices.Sorted(maps.Keys(args)) {
		properties[arg] = narrowedProperty(properties[arg].(map[string]any), args[arg])
	}
	out["properties"] = properties
	return out
}

func narrowedProperty(property map[string]any, b model.ArgBounds) map[string]any {
	out := maps.Clone(property)
	if b.Minimum != nil {
		out["minimum"] = *b.Minimum
	}
	if b.Maximum != nil {
		out["maximum"] = *b.Maximum
	}
	return out
}
