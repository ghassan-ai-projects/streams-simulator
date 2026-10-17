package domain

import (
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestRefuseImplementsNamesTheFirstKindThatImplementsAKindFile(t *testing.T) {
	t.Parallel()
	spec := &model.DomainSpec{ID: "plant", Kinds: []model.Kind{{Name: "pump"}, {Name: "fan", Implements: "fan_file"}, {Name: "tank", Implements: "tank_file"}}}
	invalid, ok := model.AsValidationError(RefuseImplements(spec))
	if !ok || invalid.Contract != model.ContractDomainSpec || len(invalid.Issues) != 1 || invalid.Issues[0].Pointer != "/kinds/1/implements" ||
		!strings.Contains(invalid.Issues[0].Message, `kind "fan" implements kind file "fan_file"`) {
		t.Fatalf("RefuseImplements = %+v, want a domain refusal at /kinds/1/implements naming kind fan", invalid)
	}
}

func TestRefuseImplementsAcceptsADomainWhoseKindsImplementNothing(t *testing.T) {
	t.Parallel()
	spec := &model.DomainSpec{ID: "plant", Kinds: []model.Kind{{Name: "pump"}, {Name: "fan"}}}
	if err := RefuseImplements(spec); err != nil {
		t.Fatalf("RefuseImplements = %v, want nil", err)
	}
}
