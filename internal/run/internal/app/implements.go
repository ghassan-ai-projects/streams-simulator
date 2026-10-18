package app

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	rules "github.com/ghassan-ai-projects/streams-simulator/internal/run/internal/domain"
)

func CheckDomain(check model.DomainCheck, spec *model.DomainSpec) ([]model.ImplementsChecked, error) {
	if check == nil {
		check = rules.RefuseImplements
	}
	checked, err := check(spec)
	switch {
	case err != nil:
		return nil, fmt.Errorf("%w", err)
	case checked == nil:
		return []model.ImplementsChecked{}, nil
	}
	return checked, nil
}
