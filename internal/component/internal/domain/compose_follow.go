package domain

func (c *composer) applyFollows(entries []*built, set *linkSet) error {
	for _, wanted := range set.follows {
		if err := c.applyFollow(entries, wanted); err != nil {
			return err
		}
	}
	return nil
}

func (c *composer) applyFollow(entries []*built, wanted follow) error {
	for _, entry := range entries {
		if entry.kindName != wanted.kind {
			continue
		}
		for _, dynamics := range members(entry.kind, "dynamics") {
			if dynamics["target"] == wanted.state {
				dynamics["follow"] = followDocument(wanted)
				return nil
			}
		}
	}
	return c.refuse("/connections", "kind %s has no dynamics for state %q, which a connection feeds", wanted.kind, wanted.state)
}

func followDocument(wanted follow) map[string]any {
	document := map[string]any{"link": wanted.link, "state": wanted.source, "aggregate": "sum"}
	if wanted.sampleS > 0 {
		document["sample_s"] = wanted.sampleS
	}
	return document
}
