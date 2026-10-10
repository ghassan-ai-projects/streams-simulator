package app

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

func (d *Director) worldAdapter(args map[string]any) (*model.Adapter, error) {
	id := str(args, "adapter")
	if id == "" {
		id = "native-jsonl"
	}
	adapter, ok := d.Adapters[id]
	if !ok {
		return nil, errTool(CodeAdapterInvalid, "unknown adapter %q", id)
	}
	return adapter, nil
}

func configuredWorld(spec *domain.Compiled, adapter *model.Adapter, args map[string]any) run.Config {
	// #nosec G115 -- the seed argument has a documented uint64 range.
	seed := uint64(num(args, "seed", 1))
	sink, timeMode := worldDeliveryOptions(args)
	start, present := worldStartTime(args)
	return run.Config{Domain: spec, Adapter: adapter, Seed: seed, SinkName: sink, SinkTarget: str(args, "sink_target"),
		TimeMode: timeMode, StartTimeNS: start, StartTimeSet: present, EntityIDs: worldEntityArguments(args),
		ScenarioProfile: str(args, "scenario_profile"), Label: str(args, "label")}
}

func worldDeliveryOptions(args map[string]any) (string, string) {
	sink := str(args, "sink")
	if sink == "" {
		sink = model.SinkInproc
	}
	mode := str(args, "time_mode")
	if mode == "" {
		mode = model.TimeStepped
	}
	return sink, mode
}

func worldStartTime(args map[string]any) (int64, bool) {
	// Presence distinguishes the default from a legal epoch-zero start.
	if _, present := args["start_time"]; present {
		return num(args, "start_time", 0), true
	}
	return model.DefaultStartTimeNS, false
}

func worldEntityArguments(args map[string]any) []string {
	var entities []string
	if items, ok := args["entities"].([]any); ok {
		for _, item := range items {
			if entity, ok := item.(string); ok {
				entities = append(entities, entity)
			}
		}
	}
	return entities
}

func appendNameplateEntities(np *Nameplate, r *run.Run) {
	for _, id := range r.World.EntityIDs() {
		entity := r.World.Entity(id)
		np.Entities = append(np.Entities, NameplateEntity{ID: id, Type: entity.Type, BornAtNS: entity.BornNS})
	}
}

func appendNameplateChannels(np *Nameplate, spec *domain.Compiled) {
	for i := range spec.Spec.Channels {
		channel := &spec.Spec.Channels[i]
		item := NameplateChannel{Name: channel.Name, Unit: channel.Unit, Resolution: channel.Resolution}
		if channel.Range != nil {
			item.RangeMin, item.RangeMax = channel.Range.Min, channel.Range.Max
		}
		np.Channels = append(np.Channels, item)
	}
}

func appendNameplateEffectors(np *Nameplate, spec *domain.Compiled) {
	for i := range spec.Spec.Effectors {
		effector := &spec.Spec.Effectors[i]
		risk := effector.RiskClass
		if risk == "" {
			risk = "medium"
		}
		np.Effectors = append(np.Effectors, EffectorInfo{Name: effector.Name, Description: effector.Description,
			ArgsSchema: effector.ArgsSchema, RiskClass: risk})
	}
}
