package domain

import (
	"errors"
	"strings"
	"testing"
)

func TestTheWorkspacePolicyDefaultIsOneNamedFileInSettings(t *testing.T) {
	t.Parallel()
	got, err := ParsePath("settings/default.policies.json")
	want := WorkspacePath{Rel: "settings/default.policies.json", Class: ClassPolicies}
	if err != nil || got != want {
		t.Fatalf("ParsePath = %+v, %v; want %+v", got, err, want)
	}
	if err := got.Writable(); err != nil {
		t.Fatalf("the default is editable: %v", err)
	}
}

func TestNothingElseInSettingsAndNoPolicyFileElsewhereIsAccepted(t *testing.T) {
	t.Parallel()
	rels := []string{"settings/other.policies.json", "settings/default.json", "settings/default.policies.json/x", "settings/sub/default.policies.json",
		"settings/.policies.json", "settings", "settings/", "Settings/default.policies.json", "default.policies.json", "kinds/default.policies.json",
		"settings/../runs/w-1/run.json", "../settings/default.policies.json", "settings//default.policies.json", `settings\default.policies.json`}
	for folder := range definitionFolders {
		rels = append(rels, folder+"/default.policies.json")
	}
	for _, rel := range rels {
		if _, err := ParsePath(rel); !errors.Is(err, ErrBadPath) {
			t.Errorf("ParsePath(%q) = %v, want ErrBadPath", rel, err)
		}
	}
}

func TestThePolicyDefaultIsNotAnArgumentOfAnyAction(t *testing.T) {
	t.Parallel()
	for _, action := range Actions() {
		for _, arg := range action.CLIArgs {
			if arg.Class == ClassPolicies {
				t.Errorf("action %s takes a %s file; settings never reach a command", action.Name, ClassPolicies)
			}
		}
	}
	spec := CLIArg{Key: "path", Shape: ShapePath, Class: ClassDomain, Required: true}
	if _, err := spec.renderWorkspaceFile("settings/default.policies.json"); err == nil || !strings.Contains(err.Error(), "want a domain file of the workspace") {
		t.Fatalf("the default must not pass as a domain file argument: %v", err)
	}
	if _, err := StemPath(ClassPolicies, "default"); !errors.Is(err, ErrBadPath) {
		t.Fatalf("no action names the default by stem: %v", err)
	}
}
