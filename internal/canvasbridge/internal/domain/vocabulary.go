package domain

import "encoding/json"

type FileDocument struct {
	Path    string `json:"path"`
	Content string `json:"content"`
	Etag    string `json:"etag"`
}

type Definition struct {
	Path  string `json:"path"`
	ID    string `json:"id,omitempty"`
	Title string `json:"title,omitempty"`
}

type RunEntry struct {
	WorldID string `json:"world_id"`
	Ended   bool   `json:"ended"`
	Explore bool   `json:"explore"`
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
