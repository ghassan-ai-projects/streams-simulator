package app

import (
	"context"
	"fmt"

	rules "github.com/ghassan-ai-projects/streams-simulator/internal/run/internal/domain"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run/internal/durable"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

type ReplayResult struct {
	Matches         bool   `json:"matches"`
	VersionMatch    bool   `json:"version_match"`
	FirstDivergence *int   `json:"first_divergence,omitempty"`
	GotDigest       string `json:"got_digest"`
	WantDigest      string `json:"want_digest"`
	Emitted         int64  `json:"emitted"`
	Incomplete      bool   `json:"incomplete,omitempty"`
	Detail          string `json:"detail,omitempty"`

	JournalDiverged []string `json:"journal_diverged,omitempty"`

	DomainChanged       bool   `json:"domain_changed,omitempty"`
	CurrentDomainDigest string `json:"current_domain_digest,omitempty"`
}

func LoadArtifact(path string) (*model.RunArtifact, error) {
	raw, err := durable.ReadArtifact(path)
	if err != nil {
		return nil, err
	}
	if err := rules.ValidateArtifactDocument(raw, path); err != nil {
		return nil, err
	}
	artifact, err := rules.DecodeArtifact(raw)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", model.WithSource(err, path))
	}
	return artifact, nil
}

func ReplayArtifact(ctx context.Context, art *model.RunArtifact, spec *domain.Compiled, adapterSpec *model.Adapter, sinkTarget string) (*ReplayResult, error) {
	result, _, err := replayArtifactRun(ctx, art, spec, adapterSpec, sinkTarget)
	return result, err
}

func replayArtifactRun(ctx context.Context, art *model.RunArtifact, spec *domain.Compiled, adapterSpec *model.Adapter, sinkTarget string) (*ReplayResult, *Run, error) {
	spec, adapterSpec, result, err := prepareReplay(art, spec, adapterSpec)
	if err != nil {
		return nil, nil, err
	}
	replay, err := newReplayRun(ctx, art, spec, adapterSpec, sinkTarget)
	if err != nil {
		return nil, nil, err
	}
	if err := replayCommands(ctx, replay, art.CommandLog); err != nil {
		return nil, nil, err
	}
	return finishReplayResult(replay, art, result), replay, nil
}

type ReplayEvidence struct {
	Result  *ReplayResult
	Calls   []world.EffectorCall
	Ledger  []model.LedgerRecord
	History []model.StateSnapshot
}

func ReplayArtifactEvidence(ctx context.Context, art *model.RunArtifact, spec *domain.Compiled, adapterSpec *model.Adapter) (*ReplayEvidence, error) {
	spec, adapterSpec, result, err := prepareReplay(art, spec, adapterSpec)
	if err != nil {
		return nil, err
	}
	replay, err := newReplayRun(ctx, art, spec, adapterSpec, "")
	if err != nil {
		return nil, err
	}
	if err := replayCommands(ctx, replay, art.CommandLog); err != nil {
		return nil, err
	}
	calls := append([]world.EffectorCall{}, replay.World.EffectorCalls()...)
	return &ReplayEvidence{Result: finishReplayResult(replay, art, result), Calls: calls, Ledger: replay.ledger, History: replay.history}, nil
}

func replayInputs(art *model.RunArtifact, spec *domain.Compiled, adapterSpec *model.Adapter) (*domain.Compiled, *model.Adapter, error) {
	spec, err := replayDomain(art, spec)
	if err != nil {
		return nil, nil, err
	}
	adapterSpec, err = replayAdapter(art, adapterSpec)
	if err != nil {
		return nil, nil, err
	}
	if spec == nil || adapterSpec == nil {
		return nil, nil, fmt.Errorf("run: replay requires domain and adapter, or their embedded artifact specs")
	}
	return spec, adapterSpec, nil
}

func validateReplayInputs(art *model.RunArtifact, spec *domain.Compiled, adapterSpec *model.Adapter) error {
	if spec.Spec.ID != art.Domain.ID || spec.Spec.Version != art.Domain.Version || spec.Digest != art.Domain.Digest {
		return &digestMismatch{"domain", art.Domain.Digest, spec.Digest}
	}
	currentAdapterDigest := rules.AdapterDigest(adapterSpec)
	if adapterSpec.ID != art.Adapter.ID || adapterSpec.Version != art.Adapter.Version || currentAdapterDigest != art.Adapter.Digest {
		return &digestMismatch{"adapter", art.Adapter.Digest, currentAdapterDigest}
	}
	return nil
}

func replayConfig(art *model.RunArtifact, spec *domain.Compiled, adapterSpec *model.Adapter, sinkTarget string) Config {
	cfg := baseReplayConfig(art, spec, adapterSpec, sinkTarget)
	cfg.StartTimeNS = art.WorldConfig.StartTimeNS
	cfg.StartTimeSet = true
	cfg.EntityIDs = art.WorldConfig.EntityIDs
	cfg.ScenarioProfile = art.WorldConfig.ScenarioProfile
	cfg.ClockMultiplier = art.WorldConfig.ClockMultiplier
	cfg.Noiseless = art.WorldConfig.Noiseless
	cfg.ForceFailureMode = world.FailureMode(art.WorldConfig.ForceFailureMode)
	if sinkTarget != "" {
		cfg.SinkName = model.SinkFile
	}
	return cfg
}

func firstDivergentRecord(r *Run, art *model.RunArtifact) int {
	replayed := int64(len(r.ledger))
	if replayed == art.Counts.Emitted {
		return -1
	}
	return int(min(replayed, art.Counts.Emitted))
}
