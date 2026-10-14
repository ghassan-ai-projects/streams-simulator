package domain

import (
	"bytes"
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
	last := tail[bytes.LastIndexByte(bytes.TrimRight(tail, "\n"), '\n')+1:]
	var closing struct {
		Type string `json:"type"`
	}
	return json.Unmarshal(bytes.TrimSpace(last), &closing) == nil && closing.Type == "run.closed"
}
