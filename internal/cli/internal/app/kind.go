package app

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire"
)

const kindUsage = "usage: streamsim kind validate [--twin] (<file.kind.json>... | --stdin) | kind twin [<file.kind.json>...] (--choices <file> | --suggest) | kind capture (<file> | --stdin) [--kinds-dir <dir>]"

const populationUsage = "usage: streamsim domain population (<file.domain.json> | --stdin)"

func cmdKind(s *session, args []string) (any, error) {
	switch {
	case len(args) > 0 && args[0] == "validate":
		return cmdKindValidate(s, args[1:])
	case len(args) > 0 && args[0] == "twin":
		return cmdKindTwin(s, args[1:])
	case len(args) > 0 && args[0] == "capture":
		return cmdKindCapture(s, args[1:])
	}
	return nil, usagef(kindUsage)
}

func validateDomainFile(path string, kinds listFlag) (any, error) {
	c, err := domain.Load(path)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	checked, err := checkImplements(c.Spec, path, kinds)
	if err != nil {
		return nil, err
	}
	return domainValidation{Summary: domain.Summarize(c), Valid: true, ImplementsChecked: checked}, nil
}

func cmdPopulation(s *session, args []string) (any, error) {
	fs := newFlagSet("domain population", s.stderr)
	stdin := fs.Bool("stdin", false, "read the domain from standard input")
	if err := parseFlags(fs, args); err != nil {
		return nil, err
	}
	if *stdin != (fs.NArg() == 0) || fs.NArg() > 1 {
		return nil, usagef(populationUsage)
	}
	if *stdin {
		return populationDocument(domain.PopulationOfStream(s.input()))
	}
	return populationDocument(domain.PopulationOfFile(fs.Arg(0)))
}

func populationDocument(population domain.Population, err error) (any, error) {
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return population, nil
}

type domainValidation struct {
	domain.Summary
	Valid             bool                      `json:"valid"`
	ImplementsChecked []model.ImplementsChecked `json:"implements_checked"`
}

func checkImplements(spec *model.DomainSpec, source string, kinds listFlag) ([]model.ImplementsChecked, error) {
	checked, err := plantwire.CheckImplements(spec, kinds.items())
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", model.WithSource(err, source))
	}
	return checked, nil
}

func domainCheck(kinds listFlag) model.DomainCheck {
	return func(spec *model.DomainSpec) ([]model.ImplementsChecked, error) {
		return checkImplements(spec, spec.ID, kinds)
	}
}
