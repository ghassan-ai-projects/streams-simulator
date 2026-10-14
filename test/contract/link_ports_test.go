package contract

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestTheComposedFactoryExampleNamesThePortsOfEveryLink(t *testing.T) {
	t.Parallel()
	links := factoryUtilities(t).Spec.Links
	if len(links) == 0 {
		t.Fatal("the factory example holds no links")
	}
	for i, link := range links {
		if link.FromPort == "" || link.ToPort == "" {
			t.Errorf("link %d = %+v, want both ports", i, link)
		}
	}
}

func TestADomainLinkRefusesAMemberTheContractDoesNotName(t *testing.T) {
	t.Parallel()
	raw := factoryUtilities(t).Raw
	changed := bytes.Replace(raw, []byte(`"from_port"`), []byte(`"via_port"`), 1)
	if bytes.Equal(changed, raw) {
		t.Fatal("the factory example names no from_port to replace")
	}
	_, err := domain.Parse(changed, "factory.domain.json")
	var invalid *model.ValidationError
	if !errors.As(err, &invalid) || !strings.Contains(err.Error(), "via_port") {
		t.Fatalf("Parse = %v, want a validation error naming via_port", err)
	}
}
