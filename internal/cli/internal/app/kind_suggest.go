package app

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/component"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire"
)

func suggestChoices(library *component.Library, options twinOptions) (any, error) {
	choices, err := plantwire.WithKindFiles(options.kindFiles, options.kindDirs.items(), func(kinds map[string]map[string]any) (map[string]any, error) {
		suggested, err := component.Suggest(library, kinds)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}
		return suggested, nil
	})
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return map[string]any{"choices": choices}, nil
}
