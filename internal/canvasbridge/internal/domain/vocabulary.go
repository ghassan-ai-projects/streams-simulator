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

func DefinitionOf(rel string, content []byte) Definition {
	var head struct {
		ID        string            `json:"id"`
		Title     string            `json:"title"`
		Instances []json.RawMessage `json:"instances"`
	}
	if json.Unmarshal(content, &head) != nil {
		return Definition{Path: rel}
	}
	return Definition{Path: rel, ID: head.ID, Title: head.Title, Instances: len(head.Instances)}
}

type RunEntry struct {
	WorldID    string `json:"world_id"`
	Ended      bool   `json:"ended"`
	Explore    bool   `json:"explore"`
	Plant      string `json:"plant,omitempty"`
	DurationNs int64  `json:"duration_ns,omitempty"`
	Faults     int    `json:"faults,omitempty"`
	RecordedAt string `json:"recorded_at,omitempty"`
}

type Listing struct {
	Name         string       `json:"name"`
	Domains      []Definition `json:"domains"`
	Compositions []Definition `json:"compositions"`
	Scenarios    []Definition `json:"scenarios"`
	Layouts      []Definition `json:"layouts"`
	Runs         []RunEntry   `json:"runs"`
}

type Record struct {
	Index  int64
	Type   string
	Line   json.RawMessage
	Closes bool
	Seen   bool
}
