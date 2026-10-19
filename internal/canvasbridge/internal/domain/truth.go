package domain

import "fmt"

type RunState struct {
	Ended   bool
	Explore bool
}

var observerFiles = map[string]bool{"journal.jsonl": true, "trace.jsonl": true, "field-trace.jsonl": true, "interrupts.jsonl": true}

func ReadableRunFile(path WorkspacePath, state RunState) error {
	if path.Class != ClassRun || observerFiles[path.RunFile] || state.Ended || state.Explore {
		return nil
	}
	return fmt.Errorf("%w: %s", ErrSealed, path.Rel)
}
