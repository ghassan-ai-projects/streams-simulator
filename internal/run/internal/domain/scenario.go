package domain

import (
	"fmt"
	"regexp"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func IsScenarioOp(op string) bool {
	switch op {
	case model.OpClockAdvance, model.OpFaultInject, model.OpFaultClear, model.OpPerturbApply,
		model.OpPerturbClear, model.OpEffectorInvoke, model.OpEntityAdd, model.OpEntityRetire:
		return true
	}
	return false
}

var scenarioIDPattern = regexp.MustCompile(`^[a-z][a-z0-9-]{0,63}$`)

func ScenarioFromArtifact(art *model.RunArtifact) (*model.Scenario, int) {
	scenario := &model.Scenario{SchemaVersion: model.ScenarioSchemaVersion, ID: exportedScenarioID(art.RunID),
		Domain: art.Domain.ID, Adapter: art.Adapter.ID, Seed: art.Seed, World: scenarioWorld(art.WorldConfig),
		Commands: []model.ScenarioCommand{}}
	omitted := 0
	for _, command := range art.CommandLog {
		if !IsScenarioOp(command.Op) {
			omitted++
			continue
		}
		scenario.Commands = append(scenario.Commands, model.ScenarioCommand{Op: command.Op, Args: scenarioArgs(command.Op, command.Args)})
	}
	return scenario, omitted
}

func scenarioArgs(op string, args map[string]any) map[string]any {
	copied := cloneObject(args)
	if op == model.OpClockAdvance {
		delete(copied, "await_consumer")
	}
	return copied
}

func JournalCommand(command model.Command) model.Command {
	command.Args = scenarioArgs(command.Op, command.Args)
	return command
}

func cloneValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return cloneObject(typed)
	case []any:
		return cloneArray(typed)
	}
	return value
}

func cloneObject(object map[string]any) map[string]any {
	if object == nil {
		return nil
	}
	copied := make(map[string]any, len(object))
	// determinism-safe: the loop only copies each key into a new map.
	for key, item := range object {
		copied[key] = cloneValue(item)
	}
	return copied
}

func cloneArray(array []any) []any {
	if array == nil {
		return nil
	}
	copied := make([]any, len(array))
	for i, item := range array {
		copied[i] = cloneValue(item)
	}
	return copied
}

func exportedScenarioID(runID string) string {
	if scenarioIDPattern.MatchString(runID) {
		return runID
	}
	return "exported-run"
}

func scenarioWorld(config model.WorldConfig) model.ScenarioWorld {
	return model.ScenarioWorld{StartTimeNS: config.StartTimeNS, Profile: config.ScenarioProfile,
		EntityIDs: slices.Clone(config.EntityIDs), Noiseless: config.Noiseless, ForceFailureMode: config.ForceFailureMode,
		ClockMultiplier: config.ClockMultiplier}
}

// The Check methods return the violation text, or "" when the arguments are acceptable.
type ScenarioFacts interface {
	HasProfile(name string) bool
	HasFault(id string) bool
	CheckFaultParams(fault string, params map[string]any) string
	HasPerturbation(name string) bool
	CheckPerturbationParams(name string, params map[string]any) string
	CheckPerturbationWindow(name string, fromNS, untilNS int64) string
	HasEffector(name string) bool
	CheckEffectorArgs(effector string, args map[string]any) string
}

func CheckScenario(scenario *model.Scenario, facts ScenarioFacts) []model.Issue {
	var issues []model.Issue
	if scenario.World.Profile != "" && !facts.HasProfile(scenario.World.Profile) {
		issues = append(issues, model.LocatedIssue("/world/profile", fmt.Sprintf("the domain has no profile %q", scenario.World.Profile)))
	}
	walk := newScenarioWalk(facts, scenario.World.StartTimeNS)
	for index := range scenario.Commands {
		issues = append(issues, walk.check(index, &scenario.Commands[index])...)
	}
	return issues
}
