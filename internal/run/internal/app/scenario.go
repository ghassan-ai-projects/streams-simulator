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

func ScenarioConfig(scenario *model.Scenario, spec *domain.Compiled, adapterSpec *model.Adapter, sinkName, sinkTarget string) Config {
	settings := scenario.World
	return Config{Domain: spec, Adapter: adapterSpec, Seed: scenario.Seed, SinkName: sinkName, SinkTarget: sinkTarget,
		TimeMode: model.TimeStepped, StartTimeNS: settings.StartTimeNS, StartTimeSet: true, EntityIDs: settings.EntityIDs,
		ScenarioProfile: settings.Profile, ClockMultiplier: settings.ClockMultiplier, Noiseless: settings.Noiseless,
		ForceFailureMode: world.FailureMode(settings.ForceFailureMode)}
}

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

func newShadow(ctx context.Context, cfg Config) (*Run, error) {
	cfg = defaultConfig(cfg)
	cfg.silent = true
	created, err := newRunWorld(cfg)
	if err != nil {
		return nil, err
	}
	return newRunState(ctx, cfg, created)
}

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

func ScenarioFromArtifact(art *model.RunArtifact) (*model.Scenario, int) {
	return rules.ScenarioFromArtifact(art)
}

type scenarioFacts struct {
	created *world.World
}

func (facts scenarioFacts) HasProfile(name string) bool { return facts.created.Spec.HasProfile(name) }

func (facts scenarioFacts) HasFault(id string) bool { return facts.created.Spec.HasFault(id) }

func (facts scenarioFacts) CheckFaultParams(fault string, params map[string]any) string {
	return violationText(facts.created.CheckFaultParams(fault, params))
}

func (facts scenarioFacts) HasPerturbation(name string) bool {
	return slices.Contains(perturb.Names(), name)
}

func (facts scenarioFacts) CheckPerturbationParams(name string, params map[string]any) string {
	return violationText(perturb.CheckParams(name, params))
}

func (facts scenarioFacts) CheckPerturbationWindow(name string, fromNS, untilNS int64) string {
	return violationText(perturb.CheckWindow(name, fromNS, untilNS))
}

func (facts scenarioFacts) HasEffector(name string) bool { return facts.created.Spec.HasEffector(name) }

func (facts scenarioFacts) CheckEffectorArgs(effector string, args map[string]any) string {
	return violationText(facts.created.CheckEffectorArgs(effector, args))
}

func violationText(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
