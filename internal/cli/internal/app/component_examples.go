package app

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/component"
)

func exampleVerb(verb string, args []string) (any, error) {
	if verb == "examples" {
		if len(args) > 0 {
			return nil, usagef("component examples: unexpected argument %q", args[0])
		}
		examples, err := component.Examples()
		return map[string]any{"examples": examples}, wrapCLI(err)
	}
	if len(args) != 1 {
		return nil, usagef("usage: streamsim component example <id>")
	}
	example, err := component.Example(args[0])
	return map[string]any{"example": example}, wrapCLI(err)
}

func wrapCLI(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("streamsim: %w", err)
}
