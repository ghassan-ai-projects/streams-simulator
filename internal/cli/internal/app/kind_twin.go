package app

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/cli/internal/files"
	"github.com/ghassan-ai-projects/streams-simulator/internal/component"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire"
)

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
	if options.suggest {
		return suggestChoices(library, options)
	}
	return scaffoldTwin(library, options)
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
