package domain

func (c *composer) applyFollows(entries []*built, set *linkSet) error {
	for _, group := range followGroups(set.follows) {
		if err := c.applyFollow(entries, group); err != nil {
			return err
		}
	}
	return nil
}

func followGroups(follows []follow) [][]follow {
	var groups [][]follow
	at := map[[2]string]int{}
	for _, wanted := range follows {
		key := [2]string{wanted.kind, wanted.state}
		if index, found := at[key]; found {
			groups[index] = append(groups[index], wanted)
			continue
		}
		at[key] = len(groups)
		groups = append(groups, []follow{wanted})
	}
	return groups
}

func (c *composer) applyFollow(entries []*built, group []follow) error {
	wanted := group[0]
	for _, entry := range entries {
		if entry.kindName != wanted.kind {
			continue
		}
		for _, dynamics := range members(entry.kind, "dynamics") {
			if dynamics["target"] == wanted.state {
				dynamics["follow"] = followDocument(group)
				return nil
			}
		}
	}
	return c.refuse("/connections", "kind %s has no dynamics for state %q, which a connection feeds", wanted.kind, wanted.state)
}

func followDocument(group []follow) map[string]any {
	sources := make([]any, 0, len(group))
	for _, wanted := range group {
		sources = append(sources, map[string]any{"link": wanted.link, "state": wanted.source})
	}
	document := map[string]any{"sources": sources, "aggregate": "sum"}
	if group[0].sampleS > 0 {
		document["sample_s"] = group[0].sampleS
	}
	return document
}
