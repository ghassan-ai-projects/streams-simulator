package app

import (
	rules "github.com/ghassan-ai-projects/streams-simulator/internal/director/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

// WorldRequest is the typed input of CreateWorld: what a caller asks for,
// decoded by the surface. Empty strings select the documented defaults; a nil
// StartTimeNS selects the default start, which is distinct from an explicit
// epoch-zero start.
type WorldRequest struct {
	Domain, Adapter, Sink, SinkTarget, TimeMode string
	Seed                                        int64
	StartTimeNS                                 *int64
	Entities                                    []string
	ScenarioProfile, Label                      string
}

func (d *Director) worldAdapter(req WorldRequest) (*model.Adapter, error) {
	id := req.Adapter
	if id == "" {
		id = "native-jsonl"
	}
	adapter, ok := d.Adapters[id]
	if !ok {
		return nil, rules.ToolErrorf(rules.CodeAdapterInvalid, "unknown adapter %q", id)
	}
	return adapter, nil
}

func configuredWorld(spec *domain.Compiled, adapter *model.Adapter, req WorldRequest) run.Config {
	// #nosec G115 -- the seed argument has a documented uint64 range.
	seed := uint64(req.Seed)
	sink, timeMode := worldDeliveryOptions(req)
	start, present := worldStartTime(req)
	return run.Config{Domain: spec, Adapter: adapter, Seed: seed, SinkName: sink, SinkTarget: req.SinkTarget,
		TimeMode: timeMode, StartTimeNS: start, StartTimeSet: present, EntityIDs: req.Entities,
		ScenarioProfile: req.ScenarioProfile, Label: req.Label}
}

func worldDeliveryOptions(req WorldRequest) (string, string) {
	sink := req.Sink
	if sink == "" {
		sink = model.SinkInproc
	}
	mode := req.TimeMode
	if mode == "" {
		mode = model.TimeStepped
	}
	return sink, mode
}

func worldStartTime(req WorldRequest) (int64, bool) {
	// Presence distinguishes the default from a legal epoch-zero start.
	if req.StartTimeNS != nil {
		return *req.StartTimeNS, true
	}
	return model.DefaultStartTimeNS, false
}

func appendNameplateEntities(np *rules.Nameplate, r *run.Run) {
	for _, id := range r.World.EntityIDs() {
		entity := r.World.Entity(id)
		np.Entities = append(np.Entities, rules.NameplateEntity{ID: id, Type: entity.Type, BornAtNS: entity.BornNS})
	}
}

func appendNameplateChannels(np *rules.Nameplate, spec *domain.Compiled) {
	for i := range spec.Spec.Channels {
		channel := &spec.Spec.Channels[i]
		item := rules.NameplateChannel{Name: channel.Name, Unit: channel.Unit, Resolution: channel.Resolution}
		if channel.Range != nil {
			item.RangeMin, item.RangeMax = channel.Range.Min, channel.Range.Max
		}
		np.Channels = append(np.Channels, item)
	}
}

func appendNameplateEffectors(np *rules.Nameplate, spec *domain.Compiled) {
	for i := range spec.Spec.Effectors {
		effector := &spec.Spec.Effectors[i]
		risk := effector.RiskClass
		if risk == "" {
			risk = "medium"
		}
		np.Effectors = append(np.Effectors, rules.EffectorInfo{Name: effector.Name, Description: effector.Description,
			ArgsSchema: effector.ArgsSchema, RiskClass: risk})
	}
}
