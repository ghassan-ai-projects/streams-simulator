package app

import (
	"fmt"
	"io"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/component"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire"
)

type validateOptions struct {
	stdin bool
	twin  bool
	paths []string
}

func cmdKindValidate(s *session, args []string) (any, error) {
	options, err := parseValidateOptions(s, args)
	if err != nil {
		return nil, err
	}
	if options.stdin {
		return validateKindStream(s, options.rules())
	}
	return validateKindFiles(options.paths, options.rules())
}

func parseValidateOptions(s *session, args []string) (validateOptions, error) {
	var options validateOptions
	fs := newFlagSet("kind validate", s.stderr)
	fs.BoolVar(&options.stdin, "stdin", false, "read one kind document from standard input (at most 1 MiB) instead of files")
	fs.BoolVar(&options.twin, "twin", false, "also report everything a twin of the kind needs that the document lacks")
	if err := parseFlags(fs, args); err != nil {
		return validateOptions{}, err
	}
	options.paths = fs.Args()
	if options.stdin != (len(options.paths) == 0) {
		return validateOptions{}, usagef(kindUsage)
	}
	return options, nil
}

func (o validateOptions) rules() plantwire.TwinRules {
	if o.twin {
		return component.TwinRequirements
	}
	return nil
}

func validateKindFiles(paths []string, rules plantwire.TwinRules) (any, error) {
	validated, err := plantwire.ValidateKindFilesWith(paths, rules)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	document := map[string]any{"valid": true, "kinds": validated.Kinds}
	if len(validated.Boards) > 0 {
		document["boards"] = validated.Boards
	}
	return document, nil
}

func validateKindStream(s *session, rules plantwire.TwinRules) (any, error) {
	kind, err := plantwire.ValidateKindStream(s.input(), rules)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return map[string]any{"valid": true, "kinds": []plantwire.KindSummary{kind}}, nil
}

func (s *session) input() io.Reader {
	if s.stdin == nil {
		return strings.NewReader("")
	}
	return s.stdin
}
