package domain

import (
	"encoding/json"
	"fmt"
)

type RunState struct {
	Ended   bool
	Explore bool
}

var observerFiles = map[string]bool{"journal.jsonl": true, "trace.jsonl": true}

func ReadableRunFile(path WorkspacePath, state RunState) error {
	if path.Class != ClassRun || observerFiles[path.RunFile] || state.Ended || state.Explore {
		return nil
	}
	return fmt.Errorf("%w: %s", ErrSealed, path.Rel)
}

func ClosedJournal(tail []byte) bool {
	var closing struct {
		Type string `json:"type"`
	}
	return json.Unmarshal(lastLine(tail), &closing) == nil && closing.Type == "run.closed"
}
