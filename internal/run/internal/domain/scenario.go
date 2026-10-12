package domain

import (
	"fmt"
	"regexp"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// IsScenarioOp reports whether a command-log operation is a scenario command
// (docs/design/SCENARIO_FILE.md §2). The run's lifecycle records and the
// record-only env.inject are not.
func IsScenarioOp(op string) bool {
	switch op {
	case model.OpClockAdvance, model.OpFaultInject, model.OpFaultClear, model.OpPerturbApply,
		model.OpPerturbClear, model.OpEffectorInvoke, model.OpEntityAdd, model.OpEntityRetire:
		return true
	}
	return false
}

var scenarioIDPattern = regexp.MustCompile(`^[a-z][a-z0-9-]{0,63}$`)

// ScenarioFromArtifact is the scenario that reproduces a recorded run: its
// inputs and every command-log entry that is a scenario command, in order,
// with copies of their arguments. It also returns how many entries were
// omitted.
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

// scenarioArgs copies a logged command's arguments into a scenario command.
// A clock advance keeps only its target: whether a live consumer was awaited
// is pacing of that session, not an input to the world (SCENARIO_FILE §2).
func scenarioArgs(op string, args map[string]any) map[string]any {
	copied := cloneObject(args)
	if op == model.OpClockAdvance {
		delete(copied, "await_consumer")
	}
	return copied
}

// cloneValue deep-copies a decoded JSON value; a nil object or array stays
// nil, so it still encodes as null.
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

// exportedScenarioID names an exported scenario after its run when the run id
// is a valid scenario id.
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

// ScenarioFacts answers what the static scenario check needs to know about
// the domain and the perturbation catalog. A Check method's error states the
// violation.
type ScenarioFacts interface {
	HasProfile(name string) bool
	HasFault(id string) bool
	CheckFaultParams(fault string, params map[string]any) error
	HasPerturbation(name string) bool
	CheckPerturbationParams(name string, params map[string]any) error
	CheckPerturbationWindow(name string, fromNS, untilNS int64) error
	HasEffector(name string) bool
	CheckEffectorArgs(effector string, args map[string]any) error
}

// CheckScenario returns every way a scenario's world settings and commands
// break the domain, the perturbation catalog or the run's rules, as far as
// that is known before the run unfolds (SCENARIO_FILE §3), each located in
// the scenario document. Facts that depend on the run as it goes (which
// entities exist, the ids it assigns) are not checked here. An empty result
// means the scenario may be tried.
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
