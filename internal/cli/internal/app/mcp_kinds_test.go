package app

import (
	"io"
	"slices"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
)

func TestTheDirectorCheckComparesADomainWithTheKindsDirectoriesOfTheMCPOptions(t *testing.T) {
	t.Parallel()
	spec, err := domain.Load(pumpDomainImplementing(t))
	if err != nil {
		t.Fatal(err)
	}
	kinds := pumpKindsDirectory(t, asWritten)
	options, err := parseMCPOptions([]string{"--role", "director", "--kinds-dir", kinds}, io.Discard)
	if err != nil || !slices.Equal(options.kinds.items(), []string{kinds}) {
		t.Fatalf("parseMCPOptions = %+v, %v, want the kinds directory", options, err)
	}
	if checked, err := domainCheck(options.kinds)(spec.Spec); err != nil || len(checked) != 1 || checked[0].Kind != "pump" {
		t.Fatalf("check with the kind file = %v, %v, want the domain accepted and the pump kind file reported", checked, err)
	}
	withoutKinds, err := parseMCPOptions([]string{"--role", "director"}, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	_, refused := domainCheck(withoutKinds.kinds)(spec.Spec)
	if refused == nil || !strings.Contains(refused.Error(), "no kind file of that name was loaded") {
		t.Fatalf("check without kind files = %v, want the implementing domain refused for lack of its kind file", refused)
	}
}
