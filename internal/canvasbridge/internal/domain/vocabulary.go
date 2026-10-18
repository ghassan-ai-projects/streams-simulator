package domain

import "encoding/json"

type FileDocument struct {
	Path    string `json:"path"`
	Content string `json:"content"`
	Etag    string `json:"etag"`
}

type Definition struct {
	Path      string `json:"path"`
	ID        string `json:"id,omitempty"`
	Title     string `json:"title,omitempty"`
	Instances int    `json:"instances,omitempty"`
}

type instanceSet struct {
	Names []json.RawMessage `json:"names"`
	Count struct {
		Default int `json:"default"`
	} `json:"count"`
}

func (set instanceSet) expanded() int {
	if len(set.Names) > 0 {
		return len(set.Names)
	}
	return set.Count.Default
}

func DefinitionOf(rel string, content []byte) Definition {
	var head struct {
		ID        string        `json:"id"`
		Title     string        `json:"title"`
		Instances []instanceSet `json:"instances"`
		Equipment []instanceSet `json:"equipment"`
	}
	if json.Unmarshal(content, &head) != nil {
		return Definition{Path: rel}
	}
	return Definition{Path: rel, ID: head.ID, Title: head.Title, Instances: expandedCount(head.Instances) + expandedCount(head.Equipment)}
}

func expandedCount(sets []instanceSet) int {
	total := 0
	for _, set := range sets {
		total += set.expanded()
	}
	return total
}

type RunEntry struct {
	WorldID          string  `json:"world_id"`
	Ended            bool    `json:"ended"`
	Complete         bool    `json:"complete"`
	IncompleteReason string  `json:"incomplete_reason"`
	Explore          bool    `json:"explore"`
	Seed             *uint64 `json:"seed"`
	Plant            string  `json:"plant,omitempty"`
	DurationNs       int64   `json:"duration_ns,omitempty"`
	Faults           int     `json:"faults,omitempty"`
	RecordedAt       string  `json:"recorded_at,omitempty"`
	Connected        bool    `json:"connected,omitempty"`
	Running          bool    `json:"running,omitempty"`
}

func (e RunEntry) AsSession(running bool) RunEntry {
	e.Connected, e.Running = true, running
	return e
}

type Listing struct {
	Name         string           `json:"name"`
	Domains      []Definition     `json:"domains"`
	Compositions []Definition     `json:"compositions"`
	Scenarios    []Definition     `json:"scenarios"`
	Layouts      []Definition     `json:"layouts"`
	Kinds        []KindEntry      `json:"kinds"`
	Components   []ComponentEntry `json:"components"`
	Runs         []RunEntry       `json:"runs"`
}

type Record struct {
	Index  int64
	Type   string
	Line   json.RawMessage
	Closes bool
	Seen   bool
}
