package app

// Replay and verify: the run artifact is the single file that reproduces a
// run. Replaying the command log — not re-issuing the original MCP calls —
// is what makes an improvised session exactly reproducible, and it needs no
// server.

import (
	"context"
	"fmt"

	rules "github.com/ghassan-ai-projects/streams-simulator/internal/run/internal/domain"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run/internal/durable"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

// ReplayResult is the outcome of replaying a run artifact.
type ReplayResult struct {
	Matches         bool   `json:"matches"`
	VersionMatch    bool   `json:"version_match"`
	FirstDivergence *int   `json:"first_divergence,omitempty"` // 0-based record index; nil means no divergence
	GotDigest       string `json:"got_digest"`
	WantDigest      string `json:"want_digest"`
	Emitted         int64  `json:"emitted"`
	Incomplete      bool   `json:"incomplete,omitempty"` // the original run did not reach a clean end
	Detail          string `json:"detail,omitempty"`
}

// LoadArtifact reads and validates a run artifact.
func LoadArtifact(path string) (*model.RunArtifact, error) {
	raw, err := durable.ReadArtifact(path)
	if err != nil {
		return nil, err
	}
	if err := rules.ValidateArtifactDocument(raw, path); err != nil {
		return nil, err
	}
	return rules.DecodeArtifact(raw)
}

// ReplayArtifact re-executes a run artifact's command log against a fresh
// world and compares the delivered trace to the expected digest. sinkTarget
// selects the replay sink ("" = inproc).
func ReplayArtifact(ctx context.Context, art *model.RunArtifact, spec *domain.Compiled, adapterSpec *model.Adapter, sinkTarget string) (*ReplayResult, error) {
	spec, adapterSpec, result, err := prepareReplay(art, spec, adapterSpec)
	if err != nil {
		return nil, err
	}
	replay, err := newReplayRun(ctx, art, spec, adapterSpec, sinkTarget)
	if err != nil {
		return nil, err
	}
	if err := replayCommands(ctx, replay, art.CommandLog); err != nil {
		return nil, err
	}
	return finishReplayResult(replay, art, result), nil
}

// ReplayEvidence is what replaying an artifact recovers beyond the digest
// comparison: the effector calls the original run made, which the artifact's
// command log cannot state (applied, refused or silent).
type ReplayEvidence struct {
	Result *ReplayResult
	Calls  []world.EffectorCall
}

// ReplayArtifactEvidence replays like ReplayArtifact and also returns the
// replayed world's effector calls, so offline scoring can grade the loop.
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
	return &ReplayEvidence{Result: finishReplayResult(replay, art, result), Calls: calls}, nil
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
		return fmt.Errorf("run: domain digest mismatch: artifact=%s current=%s", art.Domain.Digest, spec.Digest)
	}
	currentAdapterDigest := rules.AdapterDigest(adapterSpec)
	if adapterSpec.ID != art.Adapter.ID || adapterSpec.Version != art.Adapter.Version || currentAdapterDigest != art.Adapter.Digest {
		return fmt.Errorf("run: adapter digest mismatch: artifact=%s current=%s", art.Adapter.Digest, currentAdapterDigest)
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

// executeCommand applies one logged command to a run (replay path).
func executeCommand(ctx context.Context, r *Run, cmd *model.Command) error {
	switch cmd.Op {
	case model.OpClockAdvance:
		return replayAdvance(ctx, r, cmd.Args)
	case model.OpFaultInject, model.OpFaultClear:
		return replayFault(r, cmd)
	case model.OpPerturbApply, model.OpPerturbClear:
		return replayPerturb(r, cmd)
	default:
		return replayActuation(r, cmd)
	}
}

// firstDivergentRecord names the first record the replay is missing or has
// beyond the original run, from the delivery counts. The artifact carries only
// the trace digest, not the records, so when the counts agree the position of
// the first differing record is unknown and -1 is returned.
func firstDivergentRecord(r *Run, art *model.RunArtifact) int {
	replayed := int64(len(r.ledger))
	if replayed == art.Counts.Emitted {
		return -1
	}
	return int(min(replayed, art.Counts.Emitted))
}
