package app

// Replay and verify: the run artifact is the single file that reproduces a
// run. Replaying the command log — not re-issuing the original MCP calls —
// is what makes an improvised session exactly reproducible, and it needs no
// server.

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
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
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("run: read artifact %s: %w", path, err)
	}
	if err := validateArtifactDocument(raw, path); err != nil {
		return nil, err
	}
	return decodeArtifact(raw)
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
	currentAdapterDigest := adapterDigest(adapterSpec)
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

func commandString(args map[string]any, key string) string {
	if value, ok := args[key].(string); ok {
		return value
	}
	return ""
}

func commandTime(args map[string]any, key string) int64 {
	switch value := args[key].(type) {
	case float64:
		return int64(value)
	case int64:
		return value
	case json.Number:
		if n, err := value.Int64(); err == nil {
			return n
		}
	}
	return 0
}

func asMap(v any) map[string]any {
	if m, ok := v.(map[string]any); ok {
		return m
	}
	return nil
}

// firstDivergentRecord compares the replayed trace against the original
// artifact's trace digest source by finding the first differing line. The
// original trace is not stored in the artifact (only its digest), so the
// divergence index is computed against the expected record count from the
// ledger length when available; otherwise it reports a digest mismatch.
func firstDivergentRecord(r *Run, art *model.RunArtifact) int {
	// The ledger preserves delivery order; a replayed ledger of different
	// length is the first divergence.
	if int64(len(r.ledger)) != art.Counts.Emitted {
		if int64(len(r.ledger)) < art.Counts.Emitted {
			return int(len(r.ledger))
		}
		return int(art.Counts.Emitted)
	}
	return -1
}
