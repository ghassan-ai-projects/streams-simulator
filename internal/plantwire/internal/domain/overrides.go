package domain

import "github.com/ghassan-ai-projects/streams-simulator/internal/model"

func overridesOf(base, effective *model.Kind) map[string]any {
	overrides := map[string]any{}
	if added := addedChannels(base, effective); len(added) > 0 {
		overrides["add"] = metricTemplates(effective, added)
	}
	if narrowed := narrowedRanges(base, effective); len(narrowed) > 0 {
		overrides["narrow"] = narrowed
	}
	return overrides
}

func addedChannels(base, effective *model.Kind) []model.Channel {
	var added []model.Channel
	for _, ch := range effective.Channels {
		if base.ChannelDecl(ch.Name) == nil {
			added = append(added, ch)
		}
	}
	return added
}

func narrowedRanges(base, effective *model.Kind) []any {
	var narrowed []any
	for i := range base.Channels {
		own, kind := effective.ChannelDecl(base.Channels[i].Name), base.Channels[i].Range
		if own != nil && own.Range != nil && (kind == nil || *own.Range != *kind) {
			narrowed = append(narrowed, map[string]any{"name": own.Name, "eu_range": map[string]any{"min": own.Range.Min, "max": own.Range.Max}})
		}
	}
	return narrowed
}
