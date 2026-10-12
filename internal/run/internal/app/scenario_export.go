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

// PruneRefusedCalls is the scenario without the effector calls the run
// refuses at admission (world.ErrInvalidInvocation), and how many it dropped.
// A recorded run keeps a consumer's call to an entity that has gone, or with
// arguments the effector refuses; the world refused it before recording or
// drawing anything, so the scenario without it reproduces the same trace and
// world. Every other command is kept, for the scenario check to judge.
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

// refusedAtAdmission executes one command on the shadow run and reports an
// effector call the world refused at admission.
func refusedAtAdmission(ctx context.Context, shadow *Run, step model.ScenarioCommand) bool {
	command := model.Command{Op: step.Op, Args: step.Args}
	return errors.Is(executeCommand(ctx, shadow, &command, keepWorldRefusals), world.ErrInvalidInvocation)
}

// CheckArtifactInputs reports whether the domain and adapter are the ones a
// run artifact recorded, by id, version and digest; a difference wraps
// ErrDigestMismatch.
func CheckArtifactInputs(art *model.RunArtifact, spec *domain.Compiled, adapterSpec *model.Adapter) error {
	return validateReplayInputs(art, spec, adapterSpec)
}

// ScenarioExport is a run expressed as a scenario: the scenario, the number of
// command-log entries it omits, the number of refused effector calls it drops,
// and its file content.
type ScenarioExport struct {
	Scenario *model.Scenario
	Omitted  int
	Refused  int
	Raw      []byte
}

// ExportScenario is the scenario that reproduces a recorded run (SCENARIO_FILE
// §4). The domain and adapter must still be the ones the artifact recorded (an
// error wrapping ErrDigestMismatch otherwise), and the scenario is checked
// against its contract and then as a scenario run would check it. A run that
// cannot be expressed as a runnable scenario exports nothing.
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

// ExportScenarioFile loads the run artifact at artifactPath and the domain and
// adapter it recorded from the directories, and exports the run as a scenario.
// An error in the export names the artifact as its source.
func ExportScenarioFile(ctx context.Context, artifactPath string, dirs Directories) (ScenarioExport, error) {
	art, err := LoadArtifact(artifactPath)
	if err != nil {
		return ScenarioExport{}, fmt.Errorf("streamsim: %w", err)
	}
	spec, adap, err := LoadInputs(dirs.Domains, dirs.Adapters, art.Domain.ID, art.Adapter.ID)
	if err != nil {
		return ScenarioExport{}, fmt.Errorf("%w", err)
	}
	exported, err := ExportScenario(ctx, art, spec, adap)
	if err != nil {
		return ScenarioExport{}, fmt.Errorf("streamsim: scenario from-run: %w", model.WithSource(err, artifactPath))
	}
	return exported, nil
}

// checkedExport is the scenario without refused calls, how many were
// dropped, and its file content, once both checks pass.
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

// encodeScenario is the scenario's file content, checked against its own
// contract.
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
