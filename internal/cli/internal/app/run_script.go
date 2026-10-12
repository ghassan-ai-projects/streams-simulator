package app

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

// scenarioFromFlags is the scenario the run flags describe: the scripted
// faults, then perturbations, then effector invocations, then one clock
// advance to the end of --duration. Effector command ids are cli-<n> by
// position, so two identical invocations record identical command logs. It
// also returns, per command, the flag occurrence it came from, so a refusal
// names what the user wrote.
func scenarioFromFlags(options runOptions) (*model.Scenario, []string, error) {
	script, err := scriptedCommands(options)
	if err != nil {
		return nil, nil, err
	}
	end := options.startTime + int64(options.durationS*1e9)
	script = append(script, scripted{model.ScenarioCommand{Op: model.OpClockAdvance, Args: map[string]any{"to_ns": end}},
		fmt.Sprintf("--duration %v", options.durationS)})
	commands, origins := splitScript(script)
	return &model.Scenario{SchemaVersion: model.ScenarioSchemaVersion, ID: "cli-run", Domain: options.domainID,
		Adapter: options.adapterID, Seed: options.seed, Commands: commands,
		World: model.ScenarioWorld{StartTimeNS: options.startTime, Profile: options.profile}}, origins, nil
}

// splitScript separates the scripted commands from the flags behind them.
func splitScript(script []scripted) ([]model.ScenarioCommand, []string) {
	commands, origins := make([]model.ScenarioCommand, len(script)), make([]string, len(script))
	for i, step := range script {
		commands[i], origins[i] = step.command, step.origin
	}
	return commands, origins
}

// scripted is one scenario command and the flag occurrence it came from.
type scripted struct {
	command model.ScenarioCommand
	origin  string
}

// scriptedCommands are the faults, then the perturbations, then the effector
// invocations the flags script.
func scriptedCommands(options runOptions) ([]scripted, error) {
	var script []scripted
	for _, build := range []func(runOptions) ([]scripted, error){faultCommands, perturbationCommands, effectorCommands} {
		commands, err := build(options)
		if err != nil {
			return nil, err
		}
		script = append(script, commands...)
	}
	return script, nil
}

func faultCommands(options runOptions) ([]scripted, error) {
	var script []scripted
	for _, item := range options.faults.items() {
		command, err := faultCommand(item, options.startTime)
		if err != nil {
			return nil, err
		}
		script = append(script, command)
	}
	return script, nil
}

// faultCommand reads one entity=fault@offset_s entry.
func faultCommand(item string, start int64) (scripted, error) {
	entity, fault, offset, err := parseTriple(item, "=", "@")
	if err != nil {
		return scripted{}, err
	}
	onset, err := offsetNS("--fault", item, offset)
	return scripted{model.ScenarioCommand{Op: model.OpFaultInject, Args: map[string]any{
		"entity_id": entity, "fault": fault, "onset_ns": start + onset, "params": nil}}, fmt.Sprintf("--fault %q", item)}, err
}

// offsetNS reads an offset in seconds from the start as nanoseconds: a
// finite number within the run span. An absent offset is the start itself.
func offsetNS(flagName, item, text string) (int64, error) {
	if text == "" {
		return 0, nil
	}
	seconds, err := strconv.ParseFloat(text, 64)
	if err != nil || !(math.Abs(seconds)*1e9 <= float64(run.MaxRunSpanNS)) {
		return 0, usagef("%s %q: offset %q must be a number of seconds within ±%d", flagName, item, text, run.MaxRunSpanNS/1e9)
	}
	return int64(seconds * 1e9), nil
}

func perturbationCommands(options runOptions) ([]scripted, error) {
	var script []scripted
	for _, item := range options.perts.items() {
		name, from, until, err := scriptedPerturbation(item, options.startTime)
		if err != nil {
			return nil, err
		}
		script = append(script, scripted{model.ScenarioCommand{Op: model.OpPerturbApply, Args: map[string]any{
			"perturbation": name, "params": nil, "from_ns": from, "until_ns": until}}, fmt.Sprintf("--perturb %q", item)})
	}
	return script, nil
}

func effectorCommands(options runOptions) ([]scripted, error) {
	var script []scripted
	for n, item := range options.effectors.items() {
		spec, err := parseEffectorSpec(item)
		if err != nil {
			return nil, err
		}
		script = append(script, scripted{model.ScenarioCommand{Op: model.OpEffectorInvoke, Args: map[string]any{
			"effector": spec.effector, "entity_id": spec.entity, "command_id": fmt.Sprintf("cli-%d", n),
			"args": spec.args, "at_ns": options.startTime + spec.offsetNS}}, fmt.Sprintf("--effector %q", item)})
	}
	return script, nil
}

// scriptedPerturbation reads name@from_s[@until_s]; an end that is absent or
// not after the start offset 0 means no end.
func scriptedPerturbation(item string, start int64) (string, int64, int64, error) {
	name, window, _ := strings.Cut(item, "@")
	fromText, untilText, _ := strings.Cut(window, "@")
	from, err := offsetNS("--perturb", item, fromText)
	if err != nil {
		return "", 0, 0, err
	}
	until, err := offsetNS("--perturb", item, untilText)
	if err != nil || until <= 0 {
		return name, start + from, 0, err
	}
	return name, start + from, start + until, nil
}

// effectorSpec is one parsed --effector entry.
type effectorSpec struct {
	effector, entity string
	offsetNS         int64
	args             map[string]any
}

// parseEffectorSpec reads effector@entity@offset_s[@json-object]. The
// arguments object is optional: an effector with required arguments needs it.
func parseEffectorSpec(item string) (effectorSpec, error) {
	parts := strings.SplitN(item, "@", 4)
	if len(parts) < 2 {
		return effectorSpec{}, usagef("bad triple %q", item)
	}
	parts = append(parts, "", "")
	offset, err := offsetNS("--effector", item, parts[2])
	if err != nil {
		return effectorSpec{}, err
	}
	spec := effectorSpec{effector: parts[0], entity: parts[1], offsetNS: offset, args: map[string]any{}}
	return spec.withArguments(parts[3])
}

// withArguments reads the arguments object; an absent one is empty.
func (spec effectorSpec) withArguments(text string) (effectorSpec, error) {
	if text == "" {
		return spec, nil
	}
	if err := model.DecodeBytes([]byte(text), &spec.args); err != nil {
		return effectorSpec{}, usagef("effector %s: arguments must be a JSON object: %v", spec.effector, err)
	}
	return spec, nil
}
