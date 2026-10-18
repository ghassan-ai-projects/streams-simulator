package app

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire"
)

const kindUsage = "usage: streamsim kind validate [--twin] (<file.kind.json>... | --stdin) | kind twin [<file.kind.json>...] (--choices <file> | --suggest)"

func cmdKind(s *session, args []string) (any, error) {
	switch {
	case len(args) > 0 && args[0] == "validate":
		return cmdKindValidate(s, args[1:])
	case len(args) > 0 && args[0] == "twin":
		return cmdKindTwin(s, args[1:])
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
