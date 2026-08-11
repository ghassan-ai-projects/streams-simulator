package mcp

// The director role: build, drive, perturb and audit worlds; hold the
// truth. The handlers stay thin — business logic lives in world, perturb,
// adapter, truth, score and run; these methods translate MCP arguments into
// those packages' calls and shape the responses.

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/ghassan-ai-projects/streams-simulator/internal/audit"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/score"
	"github.com/ghassan-ai-projects/streams-simulator/internal/truth"
)

// WorldRecord is one created world under the director's control.
type WorldRecord struct {
	Run       *run.Run
	Token     string
	Nameplate *Nameplate
	Operator  *OperatorView
	Started   bool
	Ended     bool
	RunEnded  bool
}

// Director holds the director-role state: the catalog, the installed
// adapters, the world registry, and the truth store.
type Director struct {
	mu       sync.Mutex
	Catalog  *domain.Catalog
	Adapters map[string]*model.Adapter
	Worlds   map[string]*WorldRecord
	byRun    map[string]string // run id -> world id
	Truth    *truth.Store
	OutDir   string
	ctx      context.Context
	seq      int
}

// NewDirector builds the director with its registries.
func NewDirector(ctx context.Context, cat *domain.Catalog, adapters map[string]*model.Adapter, outDir string) *Director {
	d := &Director{
		Catalog: cat, Adapters: adapters, Worlds: map[string]*WorldRecord{},
		byRun: map[string]string{},
		Truth: truth.NewStore(), OutDir: outDir, ctx: ctx,
	}
	d.Truth.OpenChecker = func(runID string) bool {
		d.mu.Lock()
		defer d.mu.Unlock()
		worldID, ok := d.byRun[runID]
		if !ok {
			return false
		}
		w := d.Worlds[worldID]
		return w != nil && !w.RunEnded
	}
	return d
}

// World returns the record for a world id, or nil.
func (d *Director) World(worldID string) *WorldRecord {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.Worlds[worldID]
}

// CreateWorld builds a world (sim.world.create).
func (d *Director) CreateWorld(args map[string]any) (map[string]any, error) {
	domainID := str(args, "domain")
	spec, err := d.Catalog.Describe(domainID)
	if err != nil {
		return nil, errTool(CodeDomainInvalid, "%v", err)
	}
	adapterID := str(args, "adapter")
	if adapterID == "" {
		adapterID = "native-jsonl"
	}
	adap, ok := d.Adapters[adapterID]
	if !ok {
		return nil, errTool(CodeAdapterInvalid, "unknown adapter %q", adapterID)
	}
	// #nosec G115 -- the seed arg is a documented uint64 range value.
	seed := uint64(num(args, "seed", 1))
	sinkName := str(args, "sink")
	if sinkName == "" {
		sinkName = model.SinkInproc
	}
	timeMode := str(args, "time_mode")
	if timeMode == "" {
		timeMode = model.TimeStepped
	}
	startNS := num(args, "start_time", model.DefaultStartTimeNS)
	var entityIDs []string
	if ents, ok := args["entities"].([]any); ok {
		for _, e := range ents {
			if s, ok := e.(string); ok {
				entityIDs = append(entityIDs, s)
			}
		}
	}
	// Reserve a director-local identity before constructing the run. Two
	// worlds with identical simulation inputs are still distinct resources;
	// using the deterministic default run id here would overwrite the first
	// world in the registry and make truth lookup ambiguous.
	d.mu.Lock()
	d.seq++
	seq := d.seq
	d.mu.Unlock()
	cfg := run.Config{
		Domain: spec, Adapter: adap, Seed: seed, SinkName: sinkName,
		TimeMode: timeMode, StartTimeNS: startNS, EntityIDs: entityIDs,
		ScenarioProfile: str(args, "scenario_profile"),
		Label:           str(args, "label"),
		RunID:           "r-" + strconv.Itoa(seq),
	}
	r, err := run.New(d.ctx, cfg)
	if err != nil {
		return nil, errTool(CodeDomainInvalid, "%v", err)
	}
	worldID := "w-" + strconv.Itoa(seq)
	token, err := capabilityToken()
	if err != nil {
		return nil, errTool(CodeDomainInvalid, "capability token generation failed: %v", err)
	}
	nameplate := buildNameplate(r)
	nameplate.WorldID = worldID
	ov := NewOperatorView(worldID, token, nameplate, r, r)
	rec := &WorldRecord{Run: r, Token: token, Nameplate: nameplate, Operator: ov}
	d.mu.Lock()
	d.Worlds[worldID] = rec
	d.byRun[r.ID] = worldID
	d.mu.Unlock()
	return map[string]any{
		"world_id": worldID, "world_digest": r.Digest(), "entity_ids": r.World.EntityIDs(),
		"clock": model.FormatTime(r.World.Clock()), "token": token,
		"simulated": true,
	}, nil
}

func capabilityToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return "t-" + base64.RawURLEncoding.EncodeToString(buf), nil
}

func buildNameplate(r *run.Run) *Nameplate {
	spec := r.Domain()
	np := &Nameplate{
		WorldID: r.ID, Domain: spec.Spec.ID, DomainVersion: spec.Spec.Version,
		Simulated: true,
	}
	for _, id := range r.World.EntityIDs() {
		ent := r.World.Entity(id)
		np.Entities = append(np.Entities, NameplateEntity{ID: id, Type: ent.Type, BornAtNS: ent.BornNS})
	}
	for i := range spec.Spec.Channels {
		ch := &spec.Spec.Channels[i]
		nc := NameplateChannel{Name: ch.Name, Unit: ch.Unit, Resolution: ch.Resolution}
		if ch.Range != nil {
			nc.RangeMin, nc.RangeMax = ch.Range.Min, ch.Range.Max
		}
		np.Channels = append(np.Channels, nc)
	}
	for i := range spec.Spec.Effectors {
		e := &spec.Spec.Effectors[i]
		risk := e.RiskClass
		if risk == "" {
			risk = "medium"
		}
		np.Effectors = append(np.Effectors, EffectorInfo{
			Name: e.Name, Description: e.Description, ArgsSchema: e.ArgsSchema, RiskClass: risk,
		})
	}
	return np
}

// DescribeWorld reports config, digest, clock and emitted count.
func (d *Director) DescribeWorld(worldID string) (map[string]any, error) {
	w := d.World(worldID)
	if w == nil {
		return nil, errTool(CodeWorldNotFound, "unknown world %q", worldID)
	}
	return map[string]any{
		"world_id": worldID, "domain": w.Run.Domain().Spec.ID,
		"seed": float64(w.Run.Config.Seed), "clock": model.FormatTime(w.Run.World.Clock()),
		"emitted": w.Run.World.EmittedCount(), "simulated": true,
	}, nil
}

// DestroyWorld finalizes and removes a world.
func (d *Director) DestroyWorld(worldID string) (map[string]any, error) {
	w := d.World(worldID)
	if w == nil {
		return nil, errTool(CodeWorldNotFound, "unknown world %q", worldID)
	}
	if !w.RunEnded {
		dir := filepath.Join(d.OutDir, worldID)
		if _, err := w.Run.End(dir); err != nil {
			return nil, errTool(CodeDomainInvalid, "%v", err)
		}
	}
	d.mu.Lock()
	delete(d.Worlds, worldID)
	d.mu.Unlock()
	return map[string]any{"world_id": worldID, "destroyed": true}, nil
}

// Advance moves the clock (sim.clock.advance).
func (d *Director) Advance(worldID string, toNS int64, await bool) (map[string]any, error) {
	w := d.World(worldID)
	if w == nil {
		return nil, errTool(CodeWorldNotFound, "unknown world %q", worldID)
	}
	emitted, err := w.Run.Advance(toNS, await)
	if err != nil {
		return nil, errTool(CodeClockBackwards, "%v", err)
	}
	return map[string]any{
		"emitted": emitted, "clock": model.FormatTime(w.Run.World.Clock()),
		"effects_applied": w.Run.World.ActiveFaultsCount(), "simulated": true,
	}, nil
}

// ClockState reports the clock and queue.
func (d *Director) ClockState(worldID string) (map[string]any, error) {
	w := d.World(worldID)
	if w == nil {
		return nil, errTool(CodeWorldNotFound, "unknown world %q", worldID)
	}
	return map[string]any{
		"clock":             model.FormatTime(w.Run.World.Clock()),
		"next_scheduled_ns": w.Run.World.NextEventNS(),
		"pending_effects":   w.Run.World.PendingKicks(),
	}, nil
}

// InjectFault applies a world fault.
func (d *Director) InjectFault(worldID, entityID, fault string, onsetNS int64, params map[string]any) (map[string]any, error) {
	w := d.World(worldID)
	if w == nil {
		return nil, errTool(CodeWorldNotFound, "unknown world %q", worldID)
	}
	fid, err := w.Run.InjectFault(entityID, fault, onsetNS, params)
	if err != nil {
		return nil, errTool(CodeDomainInvalid, "%v", err)
	}
	return map[string]any{"fault_id": fid, "simulated": true}, nil
}

