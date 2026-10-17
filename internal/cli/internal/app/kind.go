package app

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire"
)

const kindUsage = "usage: streamsim kind validate <file.kind.json>... | kind twin [<file.kind.json>...] --choices <file>"

func cmdKind(s *session, args []string) (any, error) {
	switch {
	case len(args) > 0 && args[0] == "validate":
		return cmdKindValidate(s, args[1:])
	case len(args) > 0 && args[0] == "twin":
		return cmdKindTwin(s, args[1:])
	}
	return nil, usagef(kindUsage)
}

func cmdKindValidate(s *session, args []string) (any, error) {
	fs := newFlagSet("kind validate", s.stderr)
	if err := parseFlags(fs, args); err != nil {
		return nil, err
	}
	if fs.NArg() == 0 {
		return nil, usagef(kindUsage)
	}
	return validateKindFiles(fs.Args())
}

func validateKindFiles(paths []string) (any, error) {
	kinds, err := plantwire.ValidateKindFiles(paths)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return map[string]any{"valid": true, "kinds": kinds}, nil
}

func validateDomainFile(path string, kinds listFlag) (any, error) {
	c, err := domain.Load(path)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	if err := checkImplements(c.Spec, path, kinds); err != nil {
		return nil, err
	}
	return domainValidation{Summary: domain.Summarize(c), Valid: true}, nil
}

type domainValidation struct {
	domain.Summary
	Valid bool `json:"valid"`
}

func checkImplements(spec *model.DomainSpec, source string, kinds listFlag) error {
	if err := plantwire.CheckImplements(spec, kinds.items()); err != nil {
		return fmt.Errorf("streamsim: %w", model.WithSource(err, source))
	}
	return nil
}

func domainCheck(kinds listFlag) func(*model.DomainSpec) error {
	return func(spec *model.DomainSpec) error {
		return checkImplements(spec, spec.ID, kinds)
	}
}

func misplacedFlag(names []string, verb string) (string, string) {
	for _, name := range names {
		if applies := verbOfFlag(name); applies != verb {
			return name, applies
		}
	}
	return "", ""
}

func verbOfFlag(name string) string {
	if name == kindsDirFlag {
		return "validate"
	}
	return "lint"
}
