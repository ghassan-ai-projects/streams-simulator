package domain

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func RefuseImplements(spec *model.DomainSpec) error {
	for i := range spec.Kinds {
		if kind := &spec.Kinds[i]; kind.Implements != "" {
			message := fmt.Sprintf("kind %q implements kind file %q, but this director was given no kind files to check it against (start it with a kinds directory)", kind.Name, kind.Implements)
			return &model.ValidationError{Contract: model.ContractDomainSpec, Summary: fmt.Sprintf("domain: %s: %s", spec.ID, message),
				Issues: []model.Issue{model.LocatedIssue(fmt.Sprintf("/kinds/%d/implements", i), message)}}
		}
	}
	return nil
}
