package run

// Replay and verify: the run artifact is the single file that reproduces a
// run. Replaying the command log — not re-issuing the original MCP calls —
// is what makes an improvised session exactly reproducible, and it needs no
// server.

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
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
	var doc any
	if err := model.DecodeBytes(raw, &doc); err != nil {
		return nil, fmt.Errorf("run: %s not valid JSON: %w", path, err)
	}
	if err := model.ValidateRunArtifact(raw); err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	var art model.RunArtifact
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&art); err != nil {
		return nil, fmt.Errorf("run: decode artifact: %w", err)
	}
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		if err == nil {
			return nil, fmt.Errorf("run: artifact has trailing JSON")
		}
		return nil, fmt.Errorf("run: artifact trailing data: %w", err)
	}
	return &art, nil
}

// ReplayArtifact re-executes a run artifact's command log against a fresh
// world and compares the delivered trace to the expected digest. sinkTarget
// selects the replay sink ("" = inproc).
func ReplayArtifact(ctx context.Context, art *model.RunArtifact, spec *domain.Compiled, adapterSpec *model.Adapter, sinkTarget string) (*ReplayResult, error) {
	if art == nil {
		return nil, fmt.Errorf("run: replay requires an artifact")
	}
	var err error
	if spec == nil && len(art.DomainSpec) > 0 {
		spec, err = domain.Parse(art.DomainSpec, "run-artifact.domain_spec")
		if err != nil {
			return nil, fmt.Errorf("run: embedded domain spec: %w", err)
		}
	}
	if adapterSpec == nil && len(art.AdapterSpec) > 0 {
		adapterSpec, err = adapter.LoadBytes(art.AdapterSpec, "run-artifact.adapter_spec")
		if err != nil {
			return nil, fmt.Errorf("run: embedded adapter spec: %w", err)
		}
	}
	if spec == nil || adapterSpec == nil {
		return nil, fmt.Errorf("run: replay requires domain and adapter, or their embedded artifact specs")
	}
	res := &ReplayResult{
		WantDigest:   art.ExpectedTraceDigest,
		VersionMatch: art.SimVersion == model.SimVersion,
	}
	if spec.Spec.ID != art.Domain.ID || spec.Spec.Version != art.Domain.Version || spec.Digest != art.Domain.Digest {
		return nil, fmt.Errorf("run: domain digest mismatch: artifact=%s current=%s", art.Domain.Digest, spec.Digest)
	}
	currentAdapterDigest := adapterDigest(adapterSpec)
	if adapterSpec.ID != art.Adapter.ID || adapterSpec.Version != art.Adapter.Version || currentAdapterDigest != art.Adapter.Digest {
		return nil, fmt.Errorf("run: adapter digest mismatch: artifact=%s current=%s", art.Adapter.Digest, currentAdapterDigest)
	}
	if !res.VersionMatch {
		res.Detail = fmt.Sprintf("artifact built by sim %s, current sim %s; a different version may legitimately differ", art.SimVersion, model.SimVersion)
	}
	cfg := Config{
		Domain:          spec,
		Adapter:         adapterSpec,
		Seed:            art.Seed,
		SinkName:        model.SinkInproc,
		SinkTarget:      sinkTarget,
		TimeMode:        art.TimeMode,
		StartTimeNS:     art.WorldConfig.StartTimeNS,
		StartTimeSet:    true,
		EntityIDs:       art.WorldConfig.EntityIDs,
		ScenarioProfile: art.WorldConfig.ScenarioProfile,
		RunID:           art.RunID,
		Label:           "replay",
	}
	if sinkTarget != "" {
		cfg.SinkName = model.SinkFile
	}
	r, err := New(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	if art.WorldDigest != "" && r.Digest() != art.WorldDigest {
		return nil, fmt.Errorf("run: world digest mismatch: artifact=%s current=%s", art.WorldDigest, r.Digest())
	}
	for _, cmd := range art.CommandLog {
		if err := executeCommand(ctx, r, &cmd); err != nil {
			return nil, fmt.Errorf("run: replay command %d (%s): %w", cmd.Seq, cmd.Op, err)
		}
	}
	_, err = r.End("")
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	res.GotDigest = r.TraceDigest()
	res.Emitted = r.World.EmittedCount()
	res.Incomplete = art.Incomplete
	if art.Incomplete && res.Detail == "" {
		res.Detail = art.Error
	}
	if res.GotDigest != res.WantDigest {
		idx := firstDivergentRecord(r, art)
		if idx >= 0 {
			res.FirstDivergence = &idx
		}
		return res, nil
	}
	res.Matches = true
	return res, nil
}

// executeCommand applies one logged command to a run (replay path).
func executeCommand(ctx context.Context, r *Run, cmd *model.Command) error {
	args := cmd.Args
	str := func(k string) string {
		if v, ok := args[k].(string); ok {
			return v
		}
		return ""
	}
	num := func(k string) int64 {
		switch v := args[k].(type) {
		case float64:
			return int64(v)
		case int64:
			return v
		case json.Number:
			n, err := v.Int64()
			if err == nil {
				return n
			}
		}
		return 0
	}
	switch cmd.Op {
	case model.OpClockAdvance:
		_, err := r.Advance(ctx, num("to_ns"), false)
		if err != nil {
			return fmt.Errorf("streamsim: %w", err)
		}
		return nil
	case model.OpFaultInject:
		_, err := r.InjectFault(str("entity_id"), str("fault"), num("onset_ns"), asMap(args["params"]))
		if err != nil {
			return fmt.Errorf("streamsim: %w", err)
		}
		return nil
	case model.OpFaultClear:
		return r.ClearFault(str("fault_id"), num("at_ns"))
	case model.OpPerturbApply:
		_, err := r.ApplyPerturb(str("perturbation"), asMap(args["params"]), num("from_ns"), num("until_ns"))
		if err != nil {
			return fmt.Errorf("streamsim: %w", err)
		}
		return nil
	case model.OpPerturbClear:
		return r.ClearPerturb(str("perturb_id"))
	case model.OpEffectorInvoke:
		_, err := r.InvokeEffector(str("effector"), str("entity_id"), str("command_id"), asMap(args["args"]), num("at_ns"))
		if err != nil {
			return fmt.Errorf("streamsim: %w", err)
		}
		return nil
	case model.OpEntityAdd:
		return r.AddEntity(str("entity_id"), num("at_ns"))
	case model.OpEntityRetire:
		return r.RetireEntity(str("entity_id"), str("reason"), num("at_ns"))
	case model.OpEnvInject:
		// Environment faults act on the consumer's process, which replay has
		// no right to touch; the command is replayed as a record.
		return nil
	case model.OpWorldCreate, model.OpRunBegin, model.OpRunEnd, model.OpClockRun:
		return nil
	}
	return fmt.Errorf("run: unknown command op %q", cmd.Op)
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
