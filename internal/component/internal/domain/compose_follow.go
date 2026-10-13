package domain

import (
	"maps"
	"slices"
	"strings"
)

func (c *composer) rejectCycle(set *linkSet) error {
	state := map[string]int{}
	for _, kind := range slices.Sorted(maps.Keys(set.edges)) {
		if path := cycleFrom(kind, set.edges, state, nil); path != nil {
			return c.refuse("/connections", "connections make kinds read each other in a cycle: %s", strings.Join(path, " -> "))
		}
	}
	return nil
}

func cycleFrom(kind string, edges map[string][]string, state map[string]int, trail []string) []string {
	if state[kind] == 2 {
		return nil
	}
	trail = append(trail, kind)
	if state[kind] == 1 {
		return trail[slices.Index(trail, kind):]
	}
	state[kind] = 1
	path := firstCycle(edges[kind], edges, state, trail)
	state[kind] = 2
	return path
}

func firstCycle(next []string, edges map[string][]string, state map[string]int, trail []string) []string {
	for _, kind := range next {
		if path := cycleFrom(kind, edges, state, trail); path != nil {
			return path
		}
	}
	return nil
}

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
				dynamics["follow"] = map[string]any{"link": wanted.link, "state": wanted.source, "aggregate": "sum"}
				return nil
			}
		}
	}
	return c.refuse("/connections", "kind %s has no dynamics for state %q, which a connection feeds", wanted.kind, wanted.state)
}
