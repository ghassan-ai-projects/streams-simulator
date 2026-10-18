package domain

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func RefuseImplements(spec *model.DomainSpec) ([]model.ImplementsChecked, error) {
	for i := range spec.Kinds {
		if kind := &spec.Kinds[i]; kind.Implements != "" {
			message := fmt.Sprintf("kind %q implements kind file %q, but no check against kind files was given (name a kinds directory)", kind.Name, kind.Implements)
			return nil, &model.ValidationError{Contract: model.ContractDomainSpec, Summary: fmt.Sprintf("domain: %s: %s", spec.ID, message),
				Issues: []model.Issue{model.LocatedIssue(fmt.Sprintf("/kinds/%d/implements", i), message)}}
		}
	}
	return []model.ImplementsChecked{}, nil
}
