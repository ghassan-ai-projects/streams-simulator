package app

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/component"
)

func cmdComponent(s *session, args []string) (any, error) {
	fs := newFlagSet("component", s.stderr)
	var dirs listFlag
	registerComponentDirs(fs, &dirs)
	verb, err := parseCommandVerb(fs, args, "list")
	if err != nil {
		return nil, err
	}
	directories := s.componentDirs(dirs)
	library, err := component.NewLibrary(directories)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return componentVerb(verb, library, directories, fs.Args())
}

func (s *session) componentDirs(flags listFlag) []string {
	environment := ""
	if s.getenv != nil {
		environment = s.getenv(component.DirectoriesEnv)
	}
	return component.CustomerDirectories(flags, environment)
}

func componentVerb(verb string, library *component.Library, directories, args []string) (any, error) {
	switch {
	case verb == "examples" || verb == "example":
		return exampleVerb(verb, args)
	case verb == "show":
		return showComponent(library, args)
	case len(args) > 0 && (verb == "list" || verb == "validate"):
		return nil, usagef("component %s: unexpected argument %q", verb, args[0])
	case verb == "list":
		return map[string]any{"components": library.List()}, nil
	case verb == "validate":
		return validateComponents(library, directories), nil
	}
	return nil, usagef("component: unknown verb %q", verb)
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
