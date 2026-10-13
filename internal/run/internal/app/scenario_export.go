package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

func PruneRefusedCalls(ctx context.Context, cfg Config, scenario *model.Scenario) (*model.Scenario, int, error) {
	shadow, err := newShadow(ctx, cfg)
	if err != nil {
		return nil, 0, err
	}
	pruned := *scenario
	pruned.Commands = []model.ScenarioCommand{}
	for _, step := range scenario.Commands {
		if !refusedAtAdmission(ctx, shadow, step) {
			pruned.Commands = append(pruned.Commands, step)
		}
	}
	return &pruned, len(scenario.Commands) - len(pruned.Commands), nil
}

func refusedAtAdmission(ctx context.Context, shadow *Run, step model.ScenarioCommand) bool {
	command := model.Command{Op: step.Op, Args: step.Args}
	return errors.Is(executeCommand(ctx, shadow, &command, keepWorldRefusals), world.ErrInvalidInvocation)
}

func CheckArtifactInputs(art *model.RunArtifact, spec *domain.Compiled, adapterSpec *model.Adapter) error {
	return validateReplayInputs(art, spec, adapterSpec)
}

type ScenarioExport struct {
	Scenario *model.Scenario
	Omitted  int
	Refused  int
	Raw      []byte
}

func ExportScenario(ctx context.Context, art *model.RunArtifact, spec *domain.Compiled, adapterSpec *model.Adapter) (ScenarioExport, error) {
	scenario, omitted := ScenarioFromArtifact(art)
	if err := CheckArtifactInputs(art, spec, adapterSpec); err != nil {
		return ScenarioExport{}, err
	}
	cfg := ScenarioConfig(scenario, spec, adapterSpec, model.SinkInproc, "")
	pruned, refused, raw, err := checkedExport(ctx, cfg, scenario)
	if err != nil {
		return ScenarioExport{}, fmt.Errorf("the run cannot be expressed as a runnable scenario: %w", err)
	}
	return ScenarioExport{Scenario: pruned, Omitted: omitted, Refused: refused, Raw: raw}, nil
}

func ExportScenarioFile(ctx context.Context, artifactPath string, dirs Directories) (ScenarioExport, error) {
	art, err := LoadArtifact(artifactPath)
	if err != nil {
		return ScenarioExport{}, fmt.Errorf("streamsim: %w", err)
	}
	spec, adap, err := ResolveInputs(art, dirs)
	if err != nil {
		return ScenarioExport{}, fmt.Errorf("%w", err)
	}
	exported, err := ExportScenario(ctx, art, spec, adap)
	if err != nil {
		return ScenarioExport{}, fmt.Errorf("streamsim: scenario from-run: %w", model.WithSource(err, artifactPath))
	}
	return exported, nil
}

func checkedExport(ctx context.Context, cfg Config, scenario *model.Scenario) (*model.Scenario, int, []byte, error) {
	pruned, refused, err := PruneRefusedCalls(ctx, cfg, scenario)
	if err != nil {
		return nil, 0, nil, fmt.Errorf("run: %w", err)
	}
	raw, err := encodeScenario(pruned)
	if err != nil {
		return nil, 0, nil, err
	}
	if err := CheckScenario(ctx, cfg, pruned); err != nil {
		return nil, 0, nil, fmt.Errorf("run: %w", err)
	}
	return pruned, refused, raw, nil
}

func encodeScenario(scenario *model.Scenario) ([]byte, error) {
	raw, err := json.MarshalIndent(scenario, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode scenario: %w", err)
	}
	if err := model.ValidateScenario(raw); err != nil {
		return nil, fmt.Errorf("exported scenario: %w", err)
	}
	return append(raw, '\n'), nil
}
