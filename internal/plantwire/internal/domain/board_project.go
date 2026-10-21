package domain

import (
	"fmt"
	"maps"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

type Interface struct {
	Kind   *model.Kind
	Mapped MappedKind
}

type Interfaces map[string]Interface

func (f Interfaces) Of(kind *model.Kind) (*model.Kind, *MappedKind) {
	if kind == nil {
		return nil, nil
	}
	face, mapped := f[kind.Name]
	if !mapped {
		return kind, nil
	}
	return face.Kind, &face.Mapped
}

var valueTypes = map[string]string{"number": "number", "integer": "counter", "boolean": "boolean", "enum": "string", "string": "string"}

type projection struct {
	raw    *model.Kind
	mapped MappedKind
}

func ProjectKind(raw *model.Kind, mapped MappedKind, file map[string]any) (*model.Kind, error) {
	projected, err := projection{raw: raw, mapped: mapped}.kind(file)
	if err != nil {
		return nil, fmt.Errorf("kind %q differs from its kind file %q through the board mapping %q at %w: %w", raw.Name, raw.Implements, raw.BoardType, err, ErrKindDrift)
	}
	return projected, nil
}

func (p projection) kind(file map[string]any) (*model.Kind, error) {
	channels, err := p.channels(listOf(file["metrics"]))
	if err != nil {
		return nil, err
	}
	effectors, err := p.effectors(listOf(file["commands"]))
	if err != nil {
		return nil, err
	}
	projected := *p.raw
	projected.Channels, projected.Effectors = channels, effectors
	return &projected, nil
}

func (p projection) channels(metrics []any) ([]model.Channel, error) {
	channels := make([]model.Channel, 0, len(metrics))
	for _, entry := range metrics {
		channel, err := p.channel(objectOf(entry))
		if err != nil {
			return nil, err
		}
		channels = append(channels, channel)
	}
	return channels, unmapped("metrics", p.mapped.Metrics, metricNames(channels))
}

func (p projection) channel(metric map[string]any) (model.Channel, error) {
	name, _ := metric["name"].(string)
	mapped, known := p.mapped.Metrics[name]
	if !known {
		return model.Channel{}, fmt.Errorf("/metrics/%s: the board mapping does not map it", name)
	}
	raw := p.raw.ChannelDecl(mapped.Raw)
	if raw == nil {
		return model.Channel{}, fmt.Errorf("/metrics/%s/raw: %q is not a channel of the kind", name, mapped.Raw)
	}
	if err := readingFits(name, metric, mapped, raw, p.mapped.RawEnums[mapped.Raw]); err != nil {
		return model.Channel{}, err
	}
	return semanticChannel(*raw, metric), nil
}

func semanticChannel(channel model.Channel, metric map[string]any) model.Channel {
	channel.Name, _ = metric["name"].(string)
	channel.Unit, _ = metric["unit"].(string)
	channel.Resolution, _ = numberOf(metric["resolution"])
	channel.Range = channelRange(metric["eu_range"])
	channel.EnumValues = wordsOf(metric["enum_values"])
	dataType, _ := metric["data_type"].(string)
	channel.ValueType, channel.States, channel.Description = valueTypes[dataType], nil, ""
	return channel
}

func (p projection) effectors(commands []any) ([]model.Effector, error) {
	effectors := make([]model.Effector, 0, len(commands))
	for _, entry := range commands {
		effector, err := p.effector(objectOf(entry))
		if err != nil {
			return nil, err
		}
		effectors = append(effectors, effector)
	}
	names := make([]string, len(effectors))
	for i := range effectors {
		names[i] = effectors[i].Name
	}
	return effectors, unmapped("commands", p.mapped.Commands, names)
}

func (p projection) effector(command map[string]any) (model.Effector, error) {
	name, _ := command["name"].(string)
	mapped, known := p.mapped.Commands[name]
	if !known {
		return model.Effector{}, fmt.Errorf("/commands/%s: the board mapping does not map it", name)
	}
	return p.semanticEffector(name, mapped, objectOf(command["args_schema"]))
}

func (p projection) semanticEffector(name string, mapped MappedCommand, schema map[string]any) (model.Effector, error) {
	raw := p.raw.EffectorDecl(mapped.Raw)
	if raw == nil {
		return model.Effector{}, fmt.Errorf("/commands/%s/raw: %q is not an effector of the kind", name, mapped.Raw)
	}
	if err := commandFits("/commands/"+name, schema, mapped, raw.ArgsSchema, p.mapped.RawEnums); err != nil {
		return model.Effector{}, err
	}
	effector := *raw
	effector.Name, effector.ArgsSchema, effector.Description = name, schema, ""
	return effector, nil
}

func unmapped[V any](member string, mapped map[string]V, declared []string) error {
	for _, name := range slices.Sorted(maps.Keys(mapped)) {
		if !slices.Contains(declared, name) {
			return fmt.Errorf("/%s/%s: the board mapping maps it, the kind file has no such member", member, name)
		}
	}
	return nil
}

func metricNames(channels []model.Channel) []string {
	names := make([]string, len(channels))
	for i := range channels {
		names[i] = channels[i].Name
	}
	return names
}
