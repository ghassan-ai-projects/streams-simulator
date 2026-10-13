package domain

func (c *composer) assemble(entries []*built, links *linkSet) map[string]any {
	document := c.baseDocument(entries)
	if len(links.links) > 0 {
		document["link_types"], document["links"] = links.types, links.links
	}
	if c.composition.Site != nil {
		document["site"] = c.site()
	}
	if c.composition.Description != "" {
		document["description"] = c.composition.Description
	}
	return document
}

func (c *composer) baseDocument(entries []*built) map[string]any {
	composition := c.composition
	return map[string]any{"id": composition.ID, "version": composition.Version, "title": composition.Title,
		"axes": composition.Axes, "stresses": composition.Stresses, "kinds": kindList(entries), "instances": instanceList(entries),
		"profiles": composition.Profiles, "ground_truth": composition.GroundTruth}
}

func kindList(entries []*built) []any {
	kinds := make([]any, len(entries))
	for i, entry := range entries {
		kinds[i] = entry.kind
	}
	return kinds
}

func instanceList(entries []*built) []any {
	sets := make([]any, len(entries))
	for i, entry := range entries {
		sets[i] = entry.instances
	}
	return sets
}

func (c *composer) site() map[string]any {
	site := map[string]any{"name": c.composition.Site.Name}
	if c.composition.Site.Timezone != "" {
		site["timezone"] = c.composition.Site.Timezone
	}
	if len(c.composition.Areas) > 0 {
		site["areas"] = c.areaList()
	}
	return site
}

func (c *composer) areaList() []any {
	areas := make([]any, len(c.composition.Areas))
	for i, area := range c.composition.Areas {
		entry := map[string]any{"name": area.Name}
		if area.Parent != "" {
			entry["parent"] = area.Parent
		}
		areas[i] = entry
	}
	return areas
}
