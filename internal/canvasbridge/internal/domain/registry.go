package domain

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"
)

type ActionKind string

const (
	KindMCP ActionKind = "mcp"
	KindCLI ActionKind = "cli"
)

type Action struct {
	Name     string     `json:"name"`
	Kind     ActionKind `json:"kind"`
	Tool     string     `json:"tool,omitempty"`
	Command  []string   `json:"command,omitempty"`
	Keys     []string   `json:"-"`
	CLIArgs  []CLIArg   `json:"-"`
	Withheld []string   `json:"-"`
	Catalog  bool       `json:"-"`
	Library  bool       `json:"-"`
	Kinds    bool       `json:"-"`
	Etag     bool       `json:"-"`
	Mutating bool       `json:"mutating"`
}

func mcp(name, tool string, mutating bool, keys ...string) Action {
	return Action{Name: name, Kind: KindMCP, Tool: tool, Keys: keys, Mutating: mutating}
}

var registry = []Action{
	mcp("catalog.list", "sim.catalog.list", false, "group"),
	mcp("catalog.describe", "sim.catalog.describe", false, "domain"),
	mcp("adapter.list", "sim.adapter.list", false),
	withholding(mcp("world.create", "sim.world.create", true, "domain", "seed", "entities", "scenario_profile", "adapter", "time_mode",
		"start_time", "label", "explore"), "token", "operator_endpoint"),
	mcp("world.describe", "sim.world.describe", false, "world_id"),
	mcp("world.destroy", "sim.world.destroy", true, "world_id"),
	mcp("clock.advance", "sim.clock.advance", true, "world_id", "by_ns", "to_ns"),
	mcp("clock.state", "sim.clock.state", false, "world_id"),
	mcp("fault.inject", "sim.fault.inject", true, "world_id", "entity_id", "fault", "onset_ns", "params"),
	mcp("fault.clear", "sim.fault.clear", true, "world_id", "fault_id"),
	mcp("effector.list", "sim.world.effectors", false, "world_id"),
	mcp("effector.invoke", "sim.world.invoke", true, "world_id", "entity_id", "effector", "args", "command_id"),
	mcp("perturb.apply", "sim.perturb.apply", true, "world_id", "perturbation", "params", "from_ns", "until_ns"),
	mcp("perturb.clear", "sim.perturb.clear", true, "world_id", "perturb_id"),
	mcp("env.inject", "sim.env.inject", true, "world_id", "target", "fault", "params"),
	mcp("run.begin", "sim.run.begin", true, "world_id", "label"),
	mcp("run.end", "sim.run.end", true, "world_id"),
	mcp("truth.reveal", "sim.truth.reveal", false, "run_id"),
	mcp("truth.status", "sim.truth.seal_status", false, "run_id"),
	cliComponentList, cliComponentShow, cliComponentValidate, cliExamples, cliExample, cliNewPlant, cliCompose, cliValidate, cliLint,
	cliKindValidate, cliKindTwin, cliVerify, cliReport, cliScenario,
}

func withholding(action Action, keys ...string) Action {
	action.Withheld = keys
	return action
}

func (a Action) Withhold(result json.RawMessage) (json.RawMessage, error) {
	if len(a.Withheld) == 0 {
		return result, nil
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(result, &fields); err != nil {
		return nil, fmt.Errorf("canvasbridge: %s result: %w", a.Name, err)
	}
	for _, key := range a.Withheld {
		delete(fields, key)
	}
	return a.encode(fields)
}

func (a Action) encode(fields map[string]json.RawMessage) (json.RawMessage, error) {
	out, err := json.Marshal(fields)
	if err != nil {
		return nil, fmt.Errorf("canvasbridge: %s result: %w", a.Name, err)
	}
	return out, nil
}

func Actions() []Action {
	return slices.Clone(registry)
}

func Lookup(name string) (Action, error) {
	index := slices.IndexFunc(registry, func(a Action) bool { return a.Name == name })
	if index < 0 {
		return Action{}, fmt.Errorf("%w: %q", ErrUnknownAction, name)
	}
	return registry[index], nil
}

func (a Action) ToolArguments(args map[string]any) (map[string]any, error) {
	out := make(map[string]any, len(args))
	for _, key := range slices.Sorted(maps.Keys(args)) {
		if !slices.Contains(a.Keys, key) {
			return nil, fmt.Errorf("%w: %s takes no argument %q", ErrBadArguments, a.Name, key)
		}
		out[key] = args[key]
	}
	return out, nil
}
