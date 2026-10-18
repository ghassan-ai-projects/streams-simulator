package app

import (
	"flag"
	"fmt"
	"io"

	"github.com/ghassan-ai-projects/streams-simulator/internal/cli/internal/files"
	"github.com/ghassan-ai-projects/streams-simulator/internal/component"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire"
)

const twinUsage = "usage: streamsim kind twin [<file.kind.json>...] --choices <choices.json> [--kinds-dir <dir>] [--out <file>] [--components-dir <dir>]"

type twinOptions struct {
	choices   string
	out       string
	kindFiles []string
	kindDirs  listFlag
	libraries listFlag
}

type twinDomain struct {
	composedDomain
	Implements []string `json:"implements"`
}

func cmdKindTwin(s *session, args []string) (any, error) {
	options, err := parseTwinOptions(args, s.stderr)
	if err != nil {
		return nil, err
	}
	library, err := component.NewLibrary(s.componentDirs(options.libraries))
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return scaffoldTwin(library, options)
}

func parseTwinOptions(args []string, stderr io.Writer) (twinOptions, error) {
	var options twinOptions
	fs := newFlagSet("kind twin", stderr)
	registerTwinFlags(fs, &options)
	if err := parseFlags(fs, args); err != nil {
		return twinOptions{}, err
	}
	if options.choices == "" {
		return twinOptions{}, usagef(twinUsage)
	}
	options.kindFiles = fs.Args()
	return options, nil
}

func registerTwinFlags(fs *flag.FlagSet, options *twinOptions) {
	fs.StringVar(&options.choices, "choices", "", "the choices document: the instrument of each metric and the response of each command")
	fs.StringVar(&options.out, "out", "", "file to write the twin domain to (default: print it inside the output document)")
	registerKindsDirs(fs, &options.kindDirs)
	registerComponentDirs(fs, &options.libraries)
}

func scaffoldTwin(library *component.Library, options twinOptions) (any, error) {
	twin, err := plantwire.Scaffold(options.kindFiles, options.kindDirs.items(), options.choices, twinOf(library, options.choices))
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return placeTwin(twin, options)
}

func twinOf(library *component.Library, choicesPath string) func(map[string]map[string]any) (*component.Composed, *model.DomainSpec, error) {
	return func(kinds map[string]map[string]any) (*component.Composed, *model.DomainSpec, error) {
		choices, err := files.Read(choicesPath)
		if err != nil {
			return nil, nil, fmt.Errorf("%w", err)
		}
		twin, err := component.Twin(library, kinds, choices, choicesPath)
		if err != nil {
			return nil, nil, fmt.Errorf("%w", err)
		}
		return twin, twin.Domain.Spec, nil
	}
}

func placeTwin(twin *component.Composed, options twinOptions) (any, error) {
	result := twinDomain{composedDomain: composedDomain{DomainSummary: twin.Summary}, Implements: twin.Implements}
	if err := placeComposed(&result.composedDomain, twin, options.out); err != nil {
		return nil, err
	}
	return result, nil
}
