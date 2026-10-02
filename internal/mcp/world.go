package mcp

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"path/filepath"
	"strconv"
)

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
	// start_time is presence-aware: absent means the documented default, an
	// explicit 0 means epoch-0 (a legal start the CLI default must not mask).
	startNS := model.DefaultStartTimeNS
	startTimeSet := false
	if _, ok := args["start_time"]; ok {
		startNS = num(args, "start_time", 0)
		startTimeSet = true
	}
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
	worldID := "w-" + strconv.Itoa(seq)
	cfg := run.Config{
		Domain: spec, Adapter: adap, Seed: seed, SinkName: sinkName,
		SinkTarget: str(args, "sink_target"),
		TimeMode:   timeMode, StartTimeNS: startNS, StartTimeSet: startTimeSet,
		EntityIDs:       entityIDs,
		ScenarioProfile: str(args, "scenario_profile"),
		Label:           str(args, "label"),
		RunID:           "r-" + strconv.Itoa(seq),
		LedgerPath:      filepath.Join(d.OutDir, worldID, "ledger.jsonl"),
	}
	r, err := run.New(d.ctx, cfg)
	if err != nil {
		return nil, errTool(CodeDomainInvalid, "%v", err)
	}
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
	d.byToken[token] = rec
	operatorEndpoint := d.OperatorEndpoint
	d.mu.Unlock()
	out := map[string]any{
		"world_id": worldID, "world_digest": r.Digest(), "entity_ids": r.World.EntityIDs(),
		"clock": model.FormatTime(r.World.Clock()), "token": token,
		"simulated": true,
	}
	if operatorEndpoint != "" {
		out["operator_endpoint"] = operatorEndpoint
	}
	return out, nil
}

func capabilityToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("read capability token randomness: %w", err)
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
