package domain

import (
	"fmt"
	"slices"
)

const (
	etagKey = "etag"
	outKey  = "out"
)

type Guard struct {
	Path string
	Etag string
}

func (a Action) Guarded(args map[string]any) (Guard, bool, error) {
	if !a.Etag {
		return Guard{}, false, nil
	}
	etag, err := a.etagOf(args)
	if err != nil {
		return Guard{}, false, err
	}
	part, err := a.outSpec().render(args[outKey])
	return Guard{Path: part.Path, Etag: etag}, err == nil, err
}

func (a Action) outSpec() CLIArg {
	return a.CLIArgs[slices.IndexFunc(a.CLIArgs, func(spec CLIArg) bool { return spec.Key == outKey })]
}

func (a Action) etagOf(args map[string]any) (string, error) {
	raw, present := args[etagKey]
	if !present {
		return "", nil
	}
	text, ok := raw.(string)
	if !ok {
		return "", fmt.Errorf("%w: %s %q wants the version text of the file", ErrBadArguments, a.Name, etagKey)
	}
	return text, nil
}
