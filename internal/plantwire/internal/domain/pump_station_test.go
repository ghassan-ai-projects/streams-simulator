package domain

import (
	"reflect"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func pumpKindDigest(t *testing.T, spec *domain.Compiled) string {
	t.Helper()
	for i := range spec.Spec.Kinds {
		if spec.Spec.Kinds[i].Name != "pump" {
			continue
		}
		templates, err := WithKindDigests(KindTemplates([]*model.Kind{&spec.Spec.Kinds[i]}))
		if err != nil {
			t.Fatal(err)
		}
		return templates[0].(map[string]any)["kind_digest"].(string)
	}
	t.Fatal("no pump kind")
	return ""
}

func TestTheReplacementProbeBirthsTheSamePumpsWithAnotherKindDigest(t *testing.T) {
	t.Parallel()
	first, err := domain.Load(testsupport.PumpStationDomain())
	if err != nil {
		t.Fatal(err)
	}
	second, err := domain.Load(testsupport.PumpStationProbeDomain())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first.Spec.Instances, second.Spec.Instances) {
		t.Fatalf("instances differ: %+v and %+v", first.Spec.Instances, second.Spec.Instances)
	}
	if a, b := pumpKindDigest(t, first), pumpKindDigest(t, second); a == b {
		t.Fatalf("both worlds birth the pump kind with digest %s", a)
	}
}
