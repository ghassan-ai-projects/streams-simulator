package domain

import (
	"errors"
	"reflect"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestDomainDiffIsAReadOnlyCommandOverTwoDomainFilesOrRunFolders(t *testing.T) {
	t.Parallel()
	diff, err := Lookup("domain.diff")
	if err != nil || diff.Kind != KindCLI || diff.Mutating || diff.Etag || !reflect.DeepEqual(diff.Command, []string{"domain", "diff"}) {
		t.Fatalf("domain.diff = %+v, %v, want a read-only CLI action on domain diff", diff, err)
	}
	got, err := diff.Argv(map[string]any{"a": "w-1", "b": testsupport.WorkspaceDomain("fan"), "format": "markdown"})
	want := []ArgvPart{{Literal: "domain"}, {Literal: "diff"}, {Path: "runs/w-1"}, {Path: testsupport.WorkspaceDomain("fan")}, {Literal: "--format"}, {Literal: "markdown"}}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("diff argv = %+v, %v, want %+v", got, err, want)
	}
}

func TestDomainDiffRefusesAMissingSideAFileOfAnotherClassARunThatClimbsAndAnUnknownKey(t *testing.T) {
	t.Parallel()
	diff, _ := Lookup("domain.diff")
	for name, args := range map[string]map[string]any{
		"no first side":        {"b": testsupport.WorkspaceDomain("fan")},
		"no second side":       {"a": testsupport.WorkspaceDomain("fan")},
		"a composition":        {"a": "compositions/fan.composition.json", "b": testsupport.WorkspaceDomain("fan")},
		"a domain that climbs": {"a": testsupport.WorkspaceFile("domains", "../x.domain.json"), "b": testsupport.WorkspaceDomain("fan")},
		"a run that climbs":    {"a": "..", "b": testsupport.WorkspaceDomain("fan")},
		"a run folder path":    {"a": "runs/w-1", "b": testsupport.WorkspaceDomain("fan")},
		"a flag as a side":     {"a": "--out", "b": testsupport.WorkspaceDomain("fan")},
		"a number":             {"a": 3, "b": testsupport.WorkspaceDomain("fan")},
		"an unknown key":       {"a": "w-1", "b": testsupport.WorkspaceDomain("fan"), "out": "x"},
	} {
		if _, err := diff.Argv(args); !errors.Is(err, ErrBadArguments) {
			t.Errorf("%s: %v, want ErrBadArguments", name, err)
		}
	}
}
