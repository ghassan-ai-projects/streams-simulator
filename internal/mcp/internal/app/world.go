package app

import (
	"path/filepath"
	"strconv"

	"github.com/ghassan-ai-projects/streams-simulator/internal/mcp/internal/capability"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

// CreateWorld builds a world (sim.world.create).
func (d *Director) CreateWorld(args map[string]any) (map[string]any, error) {
	cfg, err := d.worldConfig(args)
	if err != nil {
		return nil, err
	}
	worldID, runID := d.reserveWorldIdentity()
	cfg.RunID = runID
	cfg.LedgerPath = filepath.Join(d.OutDir, worldID, "ledger.jsonl")
	r, err := run.New(d.ctx, cfg)
	if err != nil {
		return nil, errTool(CodeDomainInvalid, "%v", err)
	}
	return d.publishWorld(worldID, r)
}

func (d *Director) worldConfig(args map[string]any) (run.Config, error) {
	spec, err := d.Catalog.Describe(str(args, "domain"))
	if err != nil {
		return run.Config{}, errTool(CodeDomainInvalid, "%v", err)
	}
	adap, err := d.worldAdapter(args)
	if err != nil {
		return run.Config{}, err
	}
	return configuredWorld(spec, adap, args), nil
}

func (d *Director) reserveWorldIdentity() (string, string) {
	d.mu.Lock()
	d.seq++
	seq := d.seq
	d.mu.Unlock()
	worldID := "w-" + strconv.Itoa(seq)
	return worldID, "r-" + strconv.Itoa(seq)
}

func (d *Director) registerWorld(worldID string, r *run.Run, token string) string {
	nameplate := buildNameplate(r)
	nameplate.WorldID = worldID
	ov := NewOperatorView(worldID, token, nameplate, r, r)
	rec := &WorldRecord{Run: r, Token: token, Nameplate: nameplate, Operator: ov}
	d.mu.Lock()
	d.Worlds[worldID] = rec
	d.byRun[r.ID] = worldID
	d.byToken[token] = rec
	operatorEndpoint := d.OperatorEndpoint
	d.mu.Unlock()
	return operatorEndpoint
}

func buildNameplate(r *run.Run) *Nameplate {
	spec := r.Domain()
	np := &Nameplate{WorldID: r.ID, Domain: spec.Spec.ID, DomainVersion: spec.Spec.Version, Simulated: true}
	appendNameplateEntities(np, r)
	appendNameplateChannels(np, spec)
	appendNameplateEffectors(np, spec)
	return np
}

func (d *Director) publishWorld(worldID string, r *run.Run) (map[string]any, error) {
	token, err := capability.NewToken()
	if err != nil {
		return nil, errTool(CodeDomainInvalid, "capability token generation failed: %v", err)
	}
	endpoint := d.registerWorld(worldID, r, token)
	out := map[string]any{"world_id": worldID, "world_digest": r.Digest(), "entity_ids": r.World.EntityIDs(),
		"clock": model.FormatTime(r.World.Clock()), "token": token, "simulated": true}
	if endpoint != "" {
		out["operator_endpoint"] = endpoint
	}
	return out, nil
}