// ClearFault clears a fault.
func (d *Director) ClearFault(worldID, faultID string) (map[string]any, error) {
	w := d.World(worldID)
	if w == nil {
		return nil, errTool(CodeWorldNotFound, "unknown world %q", worldID)
	}
	if err := w.Run.ClearFault(faultID, 0); err != nil {
		return nil, errTool(CodeDomainInvalid, "%v", err)
	}
	return map[string]any{"cleared": true}, nil
}

// ListFaults is director-only: active faults, with ids and onsets.
func (d *Director) ListFaults(worldID string) (map[string]any, error) {
	w := d.World(worldID)
	if w == nil {
		return nil, errTool(CodeWorldNotFound, "unknown world %q", worldID)
	}
	return map[string]any{"faults": w.Run.World.ListFaults()}, nil
}

// ApplyPerturb activates a delivery perturbation.
func (d *Director) ApplyPerturb(worldID, name string, params map[string]any, fromNS, untilNS int64) (map[string]any, error) {
	w := d.World(worldID)
	if w == nil {
		return nil, errTool(CodeWorldNotFound, "unknown world %q", worldID)
	}
	id, err := w.Run.ApplyPerturb(name, params, fromNS, untilNS)
	if err != nil {
		return nil, errTool(CodeDomainInvalid, "%v", err)
	}
	return map[string]any{"perturb_id": id}, nil
}

// ClearPerturb deactivates a perturbation.
func (d *Director) ClearPerturb(worldID, id string) (map[string]any, error) {
	w := d.World(worldID)
	if w == nil {
		return nil, errTool(CodeWorldNotFound, "unknown world %q", worldID)
	}
	if err := w.Run.ClearPerturb(id); err != nil {
		return nil, errTool(CodeDomainInvalid, "%v", err)
	}
	return map[string]any{"cleared": true}, nil
}

// EnvInject records an environment fault against a configured target.
func (d *Director) EnvInject(worldID, target, fault string, params map[string]any) (map[string]any, error) {
	w := d.World(worldID)
	if w == nil {
		return nil, errTool(CodeWorldNotFound, "unknown world %q", worldID)
	}
	id, err := w.Run.EnvInject(target, fault, params, w.Run.World.Clock())
	if err != nil {
		return nil, errTool(CodeDomainInvalid, "%v", err)
	}
	return map[string]any{"env_id": id}, nil
}

// BeginRun opens a run after the director has sealed its ground-truth record.
// Truth is intentionally a separate operation because the complete label is
// generated from the selected scenario and fault, not from the human label
// carried by run.begin.
func (d *Director) BeginRun(worldID, label string) (map[string]any, error) {
	w := d.World(worldID)
	if w == nil {
		return nil, errTool(CodeWorldNotFound, "unknown world %q", worldID)
	}
	if w.Started {
		return nil, errTool(CodeDomainInvalid, "a run is already open for %q", worldID)
	}
	sealed, unblinded, err := d.Truth.SealStatus(w.Run.ID)
	if err != nil || !sealed {
		return nil, errTool(CodeTruthSealed, "ground truth must be sealed before run.begin")
	}
	if unblinded {
		return nil, errTool(CodeRunUnblinded, "run is already stamped unblinded")
	}
	w.Started = true
	return map[string]any{"run_id": w.Run.ID, "truth_sealed": true}, nil
}

// SealTruth installs the director-only ground-truth record before a run is
// opened. The record is copied and cannot be mutated through this pointer.
func (d *Director) SealTruth(runID string, rec *model.GroundTruthRecord) error {
	w := d.worldByRun(runID)
	if w == nil {
		return errTool(CodeWorldNotFound, "unknown run %q", runID)
	}
	if w.Started || w.RunEnded {
		return errTool(CodeTruthSealed, "truth must be sealed before run.begin")
	}
	if err := d.Truth.Seal(runID, rec); err != nil {
		return errTool(CodeTruthSealed, "%v", err)
	}
	return nil
}

// EndRun finalizes the run and writes the artifact.
func (d *Director) EndRun(worldID string) (map[string]any, error) {
	w := d.World(worldID)
	if w == nil {
		return nil, errTool(CodeWorldNotFound, "unknown world %q", worldID)
	}
	dir := filepath.Join(d.OutDir, worldID)
	art, err := w.Run.End(dir)
	if err != nil {
		return nil, errTool(CodeDomainInvalid, "%v", err)
	}
	w.RunEnded = true
	return map[string]any{
		"run_artifact_path": filepath.Join(dir, "run.json"),
		"trace_digest":      art.ExpectedTraceDigest,
		"reproducible":      art.Reproducible,
	}, nil
}

