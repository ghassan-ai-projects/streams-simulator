package app

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/perturb"
	rules "github.com/ghassan-ai-projects/streams-simulator/internal/run/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

// ScenarioConfig is the run configuration a scenario describes
// (docs/design/SCENARIO_FILE.md); the sink is the caller's choice.
func ScenarioConfig(scenario *model.Scenario, spec *domain.Compiled, adapterSpec *model.Adapter, sinkName, sinkTarget string) Config {
	settings := scenario.World
	return Config{Domain: spec, Adapter: adapterSpec, Seed: scenario.Seed, SinkName: sinkName, SinkTarget: sinkTarget,
		TimeMode: model.TimeStepped, StartTimeNS: settings.StartTimeNS, StartTimeSet: true, EntityIDs: settings.EntityIDs,
		ScenarioProfile: settings.Profile, ClockMultiplier: settings.ClockMultiplier, Noiseless: settings.Noiseless,
		ForceFailureMode: world.FailureMode(settings.ForceFailureMode)}
}

// CheckScenario checks a scenario against its domain and the world it creates,
// before anything is delivered and without opening a sink (SCENARIO_FILE §3).
// The static check reports every issue it finds, the profile included; a
// scenario without one is then dry-run, through the executor the real run
// uses, on the same world emitting nothing, so a command the run would refuse
// as it unfolds is found at that command. A violation is a
// *model.ValidationError.
func CheckScenario(ctx context.Context, cfg Config, scenario *model.Scenario) error {
	shadow, err := newShadow(ctx, cfg)
	if err != nil {
		return scenarioViolation(cfg, scenario, []model.Issue{model.LocatedIssue("/world", err.Error())})
	}
	if issues := rules.CheckScenario(scenario, scenarioFacts{created: shadow.World}); len(issues) > 0 {
		return scenarioViolation(cfg, scenario, issues)
	}
	return dryRunScenario(ctx, shadow, cfg, scenario)
}

// newShadow is a run of the scenario's world that emits nothing and opens
// nothing: the world a scenario is dry-run on. The profile is the static
// check's to judge, so it is not refused here.
func newShadow(ctx context.Context, cfg Config) (*Run, error) {
	cfg = defaultConfig(cfg)
	cfg.silent = true
	created, err := newRunWorld(cfg)
	if err != nil {
		return nil, err
	}
	return newRunState(ctx, cfg, created)
}

// dryRunScenario executes the commands on the shadow run, stopping at the
// first one the run refuses.
func dryRunScenario(ctx context.Context, shadow *Run, cfg Config, scenario *model.Scenario) error {
	for index, step := range scenario.Commands {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("run: scenario check: %w", err)
		}
		command := model.Command{Op: step.Op, Args: step.Args}
		if err := executeCommand(ctx, shadow, &command, keepWorldRefusals); err != nil {
			issue := model.LocatedIssue(fmt.Sprintf("/commands/%d", index), "the run refuses this "+step.Op+": "+rootCause(err).Error())
			return scenarioViolation(cfg, scenario, []model.Issue{issue})
		}
	}
	return nil
}

// rootCause is the error at the end of a chain of single wraps: the reason
// itself, without the layers that passed it on.
func rootCause(err error) error {
	for next := errors.Unwrap(err); next != nil; next = errors.Unwrap(err) {
		err = next
	}
	return err
}

func scenarioViolation(cfg Config, scenario *model.Scenario, issues []model.Issue) error {
	summary := fmt.Sprintf("scenario %q does not fit domain %q: %s", scenario.ID, cfg.Domain.Spec.ID, issues[0].Message)
	return &model.ValidationError{Contract: model.ContractScenario, Summary: summary, Issues: issues}
}

// ExecuteScenario runs a checked scenario's commands in order through the
// executor that also replays a command log. An effector refusal the world
// decides as it runs (an interlock, a reused command id) is a recorded
// outcome; any other refusal stops the scenario.
func (r *Run) ExecuteScenario(ctx context.Context, scenario *model.Scenario) error {
	for index, step := range scenario.Commands {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("run: scenario command %d (%s): %w", index, step.Op, err)
		}
		command := model.Command{Op: step.Op, Args: step.Args}
		if err := executeCommand(ctx, r, &command, keepWorldRefusals); err != nil {
			return fmt.Errorf("run: scenario command %d (%s): %w", index, step.Op, err)
		}
	}
	return nil
}

// ScenarioFromArtifact is the scenario that reproduces a recorded run, and
// the number of command-log entries it omits.
func ScenarioFromArtifact(art *model.RunArtifact) (*model.Scenario, int) {
	return rules.ScenarioFromArtifact(art)
}

// scenarioFacts answers the static scenario check from the created world's
// domain and the perturbation catalog.
type scenarioFacts struct {
	created *world.World
}

func (facts scenarioFacts) HasProfile(name string) bool { return facts.created.Spec.HasProfile(name) }

func (facts scenarioFacts) HasFault(id string) bool { return facts.created.Spec.HasFault(id) }

func (facts scenarioFacts) CheckFaultParams(fault string, params map[string]any) error {
	return facts.created.CheckFaultParams(fault, params) //nolint:wrapcheck // the violation is the issue message
}

func (facts scenarioFacts) HasPerturbation(name string) bool {
	return slices.Contains(perturb.Names(), name)
}

func (facts scenarioFacts) CheckPerturbationParams(name string, params map[string]any) error {
	return perturb.CheckParams(name, params) //nolint:wrapcheck // the violation is the issue message
}

func (facts scenarioFacts) CheckPerturbationWindow(name string, fromNS, untilNS int64) error {
	return perturb.CheckWindow(name, fromNS, untilNS) //nolint:wrapcheck // the violation is the issue message
}

func (facts scenarioFacts) HasEffector(name string) bool { return facts.created.Spec.HasEffector(name) }

func (facts scenarioFacts) CheckEffectorArgs(effector string, args map[string]any) error {
	return facts.created.CheckEffectorArgs(effector, args) //nolint:wrapcheck // the violation is the issue message
}
