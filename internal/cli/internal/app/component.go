package app

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/component"
)

func cmdComponent(s *session, args []string) (any, error) {
	options, err := parseComponentOptions(s, args)
	if err != nil {
		return nil, err
	}
	directories := s.componentDirs(options.dirs)
	library, err := component.NewLibrary(directories)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return componentVerb(options, library, directories)
}

func (s *session) componentDirs(flags listFlag) []string {
	environment := ""
	if s.getenv != nil {
		environment = s.getenv(component.DirectoriesEnv)
	}
	return component.CustomerDirectories(flags, environment)
}

func componentVerb(options componentOptions, library *component.Library, directories []string) (any, error) {
	switch {
	case options.verb == "examples" || options.verb == "example":
		return exampleVerb(options.verb, options.args)
	case options.verb == "show":
		return showComponent(library, options.args)
	case len(options.args) > 0 && (options.verb == "list" || options.verb == "validate"):
		return nil, usagef("component %s: unexpected argument %q", options.verb, options.args[0])
	case options.verb == "list":
		return map[string]any{"components": library.ListFor(options.unit)}, nil
	case options.verb == "validate":
		return validateComponents(library, directories), nil
	}
	return nil, usagef("component: unknown verb %q", options.verb)
}

func showComponent(library *component.Library, args []string) (any, error) {
	if len(args) != 1 {
		return nil, usagef("usage: streamsim component show <namespace/name[@major]>")
	}
	detail, err := library.Describe(args[0])
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return detail, nil
}

func validateComponents(library *component.Library, directories []string) any {
	equipment, instruments := library.Counts()
	return map[string]any{"valid": true, "directories": directories, "equipment": equipment, "instruments": instruments}
}
