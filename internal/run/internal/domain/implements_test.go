package domain

import (
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestRefuseImplementsNamesTheFirstKindThatImplementsAKindFile(t *testing.T) {
	t.Parallel()
	spec := &model.DomainSpec{ID: "plant", Kinds: []model.Kind{{Name: "pump"}, {Name: "fan", Implements: "fan_file"}, {Name: "tank", Implements: "tank_file"}}}
	checked, err := RefuseImplements(spec)
	invalid, ok := model.AsValidationError(err)
	if !ok || invalid.Contract != model.ContractDomainSpec || len(invalid.Issues) != 1 || invalid.Issues[0].Pointer != "/kinds/1/implements" ||
		!strings.Contains(invalid.Issues[0].Message, `kind "fan" implements kind file "fan_file"`) || checked != nil {
		t.Fatalf("RefuseImplements = %v, %+v, want a domain refusal at /kinds/1/implements naming kind fan", checked, invalid)
	}
}

func TestRefuseImplementsAcceptsADomainWhoseKindsImplementNothingAndChecksNothing(t *testing.T) {
	t.Parallel()
	spec := &model.DomainSpec{ID: "plant", Kinds: []model.Kind{{Name: "pump"}, {Name: "fan"}}}
	checked, err := RefuseImplements(spec)
	if err != nil || checked == nil || len(checked) != 0 {
		t.Fatalf("RefuseImplements = %v, %v, want an empty list and no error", checked, err)
	}
}
