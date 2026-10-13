package app

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/ghassan-ai-projects/streams-simulator/internal/cli/internal/files"
	"github.com/ghassan-ai-projects/streams-simulator/internal/component"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
)

const (
	defaultComponentsDir = "components"
	componentsDirsEnv    = "STREAMSIM_COMPONENTS_DIRS"
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
	library, err := component.NewLibraryFromDirs(directories)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return componentVerb(verb, library, directories, fs.Args())
}

func (s *session) componentDirs(flags listFlag) []string {
	dirs := []string(flags)
	if len(dirs) == 0 {
		dirs = []string{defaultComponentsDir}
	}
	if s.getenv == nil {
		return dirs
	}
	return append(dirs, filepath.SplitList(s.getenv(componentsDirsEnv))...)
}

func componentVerb(verb string, library *component.Library, directories, args []string) (any, error) {
	switch {
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

type composeOptions struct {
	composition, out string
	dirs             listFlag
}

func cmdCompose(s *session, args []string) (any, error) {
	options, err := parseComposeOptions(args, s.stderr)
	if err != nil {
		return nil, err
	}
	raw, err := files.Read(options.composition)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	library, err := component.NewLibraryFromDirs(s.componentDirs(options.dirs))
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return composeDomain(library, raw, options)
}

func parseComposeOptions(args []string, stderr io.Writer) (composeOptions, error) {
	var options composeOptions
	fs := newFlagSet("domain compose", stderr)
	fs.StringVar(&options.out, "out", "", "file to write the composed domain to (default: print it inside the output document)")
	registerComponentDirs(fs, &options.dirs)
	if err := parseFlags(fs, args); err != nil {
		return composeOptions{}, err
	}
	if fs.NArg() != 1 {
		return composeOptions{}, usagef("usage: streamsim domain compose <composition.json> [--out <file>] [--components-dir <dir>]")
	}
	options.composition = fs.Arg(0)
	return options, nil
}

func composeDomain(library *component.Library, raw []byte, options composeOptions) (any, error) {
	composed, err := component.Compose(library, raw, options.composition)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	compiled, err := domain.Parse(composed, options.composition)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	result := map[string]any{"id": compiled.Spec.ID, "digest": compiled.Digest, "channels": len(compiled.ChannelNames()),
		"faults": compiled.FaultCount(), "effectors": compiled.EffectorCount()}
	return placeComposed(result, compiled.Spec, composed, options.out)
}

func placeComposed(result map[string]any, spec any, composed []byte, out string) (any, error) {
	if out == "" {
		result["domain"] = spec
		return result, nil
	}
	if err := files.Write(out, composed); err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	result["composed"] = out
	return result, nil
}
