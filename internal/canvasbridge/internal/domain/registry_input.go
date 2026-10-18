package domain

import "fmt"

const (
	inputKey      = "content"
	MaxInputBytes = 1 << 20
)

func (a Action) Stdin(args map[string]any) ([]byte, error) {
	if a.Input == "" {
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
