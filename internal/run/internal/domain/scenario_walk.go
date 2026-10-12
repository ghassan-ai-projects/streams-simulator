package domain

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// scenarioWalk follows a scenario's commands in order, tracking the clock it
// advances, so each command is checked as it will be when it runs. index is
// the command being checked, so every issue points at its argument.
type scenarioWalk struct {
	facts ScenarioFacts
	start int64
	clock int64
	index int
}

func newScenarioWalk(facts ScenarioFacts, startNS int64) *scenarioWalk {
	return &scenarioWalk{facts: facts, start: startNS, clock: startNS}
}

// check returns the issues of one command; a command whose times are not
// exact is not checked further, since its effect cannot be known.
func (walk *scenarioWalk) check(index int, command *model.ScenarioCommand) []model.Issue {
	walk.index = index
	if issues := walk.times(command); len(issues) > 0 {
		return issues
	}
	return walk.operation(command)
}

// times checks that every time argument of the command is an exact integer.
func (walk *scenarioWalk) times(command *model.ScenarioCommand) []model.Issue {
	var issues []model.Issue
	for _, key := range timeArguments(command.Op) {
		if _, exact := ExactTime(command.Args, key); !exact {
			issues = append(issues, walk.issue(key, "a time must be an integer number of nanoseconds that fits in 64 bits, written without a fraction or exponent")...)
		}
	}
	return issues
}

// timeArguments names the time arguments of a scenario command.
func timeArguments(op string) []string {
	switch op {
	case model.OpClockAdvance:
		return []string{"to_ns"}
	case model.OpFaultInject:
		return []string{"onset_ns"}
	case model.OpPerturbApply:
		return []string{"from_ns", "until_ns"}
	case model.OpFaultClear, model.OpEffectorInvoke, model.OpEntityAdd, model.OpEntityRetire:
		return []string{"at_ns"}
	}
	return nil
}

// operation checks what the command itself names or changes.
func (walk *scenarioWalk) operation(command *model.ScenarioCommand) []model.Issue {
	switch command.Op {
	case model.OpClockAdvance:
		return walk.advance(command.Args)
	case model.OpFaultInject:
		return walk.fault(command.Args)
	case model.OpPerturbApply:
		return walk.perturbation(command.Args)
	case model.OpEffectorInvoke:
		return walk.invocation(command.Args)
	}
	return nil
}

// issue is one issue of the named argument of the command being checked.
func (walk *scenarioWalk) issue(field, format string, args ...any) []model.Issue {
	pointer := fmt.Sprintf("/commands/%d/args/%s", walk.index, field)
	return []model.Issue{model.LocatedIssue(pointer, fmt.Sprintf(format, args...))}
}

func (walk *scenarioWalk) advance(args map[string]any) []model.Issue {
	to := CommandTime(args, "to_ns")
	if to < walk.clock {
		return walk.issue("to_ns", "the clock would move backwards from %d to %d", walk.clock, to)
	}
	if ExceedsRunSpan(walk.start, to) {
		return walk.issue("to_ns", "the clock would move further than %d ns (30 days) past the start", MaxRunSpanNS)
	}
	walk.clock = to
	return nil
}

func (walk *scenarioWalk) fault(args map[string]any) []model.Issue {
	fault := CommandString(args, "fault")
	if !walk.facts.HasFault(fault) {
		return walk.issue("fault", "the domain declares no fault %q", fault)
	}
	if err := walk.facts.CheckFaultParams(fault, AsMap(args["params"])); err != nil {
		return walk.issue("params", "%v", err)
	}
	return nil
}

func (walk *scenarioWalk) perturbation(args map[string]any) []model.Issue {
	name := CommandString(args, "perturbation")
	if err := walk.facts.CheckPerturbationParams(name, AsMap(args["params"])); err != nil {
		field := "params"
		if !walk.facts.HasPerturbation(name) {
			field = "perturbation"
		}
		return walk.issue(field, "%v", err)
	}
	if err := walk.facts.CheckPerturbationWindow(name, CommandTime(args, "from_ns"), CommandTime(args, "until_ns")); err != nil {
		return walk.issue("until_ns", "%v", err)
	}
	return nil
}

func (walk *scenarioWalk) invocation(args map[string]any) []model.Issue {
	effector := CommandString(args, "effector")
	if !walk.facts.HasEffector(effector) {
		return walk.issue("effector", "the domain declares no effector %q", effector)
	}
	if err := walk.facts.CheckEffectorArgs(effector, AsMap(args["args"])); err != nil {
		return walk.issue("args", "effector %q refuses these args: %v", effector, err)
	}
	return nil
}
