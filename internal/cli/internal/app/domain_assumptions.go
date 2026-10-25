package app

import (
	"fmt"
	"io"

	"github.com/ghassan-ai-projects/streams-simulator/internal/cli/internal/files"
	"github.com/ghassan-ai-projects/streams-simulator/internal/component"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire"
)

const assumptionsUsage = "usage: streamsim domain assumptions <domain-file> [--twin <choices.json> --kinds-dir <dir>] [--components-dir <dir>]"

type assumptionsOptions struct {
	domain, twin string
	kinds        listFlag
	libraries    listFlag
}

func cmdAssumptions(s *session, args []string) (any, error) {
	options, err := parseAssumptionsOptions(args, s.stderr)
	if err != nil {
		return nil, err
	}
	raw, err := files.Read(options.domain)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	library, err := component.NewLibrary(s.componentDirs(options.libraries))
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return describeAssumptions(library, component.AssumptionsInput{Domain: raw, Source: options.domain}, options)
}

func describeAssumptions(library *component.Library, input component.AssumptionsInput, options assumptionsOptions) (any, error) {
	if options.twin == "" {
		return assumptionsOf(library, input)
	}
	choices, err := files.Read(options.twin)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return describeTwinAssumptions(library, input, options, choices)
}

func describeTwinAssumptions(library *component.Library, input component.AssumptionsInput, options assumptionsOptions, choices []byte) (any, error) {
	assumptions, err := plantwire.WithKindFiles(nil, options.kinds.items(), func(kinds map[string]map[string]any) (any, error) {
		input.Twin = &component.TwinFiles{Choices: choices, Source: options.twin, Kinds: kinds}
		return assumptionsOf(library, input)
	})
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	return assumptions, nil
}

func assumptionsOf(library *component.Library, input component.AssumptionsInput) (any, error) {
	assumptions, err := component.Assumptions(library, input)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	return assumptions, nil
}

func parseAssumptionsOptions(args []string, stderr io.Writer) (assumptionsOptions, error) {
	var options assumptionsOptions
	fs := newFlagSet("domain assumptions", stderr)
	fs.StringVar(&options.twin, "twin", "", "the twin choices file the domain was scaffolded from")
	registerKindsDirs(fs, &options.kinds)
	registerComponentDirs(fs, &options.libraries)
	if err := parseFlags(fs, args); err != nil {
		return assumptionsOptions{}, err
	}
	if fs.NArg() != 1 {
		return assumptionsOptions{}, usagef(assumptionsUsage)
	}
	options.domain = fs.Arg(0)
	return options, nil
}
