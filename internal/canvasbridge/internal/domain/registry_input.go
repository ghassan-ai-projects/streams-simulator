package domain

import (
	"fmt"
	"strings"
)

const (
	inputKey      = "content"
	runKey        = "run"
	MaxInputBytes = 1 << 20
)

func (a Action) Stdin(args map[string]any) ([]byte, error) {
	if _, given := args[a.Input]; a.Input == "" || (len(a.OneOf) > 0 && !given) {
		return nil, nil
	}
	text, ok := args[a.Input].(string)
	if !ok {
		return nil, fmt.Errorf("%w: %s needs %q as text", ErrBadArguments, a.Name, a.Input)
	}
	if len(text) > MaxInputBytes {
		return nil, fmt.Errorf("%w: %s %q holds more than %d bytes", ErrTooLarge, a.Name, a.Input, MaxInputBytes)
	}
	return []byte(text), nil
}

func (a Action) admit(args map[string]any) error {
	if err := a.closedArguments(args); err != nil {
		return err
	}
	return a.exactlyOne(args)
}

func (a Action) exactlyOne(args map[string]any) error {
	given := 0
	for _, key := range a.OneOf {
		if _, present := args[key]; present {
			given++
		}
	}
	if len(a.OneOf) > 0 && given != 1 {
		return fmt.Errorf("%w: %s takes exactly one of %s", ErrBadArguments, a.Name, strings.Join(a.OneOf, " or "))
	}
	return nil
}

func (a Action) stdinFlag(args map[string]any) []ArgvPart {
	if _, given := args[a.Input]; a.StdinFlag == "" || !given {
		return nil
	}
	return literals([]string{a.StdinFlag})
}
