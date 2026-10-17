package domain

import (
	"maps"
	"slices"
)

func (t *twinBuild) planCommands(k *twinKind) error {
	if err := t.requireDecided(k, "commands", k.file.askedCommands(), slices.Sorted(maps.Keys(k.choice.Commands))); err != nil {
		return err
	}
	for _, command := range k.file.Commands {
		plan := &commandPlan{file: command, choice: k.choice.Commands[command.Name], at: pointer("kinds", indexPart(k.index), "commands", command.Name)}
		k.commands = append(k.commands, plan)
		if err := t.planResponse(k, plan); err != nil {
			return err
		}
	}
	return nil
}

func (t *twinBuild) planResponse(k *twinKind, c *commandPlan) error {
	if c.file.SafeStop {
		return nil
	}
	if err := t.checkShape(c); err != nil {
		return err
	}
	for _, arg := range slices.Sorted(maps.Keys(c.choice.Sets)) {
		if err := t.checkSet(k, c, arg, c.choice.Sets[arg]); err != nil {
			return err
		}
		k.actuated[c.choice.Sets[arg]] = true
	}
	return nil
}

func (t *twinBuild) checkShape(c *commandPlan) error {
	choice := c.choice
	sets, drives := len(choice.Sets) > 0, len(choice.Drives) > 0
	switch {
	case choice.Response == responseNone && (sets || drives):
		return t.refuse(ErrResponseMismatch, c.at, "response none takes no sets and no drives")
	case choice.Response == responseDirect && (!sets || drives):
		return t.refuse(ErrResponseMismatch, c.at, "response direct takes sets and no drives")
	case choice.Response == responseFirstOrder && (!sets || !drives):
		return t.refuse(ErrResponseMismatch, c.at, "response first_order takes sets and drives")
	}
	return nil
}

func (t *twinBuild) checkSet(k *twinKind, c *commandPlan, arg, metric string) error {
	at := c.at + pointer("sets", arg)
	file, found := k.file.metric(metric)
	if !numericType(argumentType(c.file.ArgsSchema, arg)) {
		return t.refuse(ErrUnknownChoice, at, "command %q has no numeric argument %q", c.file.Name, arg)
	}
	if !found {
		return t.refuse(ErrUnknownChoice, at, "kind %q has no metric %q", k.file.Name, metric)
	}
	return t.suitableSet(file, at)
}

func (t *twinBuild) suitableSet(metric kindMetric, at string) error {
	switch {
	case !metric.Writable:
		return t.refuse(ErrResponseMismatch, at, "metric %q is not writable in its kind file", metric.Name)
	case metric.EURange == nil || metric.DataType == "boolean":
		return t.refuse(ErrResponseMismatch, at, "metric %q: a command sets a number or integer metric that states its eu_range", metric.Name)
	}
	return nil
}

func argumentType(schema map[string]any, arg string) string {
	properties, _ := schema["properties"].(map[string]any)
	property, _ := properties[arg].(map[string]any)
	kind, _ := property["type"].(string)
	return kind
}

func numericType(kind string) bool { return kind == "number" || kind == "integer" }

func (t *twinBuild) checkActuators(k *twinKind) error {
	for _, metric := range k.file.Metrics {
		if metric.Writable && !k.actuated[metric.Name] {
			return t.refuse(ErrResponseMismatch, pointer("kinds", indexPart(k.index), "metrics", metric.Name), "metric %q is writable in its kind file, but no command sets it", metric.Name)
		}
	}
	for _, c := range k.commands {
		if c.file.SafeStop && len(k.actuated) == 0 {
			return t.refuse(ErrResponseMismatch, c.at, "safe stop %q has nothing to stop: no command of kind %q sets a metric", c.file.Name, k.file.Name)
		}
	}
	return nil
}