// RevealTruth returns the sealed label; unblind permits revealing on an
// open run, permanently stamping it.
func (d *Director) RevealTruth(runID string, unblind bool) (map[string]any, error) {
	rec, err := d.Truth.Reveal(runID, unblind)
	if err != nil {
		return nil, errTool(CodeTruthSealed, "%v", err)
	}
	if unblind {
		if w := d.worldByRun(runID); w != nil {
			w.Run.Unblind()
		}
	}
	return map[string]any{"ground_truth": rec}, nil
}

// SealStatus reports sealing state.
func (d *Director) SealStatus(runID string) (map[string]any, error) {
	sealed, unblinded, err := d.Truth.SealStatus(runID)
	if err != nil {
		return nil, errTool(CodeWorldNotFound, "%v", err)
	}
	return map[string]any{"sealed": sealed, "unblinded": unblinded}, nil
}

// Score computes the scorecard for a run's submitted verdict.
func (d *Director) Score(runID string) (map[string]any, error) {
	w := d.worldByRun(runID)
	if w == nil {
		return nil, errTool(CodeWorldNotFound, "unknown run %q", runID)
	}
	if w.Run.UnblindedStamp() {
		return nil, errTool(CodeRunUnblinded, "scoring refused; the run is stamped unblinded")
	}
	gt, err := d.Truth.Reveal(runID, false)
	if err != nil {
		return nil, errTool(CodeTruthSealed, "%v", err)
	}
	sc, err := score.Score(w.Run, gt)
	if err != nil {
		return nil, errTool(CodeDomainInvalid, "%v", err)
	}
	if !w.Run.Reproducible() {
		sc.Detail = "non-reproducible run (wall-clock delivery or consumer not quiesced); hash-equality metrics refused"
	}
	return map[string]any{"scorecard": sc}, nil
}

// VerifyRun replays a run artifact and compares digests.
func (d *Director) VerifyRun(artifactPath string) (map[string]any, error) {
	art, err := run.LoadArtifact(artifactPath)
	if err != nil {
		return nil, errTool(CodeDomainInvalid, "%v", err)
	}
	spec, err := d.Catalog.Describe(art.Domain.ID)
	if err != nil {
		return nil, errTool(CodeDomainInvalid, "%v", err)
	}
	adap, ok := d.Adapters[art.Adapter.ID]
	if !ok {
		return nil, errTool(CodeAdapterInvalid, "unknown adapter %q", art.Adapter.ID)
	}
	res, err := run.ReplayArtifact(d.ctx, art, spec, adap, "")
	if err != nil {
		return nil, errTool(CodeDomainInvalid, "%v", err)
	}
	return map[string]any{
		"matches": res.Matches, "version_match": res.VersionMatch,
		"first_divergence": res.FirstDivergence, "detail": res.Detail,
	}, nil
}

// AuditScenario runs the trivial-baseline audit on one injection.
func (d *Director) AuditScenario(domainID, entityID, fault string, onsetNS, startNS int64, durationNS int64) (map[string]any, error) {
	spec, err := d.Catalog.Describe(domainID)
	if err != nil {
		return nil, errTool(CodeDomainInvalid, "%v", err)
	}
	panel := audit.NewPanel(spec, 1, 60*1e9)
	var ids []string
	n := spec.Spec.Entities.Count.Default
	for i := 1; i <= n; i++ {
		ids = append(ids, fmt.Sprintf("site-a/pond-%d", i))
	}
	v, err := panel.Audit(entityID, fault, onsetNS, startNS, ids, durationNS, nil)
	if err != nil {
		return nil, errTool(CodeDomainInvalid, "%v", err)
	}
	return map[string]any{"trivial": v.Trivial, "scores": v.Scores, "best": v.Best}, nil
}

func (d *Director) worldByRun(runID string) *WorldRecord {
	d.mu.Lock()
	defer d.mu.Unlock()
	worldID, ok := d.byRun[runID]
	if !ok {
		return nil
	}
	return d.Worlds[worldID]
}

func str(args map[string]any, key string) string {
	if args == nil {
		return ""
	}
	if s, ok := args[key].(string); ok {
		return s
	}
	return ""
}

func num(args map[string]any, key string, def int64) int64 {
	if args == nil {
		return def
	}
	switch v := args[key].(type) {
	case float64:
		return int64(v)
	case int64:
		return v
	case json.Number:
		if n, err := v.Int64(); err == nil {
			return n
		}
	case string:
		var n int64
		if _, err := fmt.Sscanf(v, "%d", &n); err == nil {
			return n
		}
	}
	return def
}
