package domain

import "encoding/json"

type KindEntry struct {
	Path  string `json:"path"`
	Kind  string `json:"kind,omitempty"`
	Title string `json:"title,omitempty"`
}

type ComponentEntry struct {
	Path  string    `json:"path"`
	ID    string    `json:"id,omitempty"`
	Title string    `json:"title,omitempty"`
	Class FileClass `json:"class"`
}

func KindEntryOf(rel string, content []byte) KindEntry {
	var head struct {
		Kind  string `json:"kind"`
		Title string `json:"title"`
	}
	if json.Unmarshal(content, &head) != nil {
		return KindEntry{Path: rel}
	}
	return KindEntry{Path: rel, Kind: head.Kind, Title: head.Title}
}

func ComponentEntryOf(rel string, class FileClass, content []byte) ComponentEntry {
	var head struct {
		ID    string `json:"id"`
		Title string `json:"title"`
	}
	if json.Unmarshal(content, &head) != nil {
		return ComponentEntry{Path: rel, Class: class}
	}
	return ComponentEntry{Path: rel, ID: head.ID, Title: head.Title, Class: class}
}
