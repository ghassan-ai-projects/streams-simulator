package run

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func replayDomain(art *model.RunArtifact, spec *domain.Compiled) (*domain.Compiled, error) {
	if spec != nil || len(art.DomainSpec) == 0 {
		return spec, nil
	}
	spec, err := domain.Parse(art.DomainSpec, "run-artifact.domain_spec")
	if err != nil {
		return nil, fmt.Errorf("run: embedded domain spec: %w", err)
	}
	return spec, nil
}

func replayAdapter(art *model.RunArtifact, spec *model.Adapter) (*model.Adapter, error) {
	if spec != nil || len(art.AdapterSpec) == 0 {
		return spec, nil
	}
	spec, err := adapter.LoadBytes(art.AdapterSpec, "run-artifact.adapter_spec")
	if err != nil {
		return nil, fmt.Errorf("run: embedded adapter spec: %w", err)
	}
	return spec, nil
}

func baseReplayConfig(art *model.RunArtifact, spec *domain.Compiled, adapterSpec *model.Adapter, sinkTarget string) Config {
	return Config{
		Domain:     spec,
		Adapter:    adapterSpec,
		Seed:       art.Seed,
		SinkName:   model.SinkInproc,
		SinkTarget: sinkTarget,
		TimeMode:   art.TimeMode,
		RunID:      art.RunID,
		Label:      "replay",
	}
}

func prepareReplay(art *model.RunArtifact, spec *domain.Compiled, adapterSpec *model.Adapter) (*domain.Compiled, *model.Adapter, *ReplayResult, error) {
	if art == nil {
		return nil, nil, nil, fmt.Errorf("run: replay requires an artifact")
	}
	spec, adapterSpec, err := replayInputs(art, spec, adapterSpec)
	if err != nil {
		return nil, nil, nil, err
	}
	result := &ReplayResult{WantDigest: art.ExpectedTraceDigest, VersionMatch: art.SimVersion == model.SimVersion}
	if err := validateReplayInputs(art, spec, adapterSpec); err != nil {
		return nil, nil, nil, err
	}
	annotateReplayVersion(result, art)
	return spec, adapterSpec, result, nil
}

func annotateReplayVersion(result *ReplayResult, artifact *model.RunArtifact) {
	if !result.VersionMatch {
		result.Detail = fmt.Sprintf("artifact built by sim %s, current sim %s; a different version may legitimately differ", artifact.SimVersion, model.SimVersion)
	}
}
